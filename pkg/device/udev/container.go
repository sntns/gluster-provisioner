package udev

import (
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
	"go.uber.org/fx"
)

func WithDefaultConfig() fx.Option {
	return fx.Provide(DefaultConfiguration)
}

func WithConfig(config Configuration) fx.Option {
	return fx.Provide(func() Configuration {
		return config
	})
}

func WithUdev() fx.Option {
	type invokeIn struct {
		fx.In
		fx.Lifecycle
		Loader   capability.Loader `name:"configuration"`
		Logger   capability.Logger
		Listener model.DeviceListener
	}

	invoke := func(in invokeIn) error {
		var err error
		configuration := Configuration{}
		if err = in.Loader.Load("adapter.udev", &configuration); err != nil {
			return err
		}
		if err = configuration.Validate(); err != nil {
			return err
		}
		var controller *Controller
		controller, err = NewController(in.Logger, configuration, in.Listener)
		in.Lifecycle.Append(fx.Hook{
			OnStart: controller.Start,
			OnStop:  controller.Stop,
		})
		return nil
	}

	return fx.Invoke(invoke)
}
