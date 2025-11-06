package layer

import (
	"context"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

type Formatted struct {
	capability.Logger
	DiskManager model.DiskManager
}

type FormattedState struct {
	Device      model.Disk
	Filesystems model.Filesystems
}

func NewFormatted(logger capability.Logger, manager model.DiskManager) *Formatted {
	return &Formatted{
		Logger:      logger,
		DiskManager: manager,
	}
}

func (s *Formatted) Up(ctx context.Context, state *State) error {
	partitionedState := state.Partitioned
	if partitionedState == nil {
		return ErrInvalidState
	}
	device := partitionedState.Device
	partitions := partitionedState.Partitions
	fields := map[string]interface{}{
		"device":     device,
		"partitions": partitions,
	}
	s.Debug("Formatting disk", fields)
	filesystems, err := s.DiskManager.Format(ctx, device, partitions)
	if err != nil {
		return err
	}

	state.Formatted = &FormattedState{
		Device:      device,
		Filesystems: filesystems,
	}
	return nil
}

func (s *Formatted) Down(ctx context.Context, state *State) error {
	return nil
}

func (s *Formatted) Active(ctx context.Context, state *State) (bool, error) {
	return false, nil
}

func (s *Formatted) String() string {
	return "formatted"
}

func (s *Formatted) Dependencies() []string {
	return []string{"partitioned"}
}
