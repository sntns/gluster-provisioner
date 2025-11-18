package gluster

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/restclient"
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

var _ model.GlusterVolumeManager = &Manager{}

type Manager struct {
	capability.Logger
	client *restclient.Client
}

// NewManager creates a new Gluster manager using glusterd2 REST client
// It connects to the local glusterd2 REST API endpoint
func NewManager(logger capability.Logger) *Manager {
	// Connect to local glusterd2 REST API
	// Default glusterd2 REST endpoint is http://localhost:24007
	client := restclient.New("http://localhost:24007", "", "", "", true)

	return &Manager{
		Logger: logger,
		client: client,
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

	// Get the hostname for the brick specification
	hostname, err := os.Hostname()
	if err != nil {
		fields["error"] = err
		m.Error("Failed to get hostname", fields)
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	// Check if volume already exists
	volumes, err := m.client.Volumes(volumeName)
	if err == nil && len(volumes) > 0 {
		// Volume already exists
		m.Info("Gluster volume already exists", fields)
		return nil
	}

	// Create volume request
	// We'll get peer ID from the local peer info
	peerID := hostname // Using hostname as peer identifier for now

	req := api.VolCreateReq{
		Name: volumeName,
		Subvols: []api.SubvolReq{
			{
				Type: "distribute",
				Bricks: []api.BrickReq{
					{
						PeerID: peerID,
						Path:   brickPath,
					},
				},
			},
		},
		Force: true,
		Flags: map[string]bool{
			"create-brick-dir": true,
		},
	}

	// Create the volume using REST API
	_, err = m.client.VolumeCreate(req)
	if err != nil {
		fields["error"] = err
		m.Error("Failed to create Gluster volume", fields)
		return fmt.Errorf("failed to create gluster volume: %w", err)
	}

	fields["peer_id"] = peerID
	m.Info("Gluster volume created successfully", fields)
	return nil
}

// StartVolume starts a Gluster volume
func (m *Manager) StartVolume(volumeName string) error {
	fields := map[string]interface{}{
		"volume": volumeName,
	}

	// Check if volume is already started
	status, err := m.client.VolumeStatus(volumeName)
	if err == nil && status.Info.State == api.VolStarted {
		// Volume is already running
		m.Info("Gluster volume already started", fields)
		return nil
	}

	// Start the volume using REST API
	err = m.client.VolumeStart(volumeName, false)
	if err != nil {
		fields["error"] = err
		m.Error("Failed to start Gluster volume", fields)
		return fmt.Errorf("failed to start gluster volume: %w", err)
	}

	m.Info("Gluster volume started successfully", fields)
	return nil
}
