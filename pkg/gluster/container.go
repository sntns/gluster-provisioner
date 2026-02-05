package gluster

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
		GlusterManager model.GlusterVolumeManager
	}

	provide := func(in provideIn) (out provideOut, err error) {
		manager := NewManager(in.Logger)
		out.GlusterManager = manager
		return
	}

	return fx.Provide(provide)
}
