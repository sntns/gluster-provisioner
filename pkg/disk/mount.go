package disk

import (
	"context"
	"fmt"
	"os"

	"github.com/moby/sys/mount"
	"github.com/moby/sys/mountinfo"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

// ensureMountPoint creates the mount directory if needed
func ensureMountPoint(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Mount(ctx context.Context, disk model.Disk, filesystems model.Filesystems) (model.Mountpoints, error) {
	fields := map[string]interface{}{
		"disk":        disk,
		"filesystems": len(filesystems),
	}

	info, err := mountinfo.GetMounts(nil)
	if err != nil {
		return nil, err
	}

	mounts := make([]*mountinfo.Info, len(filesystems))
mountinfos:
	for _, mi := range info {
		for i, fs := range filesystems {
			if mi.Source == fs.Path {
				mounts[i] = mi
				continue mountinfos
			}
		}
	}

	changes := []string{}
	var mountpoints model.Mountpoints
	for i, fs := range filesystems {
		info := mounts[i]
		if info != nil {
			mountpoints = append(mountpoints, model.Mountpoint{
				Label: fs.Label,
				Path:  info.Mountpoint,
			})
			continue
		}
		changes = append(changes, fmt.Sprintf("filesystem %q is not mounted", fs.Label))
	}

	if len(changes) == 0 {
		fields["mountpoints"] = mountpoints
		m.Info("Disk already mounted", fields)
		return mountpoints, nil
	}

	fields["changes"] = changes

	m.Info("Mounting filesystems", fields)
	for i, fs := range filesystems {
		lpart := disk.Layout.Partitions[i]
		info := mounts[i]
		if info != nil {
			continue
		}
		if err := ensureMountPoint(lpart.MountPoint); err != nil {
			fields["error"] = err
			m.Error("Failed to ensure mount point", fields)
			return nil, err
		}

		opts := ""
		for k, v := range lpart.Options {
			if v != nil {
				opts += fmt.Sprintf("%s=%s,", k, *v)
			} else {
				opts += fmt.Sprintf("%s,", k)
			}
		}
		if len(opts) > 0 {
			opts = opts[:len(opts)-1] // Remove trailing comma
		}
		if err := mount.Mount(fs.Path, lpart.MountPoint, lpart.Filesystem, opts); err != nil {
			fields["error"] = err
			m.Error("Failed to mount filesystem", fields)
			return nil, err
		}
		mountpoints = append(mountpoints, model.Mountpoint{
			Label: fs.Label,
			Path:  lpart.MountPoint,
		})
	}

	fields["mountpoints"] = mountpoints
	m.Info("Disk mounted", fields)

	return mountpoints, nil
}

func (m *Manager) Unmount(ctx context.Context, disk model.Disk, filesystems model.Filesystems) error {
	fields := map[string]interface{}{
		"disk":        disk,
		"filesystems": len(filesystems),
	}
	m.Info("Unmounting disk", fields)

	for i, fs := range filesystems {
		lpart := disk.Layout.Partitions[i]
		fields["filesystem"] = fs.Label
		fields["path"] = lpart.MountPoint
		m.Debug("Unmounting partition", fields)
		if err := mount.Unmount(lpart.MountPoint); err != nil {
			fields["error"] = err
			m.Error("Failed to unmount filesystem", fields)
			return err
		}
	}
	return nil
}
