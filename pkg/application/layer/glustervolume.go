package layer

import (
	"context"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

type GlusterVolume struct {
	capability.Logger
	VolumeManager model.GlusterVolumeManager
}

type GlusterVolumeState struct {
	Volumes model.GlusterVolumes
}

func NewGlusterVolume(logger capability.Logger, manager model.GlusterVolumeManager) *GlusterVolume {
	return &GlusterVolume{
		Logger:        logger,
		VolumeManager: manager,
	}
}

func (s *GlusterVolume) Up(ctx context.Context, state *State) error {
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

		fields["volume"] = volumeName
		fields["brick_path"] = brickPath

		s.Info("Creating Gluster volume", fields)

		// Create the volume
		err := s.VolumeManager.CreateVolume(volumeName, brickPath)
		if err != nil {
			fields["error"] = err
			s.Error("Failed to create Gluster volume", fields)
			return err
		}

		// Start the volume
		err = s.VolumeManager.StartVolume(volumeName)
		if err != nil {
			fields["error"] = err
			s.Error("Failed to start Gluster volume", fields)
			return err
		}

		volumes = append(volumes, model.GlusterVolume{
			Name:      volumeName,
			BrickPath: brickPath,
			Started:   true,
		})

		s.Info("Gluster volume created and started", fields)
	}

	state.GlusterVolume = &GlusterVolumeState{
		Volumes: volumes,
	}

	return nil
}

func (s *GlusterVolume) Down(ctx context.Context, state *State) error {
	return nil
}

func (s *GlusterVolume) String() string {
	return "gluster-volume"
}

func (s *GlusterVolume) Dependencies() []string {
	return []string{"mounted"}
}
