package disk

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/gpt"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

var (
	ErrUnsupportedPartitionTable = errors.New("unsupported partition table")
)

func convertPartitionTable(table *gpt.Table) model.Partitions {
	var partitions model.Partitions
	for _, p := range table.Partitions {
		partitions = append(partitions, model.Partition{
			Name: p.Name,
			Uuid: p.UUID(),
			Size: p.GetSize(),
		})
	}
	return partitions
}

// createPartitionTable creates a GPT partition table and a single partition
func createPartitionTable(device string, layout model.DiskLayout) (model.Partitions, error) {
	disk, err := diskfs.Open(device)
	if err != nil {
		return nil, fmt.Errorf("open disk: %w", err)
	}
	defer disk.Close()

	var partitions []*gpt.Partition
	for _, part := range layout.Partitions {
		partitions = append(partitions, &gpt.Partition{
			Name:  part.Label,
			Start: 2048,
			End:   uint64(disk.Size/disk.LogicalBlocksize) - 1,
			Type:  gpt.LinuxFilesystem,
		})
	}
	// Create GPT partition table if not present
	err = disk.Partition(&gpt.Table{
		Partitions:    partitions,
		ProtectiveMBR: true,
	})
	if err != nil {
		return nil, fmt.Errorf("partition: %w", err)
	}

	generic, err := disk.GetPartitionTable()
	if err != nil {
		return nil, fmt.Errorf("get partition table: %w", err)
	}

	table, _ := generic.(*gpt.Table)
	return convertPartitionTable(table), nil
}

func getPartitionTable(dev string) (model.Partitions, error) {
	disk, err := diskfs.Open(dev, diskfs.WithOpenMode(diskfs.ReadOnly))
	if err != nil {
		return nil, err
	}
	defer disk.Close()
	generic, err := disk.GetPartitionTable()
	if err != nil {
		return nil, err
	}

	table, ok := generic.(*gpt.Table)
	if !ok {
		return nil, ErrUnsupportedPartitionTable
	}

	return convertPartitionTable(table), nil
}

func (m *Manager) Partition(ctx context.Context, disk model.Disk) (partitions model.Partitions, err error) {
	fields := map[string]any{
		"disk":       disk,
		"partitions": len(disk.Layout.Partitions),
	}

	changes := []string{}
	block, err := ListBlocks(disk.Path)
	if err != nil {
		return nil, err
	}
	if len(disk.Layout.Partitions) == len(block.Children) {
		partitions, err = getPartitionTable(disk.Path)
		if err != nil {
			return nil, err
		}

		for i, tpart := range partitions {
			lpart := disk.Layout.Partitions[i]
			bpart := block.Children[i]
			// FIXME: lsblk may return empty labels
			if tpart.Name != lpart.Label {
				changes = append(changes, fmt.Sprintf("partition %d: name mismatch (%q != %q)", i, bpart.Label, lpart.Label))
				break
			}

			// FIXME: check size with some tolerance
		}
	} else {
		changes = append(changes, fmt.Sprintf("partition count mismatch (%d != %d)", len(block.Children), len(disk.Layout.Partitions)))
	}

	if len(changes) == 0 {
		fields["partitions"] = partitions
		m.Info("Disk already partitioned", fields)
		return partitions, nil
	}

	fields["changes"] = changes

	m.Info("Partitioning disk", fields)
	partitions, err = createPartitionTable(disk.Path, disk.Layout)
	if err != nil {
		fields["error"] = err
		m.Error("Failed to partition disk", fields)
		return nil, err
	}

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second) // wait for the OS to recognize new partitions
		partitions2, err := getPartitionTable(disk.Path)
		if err != nil {
			fields["error"] = err
			m.Error("Failed to get partition table after partitioning", fields)
			return nil, err
		}
		if len(partitions2) == len(partitions) {
			break
		}
		m.Warn("Partition table not yet recognized by OS, retrying...", fields)
	}

	return partitions, nil
}
