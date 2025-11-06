package disk

import (
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
	"go.uber.org/fx"
)

func WithManager() fx.Option {
	type provideIn struct {
		fx.In
		Logger capability.Logger
	}

	type provideOut struct {
		fx.Out
		Manager model.DiskManager
	}

	provide := func(in provideIn) provideOut {
		return provideOut{
			Manager: NewManager(in.Logger),
		}
	}

	return fx.Provide(provide)
}
