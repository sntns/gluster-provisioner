package disk

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/sntns/gluster-provisioner/pkg/model"
)

var (
	ErrInvalidPartition      = errors.New("invalid partition scheme")
	ErrUnsupportedFilesystem = errors.New("unsupported filesystem")
)

func formatPartitions(ctx context.Context, device string, layout model.DiskLayout, block *BlockDevice) (model.Filesystems, error) {
	var filesystems model.Filesystems
	for i, partLayout := range layout.Partitions {
		lpart := layout.Partitions[i]
		path := fmt.Sprintf("/dev/%s", block.Children[i].Name)

		// After partitioning, the kernel/udev may take a moment to create the partition node.
		// Wait a bit to avoid flaky mkfs failures (e.g., "/dev/vdb1: No such file or directory").
		deadline := time.Now().Add(10 * time.Second)
		for {
			if _, err := os.Stat(path); err == nil {
				break
			}
			if time.Now().After(deadline) {
				break
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(500 * time.Millisecond):
			}
		}

		switch partLayout.Filesystem {
		case "ext4":
			cmd := exec.CommandContext(ctx, "mkfs.ext4", "-L", lpart.Label, path)
			out, err := cmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("mkfs.ext4 failed for %s: %w: %s", path, err, string(out))
			}
			filesystems = append(filesystems, model.Filesystem{
				Label: partLayout.Label,
				Path:  path,
			})
		default:
			return nil, ErrUnsupportedFilesystem
		}
	}
	return filesystems, nil
}

func (m *Manager) Format(ctx context.Context, disk model.Disk, partitions model.Partitions) (model.Filesystems, error) {
	fields := map[string]any{
		"disk":       disk,
		"partitions": len(disk.Layout.Partitions),
	}

	block, err := ListFilesystems(disk.Path)
	if err != nil {
		return nil, err
	}

	changes := []string{}

	// Not the same partition number
	if len(block.Children) != len(partitions) {
		return nil, ErrInvalidPartition
	}

	var filesystems model.Filesystems
	for i, bpart := range block.Children {
		lpart := disk.Layout.Partitions[i]
		part := partitions[i]
		if part.Name != bpart.Label {
			changes = append(changes, fmt.Sprintf("partition %d: name mismatch (%q != %q)", i, bpart.Label, part.Name))
		}
		if bpart.Fstype != lpart.Filesystem {
			changes = append(changes, fmt.Sprintf("partition %d: filesystem mismatch (%q != %q)", i, bpart.Fstype, lpart.Filesystem))
		}

		filesystems = append(filesystems, model.Filesystem{
			Label: bpart.Label,
			Path:  fmt.Sprintf("/dev/%s", bpart.Name),
		})
	}

	if len(changes) == 0 {
		fields["filesystems"] = filesystems
		m.Info("Disk already formatted as desired", fields)
		return filesystems, nil
	}

	fields["changes"] = changes

	m.Info("Formatting disk", fields)
	filesystems, err = formatPartitions(ctx, disk.Path, disk.Layout, block)
	if err != nil {
		fields["error"] = err
		m.Error("Failed to format disk", fields)
		return nil, err
	}

	return filesystems, nil
}
