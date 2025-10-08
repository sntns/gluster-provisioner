package disk

import (
	"fmt"
	"os"

	"github.com/moby/sys/mount"
)

// CreateMountPoint creates the mount directory if needed
func CreateMountPoint(name string) (string, error) {
	mountPoint := "/media/" + name
	if err := os.MkdirAll(mountPoint, 0o755); err != nil {
		return "", fmt.Errorf("mkdir failed: %w", err)
	}
	return mountPoint, nil
}

// MountPartition mounts the given partition to the mount point
func MountPartition(part, name string) error {
	mountPoint, err := CreateMountPoint(name)
	if err != nil {
		return err
	}

	if err := mount.Mount(part, mountPoint, "ext4", ""); err != nil {
		return fmt.Errorf("mount failed: %w", err)
	}
	return nil
}
