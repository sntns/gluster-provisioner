package gluster

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

var _ model.GlusterVolumeManager = &VolumeManager{}

type VolumeManager struct {
	capability.Logger
}

func NewVolumeManager(logger capability.Logger) *VolumeManager {
	return &VolumeManager{
		Logger: logger,
	}
}

// CreateVolume creates a new Gluster volume with the given name and brick path
func (m *VolumeManager) CreateVolume(volumeName string, brickPath string) error {
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
	checkCmd := exec.Command("gluster", "volume", "info", volumeName)
	if err := checkCmd.Run(); err == nil {
		// Volume already exists
		m.Info("Gluster volume already exists", fields)
		return nil
	}

	// Get the hostname for the brick specification
	hostname, err := os.Hostname()
	if err != nil {
		fields["error"] = err
		m.Error("Failed to get hostname", fields)
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	// Create the volume
	// Format: gluster volume create <volume-name> <hostname>:<brick-path> force
	brickSpec := fmt.Sprintf("%s:%s", hostname, brickPath)
	cmd := exec.Command("gluster", "volume", "create", volumeName, brickSpec, "force")

	output, err := cmd.CombinedOutput()
	if err != nil {
		fields["error"] = err
		fields["output"] = string(output)
		m.Error("Failed to create Gluster volume", fields)

		// Check if it's because volume already exists
		if strings.Contains(string(output), "already exists") {
			m.Info("Gluster volume already exists (from output)", fields)
			return nil
		}

		return fmt.Errorf("failed to create gluster volume: %w: %s", err, string(output))
	}

	fields["output"] = string(output)
	m.Info("Gluster volume created successfully", fields)
	return nil
}

// StartVolume starts a Gluster volume
func (m *VolumeManager) StartVolume(volumeName string) error {
	fields := map[string]interface{}{
		"volume": volumeName,
	}

	// Check if volume is already started
	statusCmd := exec.Command("gluster", "volume", "status", volumeName)
	if err := statusCmd.Run(); err == nil {
		// Volume is already running
		m.Info("Gluster volume already started", fields)
		return nil
	}

	// Start the volume
	cmd := exec.Command("gluster", "volume", "start", volumeName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fields["error"] = err
		fields["output"] = string(output)
		m.Error("Failed to start Gluster volume", fields)

		// Check if it's because volume is already started
		if strings.Contains(string(output), "already started") {
			m.Info("Gluster volume already started (from output)", fields)
			return nil
		}

		return fmt.Errorf("failed to start gluster volume: %w: %s", err, string(output))
	}

	fields["output"] = string(output)
	m.Info("Gluster volume started successfully", fields)
	return nil
}
