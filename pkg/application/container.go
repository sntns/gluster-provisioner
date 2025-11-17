package application

import (
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
	"go.uber.org/fx"
)

func WithApplication() fx.Option {
	type provideIn struct {
		fx.In
		Logger        capability.Logger
		DiskManager   model.DiskManager
		Fetcher       model.DiskFetcher
		VolumeManager model.GlusterVolumeManager
	}

	type provideOut struct {
		fx.Out
		DeviceListener model.DeviceListener
	}

	provide := func(in provideIn) provideOut {
		app := NewApplication(in.Logger, in.DiskManager, in.Fetcher, in.VolumeManager)
		return provideOut{
			DeviceListener: app,
		}
	}

	return fx.Provide(provide)
}
