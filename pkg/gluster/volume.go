package gluster

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

var _ model.GlusterVolumeManager = &Manager{}

type Manager struct {
	capability.Logger
	config Configuration
}

// NewManager creates a new Gluster manager using the gluster CLI
func NewManager(logger capability.Logger, config Configuration) *Manager {
	return &Manager{
		Logger: logger,
		config: config,
	}
}

// CreateVolume creates a new Gluster volume with the given name and brick path
func (m *Manager) CreateVolume(volumeName string, brickPath string) error {
	fields := map[string]interface{}{
		"volume":     volumeName,
		"brick_path": brickPath,
	}

	// First, ensure the brick directory exists
	if err := os.MkdirAll(brickPath, 0o755); err != nil {
		fields["error"] = err
		m.Error("Failed to create brick directory", fields)
		return fmt.Errorf("failed to create brick directory: %w", err)
	}

	// Check if volume already exists
	if err := m.volumeExists(volumeName); err == nil {
		m.Info("Gluster volume already exists", fields)
		return nil
	}

	brick := fmt.Sprintf("%s:%s", m.config.Host, brickPath)
	cmd := exec.Command(
		"gluster",
		"volume",
		"create",
		volumeName,
		brick,
		"force",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		fields["error"] = err
		fields["output"] = strings.TrimSpace(string(output))
		m.Error("Failed to create Gluster volume", fields)
		return fmt.Errorf("failed to create gluster volume: %w", err)
	}

	fields["brick"] = brick
	m.Info("Gluster volume created successfully", fields)
	return nil
}

// StartVolume starts a Gluster volume
func (m *Manager) StartVolume(volumeName string) error {
	fields := map[string]interface{}{
		"volume": volumeName,
	}

	if running, err := m.volumeStarted(volumeName); err == nil && running {
		m.Info("Gluster volume already started", fields)
		return nil
	}

	cmd := exec.Command("gluster", "volume", "start", volumeName, "force")
	if output, err := cmd.CombinedOutput(); err != nil {
		fields["error"] = err
		fields["output"] = strings.TrimSpace(string(output))
		m.Error("Failed to start Gluster volume", fields)
		return fmt.Errorf("failed to start gluster volume: %w", err)
	}

	m.Info("Gluster volume started successfully", fields)
	return nil
}

func (m *Manager) volumeExists(volumeName string) error {
	cmd := exec.Command("gluster", "volume", "info", volumeName)
	if _, err := cmd.CombinedOutput(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) volumeStarted(volumeName string) (bool, error) {
	cmd := exec.Command("gluster", "volume", "status", volumeName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}
