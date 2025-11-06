package layer

import (
	"context"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

type Partitioned struct {
	capability.Logger
	DiskManager model.DiskManager
}

type PartitionedState struct {
	Device     model.Disk
	Partitions model.Partitions
}

func NewPartitioned(logger capability.Logger, manager model.DiskManager) *Partitioned {
	return &Partitioned{
		Logger:      logger,
		DiskManager: manager,
	}
}

func (s *Partitioned) Up(ctx context.Context, state *State) error {
	readyState := state.Ready
	if readyState == nil {
		return ErrInvalidState
	}
	device := readyState.Device
	fields := map[string]interface{}{
		"device": device,
	}
	s.Debug("Partitioning disk", fields)
	partitions, err := s.DiskManager.Partition(ctx, device)
	if err != nil {
		return err
	}
	fields["partitions"] = partitions
	s.Info("Disk partitioned", fields)

	state.Partitioned = &PartitionedState{
		Device:     device,
		Partitions: partitions,
	}
	return nil
}

func (s *Partitioned) Down(ctx context.Context, state *State) error {
	return nil
}

func (s *Partitioned) String() string {
	return "partitioned"
}

func (s *Partitioned) Dependencies() []string {
	return []string{"ready"}
}
