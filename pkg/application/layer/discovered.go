package layer

import (
	"context"

	"github.com/sntns/gluster-provisioner/pkg/model"
)

type Discovered struct {
	Device model.DeviceInfo
}

type DiscoveredState struct {
	Device model.DeviceInfo
}

func NewDiscovered(device model.DeviceInfo) *Discovered {
	return &Discovered{
		Device: device,
	}
}

func (s *Discovered) Up(ctx context.Context, state *State) error {
	return nil
}

func (s *Discovered) Down(ctx context.Context, state *State) error {
	return nil
}

func (s *Discovered) String() string {
	return "discovered"
}

func (s *Discovered) Dependencies() []string {
	return []string{}
}
