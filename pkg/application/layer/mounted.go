package layer

import (
	"context"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

type Mounted struct {
	capability.Logger
	DiskManager model.DiskManager
}

type MountedState struct {
	Device      model.Disk
	Mountpoints model.Mountpoints
}

func NewMounted(logger capability.Logger, manager model.DiskManager) *Mounted {
	return &Mounted{
		Logger:      logger,
		DiskManager: manager,
	}
}

func (s *Mounted) Up(ctx context.Context, state *State) error {
	formattedState := state.Formatted
	if formattedState == nil {
		return ErrInvalidState
	}
	device := formattedState.Device
	fields := map[string]interface{}{
		"device": device,
	}

	s.Debug("Mounting disk", fields)
	mountpoints, err := s.DiskManager.Mount(ctx, device, formattedState.Filesystems)
	if err != nil {
		return err
	}
	state.Mounted = &MountedState{
		Device:      device,
		Mountpoints: mountpoints,
	}
	return nil
}

func (s *Mounted) Down(ctx context.Context, state *State) error {
	return nil
}

func (s *Mounted) Active(ctx context.Context, state *State) (bool, error) {
	return false, nil
}

func (s *Mounted) String() string {
	return "mounted"
}

func (s *Mounted) Dependencies() []string {
	return []string{"formatted"}
}
