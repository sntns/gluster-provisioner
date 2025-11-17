package gluster

import (
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
	"go.uber.org/fx"
)

func WithVolumeManager() fx.Option {
	type provideIn struct {
		fx.In
		Logger capability.Logger
	}

	type provideOut struct {
		fx.Out
		VolumeManager model.GlusterVolumeManager
	}

	provide := func(in provideIn) provideOut {
		manager := NewVolumeManager(in.Logger)
		return provideOut{
			VolumeManager: manager,
		}
	}

	return fx.Provide(provide)
}
