package disk

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

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
		switch partLayout.Filesystem {
		case "ext4":
			cmd := exec.Command("mkfs.ext4", "-L", lpart.Label, path)
			if err := cmd.Run(); err != nil {
				return nil, err
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
