package layer

import (
	"context"

	"github.com/sntns/gluster-provisioner/pkg/model"
)

type Ready struct {
	Fetcher model.DiskFetcher
}

type ReadyState struct {
	Device model.Disk
}

func NewReady(fetcher model.DiskFetcher) *Ready {
	return &Ready{
		Fetcher: fetcher,
	}
}

func (s *Ready) Up(ctx context.Context, state *State) error {
	discoveredState := state.Discovered
	if discoveredState == nil {
		return ErrInvalidState
	}

	metadata, err := s.Fetcher.DiskFetchContext(ctx, discoveredState.Device)
	if err != nil {
		return err
	}
	if metadata == nil {
		return ErrMetadataNotFound
	}

	device := model.Disk{
		Path: discoveredState.Device.DevName(),
		Layout: model.DiskLayout{
			Type: model.PartitionTableTypeGPT,
			Partitions: []model.PartitionLayout{
				{
					Label:      metadata.Name,
					Filesystem: "ext4",
					MountPoint: "/media/" + metadata.Name,
				},
			},
		},
	}

	state.Ready = &ReadyState{
		Device: device,
	}
	return nil
}

func (s *Ready) Down(ctx context.Context, state *State) error {
	return nil
}

func (s *Ready) String() string {
	return "ready"
}

func (s *Ready) Dependencies() []string {
	return []string{}
}
