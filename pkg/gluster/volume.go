package gluster

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/moby/sys/mountinfo"
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

var _ model.GlusterVolumeManager = &Manager{}

type Manager struct {
	capability.Logger
}

// NewManager creates a new Gluster manager using the gluster CLI
func NewManager(logger capability.Logger) *Manager {
	return &Manager{
		Logger: logger,
	}
}

// CreateVolume creates a new Gluster volume with the given name and brick path
func (m *Manager) CreateVolume(volumeName string, brickPath string) error {
	fields := map[string]any{
		"volume":     volumeName,
		"brick_path": brickPath,
	}

	// First, ensure the brick directory exists
	if err := os.MkdirAll(brickPath, 0o755); err != nil {
		fields["error"] = err
		m.Error("Failed to create brick directory", fields)
		return err
	}

	// Check if volume already exists
	if exists, err := m.VolumeExists(volumeName); err != nil {
		fields["error"] = err
		m.Error("Failed to check if Gluster volume exists", fields)
		return err
	} else if exists {
		m.Info("Gluster volume already exists, skipping creation", fields)
		return nil
	}

	rawPeers := os.Getenv("GLUSTER_PEERS")
	if rawPeers == "" {
		err := fmt.Errorf("GLUSTER_PEERS environment variable is not set")
		fields["error"] = err
		m.Error("Failed to get Gluster peers", fields)
		return err
	}
	peers := strings.Split(rawPeers, ",")

	bricks := []string{}
	for _, peer := range peers {
		bricks = append(bricks, fmt.Sprintf("%s:%s", peer, brickPath))
	}

	args := []string{
		"volume",
		"create",
		volumeName,
		"replica",
		fmt.Sprintf("%d", len(bricks)),
	}
	args = append(args, bricks...)
	args = append(args, "force")

	cmd := exec.Command("gluster", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		fields["error"] = err
		fields["output"] = strings.TrimSpace(string(output))
		m.Error("Failed to create Gluster volume", fields)
		return err
	}

	fields["bricks"] = bricks
	m.Info("Gluster volume created successfully", fields)
	return nil
}

// StartVolume starts a Gluster volume
func (m *Manager) StartVolume(volumeName string) error {
	fields := map[string]any{
		"volume": volumeName,
	}

	if running, err := m.VolumeStarted(volumeName); err != nil {
		fields["error"] = err
		m.Error("Failed to check if Gluster volume is started", fields)
		return err
	} else if running {
		m.Info("Gluster volume already started", fields)
		return nil
	}

	cmd := exec.Command("gluster", "volume", "start", volumeName, "force")
	if output, err := cmd.CombinedOutput(); err != nil {
		fields["error"] = err
		fields["output"] = strings.TrimSpace(string(output))
		m.Error("Failed to start Gluster volume", fields)
		return err
	}

	m.Info("Gluster volume started successfully", fields)
	return nil
}

// MountVolume mounts a Gluster volume via the system mount helper.
func (m *Manager) MountVolume(volumeName string, mountPoint string) error {
	fields := map[string]any{
		"volume":      volumeName,
		"mount_point": mountPoint,
	}

	// Basic preflight: FUSE device must exist.
	if _, err := os.Stat("/dev/fuse"); err != nil {
		fields["error"] = err
		m.Error("FUSE device not available (/dev/fuse)", fields)
		return err
	}

	if err := os.MkdirAll(mountPoint, 0o755); err != nil {
		fields["error"] = err
		m.Error("Failed to create Gluster mount point", fields)
		return err
	}

	if mounted, err := m.mountPointMounted(mountPoint); err != nil {
		fields["error"] = err
		m.Error("Failed to check if mount point is mounted", fields)
		return err
	} else if mounted {
		m.Info("Gluster volume already mounted", fields)
		return nil
	}

	source := fmt.Sprintf("127.0.0.1:/%s", volumeName)
	fields["source"] = source

	cmd := exec.Command("mount", "-t", "glusterfs", source, mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		fields["error"] = err
		fields["output"] = strings.TrimSpace(string(output))
		m.Error("Failed to mount Gluster volume", fields)
		return err
	}

	fields["mount_point_base"] = filepath.Dir(mountPoint)
	m.Info("Gluster volume mounted successfully", fields)
	return nil
}

func (m *Manager) VolumeExists(volumeName string) (bool, error) {
	cmd := exec.Command("gluster", "volume", "info", volumeName)
	if _, err := cmd.CombinedOutput(); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *Manager) VolumeStarted(volumeName string) (bool, error) {
	cmd := exec.Command("gluster", "volume", "status", volumeName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

func (m *Manager) mountPointMounted(mountPoint string) (bool, error) {
	info, err := mountinfo.GetMounts(nil)
	if err != nil {
		return false, err
	}
	for _, mi := range info {
		if mi.Mountpoint == mountPoint {
			return true, nil
		}
	}
	return false, nil
}
