package layer

import (
	"context"
	"path/filepath"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

type Glusterd struct {
	capability.Logger
	GlusterManager model.GlusterVolumeManager
}

type GlusterdState struct {
	Volumes model.GlusterVolumes
}

func NewGlusterd(logger capability.Logger, manager model.GlusterVolumeManager) *Glusterd {
	return &Glusterd{
		Logger:         logger,
		GlusterManager: manager,
	}
}

func (s *Glusterd) Up(ctx context.Context, state *State) error {
	mountedState := state.Mounted
	if mountedState == nil {
		return ErrInvalidState
	}

	fields := map[string]interface{}{
		"device":      mountedState.Device,
		"mountpoints": len(mountedState.Mountpoints),
	}

	s.Debug("Creating Gluster volumes", fields)

	var volumes model.GlusterVolumes
	for _, mountpoint := range mountedState.Mountpoints {
		volumeName := mountpoint.Label
		brickPath := mountpoint.Path + "/brick"
		mountPoint := filepath.Join("/media/gluster", volumeName)

		fields["volume"] = volumeName
		fields["brick_path"] = brickPath
		fields["mount_point"] = mountPoint

		s.Info("Creating Gluster volume", fields)

		// Create the volume
		err := s.GlusterManager.CreateVolume(volumeName, brickPath)
		if err != nil {
			fields["error"] = err
			s.Error("Failed to create Gluster volume", fields)
			return err
		}

		// Start the volume
		err = s.GlusterManager.StartVolume(volumeName)
		if err != nil {
			fields["error"] = err
			s.Error("Failed to start Gluster volume", fields)
			return err
		}

		// Mount the volume via FUSE so the host can access it through the shared bind mount.
		err = s.GlusterManager.MountVolume(volumeName, mountPoint)
		if err != nil {
			fields["error"] = err
			s.Error("Failed to mount Gluster volume via FUSE", fields)
			return err
		}

		volumes = append(volumes, model.GlusterVolume{
			Name:      volumeName,
			BrickPath: brickPath,
			Started:   true,
		})

		s.Info("Gluster volume created and started", fields)
	}

	state.Glusterd = &GlusterdState{
		Volumes: volumes,
	}

	return nil
}

func (s *Glusterd) Down(ctx context.Context, state *State) error {
	return nil
}

func (s *Glusterd) String() string {
	return "glusterd"
}

func (s *Glusterd) Dependencies() []string {
	return []string{"mounted"}
}
