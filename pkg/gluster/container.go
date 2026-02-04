package gluster

import (
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
	"go.uber.org/fx"
)

func WithManager() fx.Option {
	type provideIn struct {
		fx.In
		Loader capability.Loader `name:"configuration"`
		Logger capability.Logger
	}

	type provideOut struct {
		fx.Out
		GlusterManager model.GlusterVolumeManager
	}

	provide := func(in provideIn) (out provideOut, err error) {
		configuration := Configuration{}
		if err = in.Loader.Load("gluster.volume", &configuration); err != nil {
			return
		}
		if err = configuration.Validate(); err != nil {
			return
		}
		manager := NewManager(in.Logger, configuration)
		out.GlusterManager = manager
		return
	}

	return fx.Provide(provide)
}
