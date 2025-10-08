package disk

import (
	"fmt"
	"os/exec"
)

// FormatExt4 formats the given partition as ext4 using mkfs.ext4
func FormatExt4(part string) error {
	cmd := exec.Command("mkfs.ext4", "-F", part)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mkfs.ext4 failed: %w", err)
	}
	return nil
}
