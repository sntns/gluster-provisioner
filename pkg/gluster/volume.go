package gluster

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

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
	if exists := m.VolumeExists(volumeName); exists {
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

	if running := m.VolumeStarted(volumeName); running {
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

func (m *Manager) EnsureMounted(volumeName, mountPoint string) error {
	fields := map[string]any{
		"volume": volumeName,
		"mount":  mountPoint,
	}

	// 1. If something is mounted here, kill it FIRST (no filesystem syscalls yet)
	if isMounted(mountPoint) {
		m.Warn("Existing mount detected, forcing unmount", fields)
		forceUnmount(mountPoint)
	}

	// 2. Remove mountpoint regardless of state (ignores ENOTCONN)
	_ = os.RemoveAll(mountPoint)

	// 3. Recreate clean directory
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		fields["error"] = err
		return fmt.Errorf("mkdir mountpoint failed: %w", err)
	}

	source := fmt.Sprintf("127.0.0.1:/%s", volumeName)

	deadline := time.Now().Add(2 * time.Minute)

	for {
		cmd := exec.Command(
			"mount",
			"-t", "glusterfs",
			"-o", "backupvolfile-server=127.0.0.1",
			source,
			mountPoint,
		)

		out, err := cmd.CombinedOutput()

		if err == nil {
			m.Info("Mounted successfully", fields)
			return nil
		}

		m.Error("Mount failed, retrying", map[string]any{
			"error":  err,
			"output": strings.TrimSpace(string(out)),
		})

		if time.Now().After(deadline) {
			return fmt.Errorf("mount timeout: %s", out)
		}

		time.Sleep(2 * time.Second)
	}
}

func (m *Manager) VolumeExists(volumeName string) bool {
	cmd := exec.Command("gluster", "volume", "info", volumeName)
	if _, err := cmd.CombinedOutput(); err != nil {
		m.Info("Gluster volume does not exist", map[string]any{
			"volume": volumeName,
			"error":  err,
		})
		return false
	}
	return true
}

func (m *Manager) VolumeStarted(volumeName string) bool {
	cmd := exec.Command("gluster", "volume", "status", volumeName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		m.Info("Gluster volume is not started", map[string]any{
			"volume": volumeName,
			"error":  err,
			"output": strings.TrimSpace(string(output)),
		})
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

func isMounted(target string) bool {
	cmd := exec.Command("findmnt", "-n", "-T", target)
	out, err := cmd.Output()
	return err == nil && len(out) > 0
}

func forceUnmount(mp string) {
	exec.Command("umount", "-lf", mp).Run()
	exec.Command("fusermount", "-uz", mp).Run()
	time.Sleep(1 * time.Second)
}
