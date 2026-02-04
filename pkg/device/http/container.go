package http

import (
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/device"
	"github.com/sntns/gluster-provisioner/pkg/model"
	"go.uber.org/fx"
)

func WithDeviceHttp() fx.Option {
	type invokeIn struct {
		fx.In
		fx.Lifecycle
		Logger   capability.Logger
		Listener model.DeviceListener
	}

	invoke := func(in invokeIn) (err error) {
		// Provide a default configuration with no filters
		defaultConfig := Configuration{
			Filters: device.Filters{
				Items: nil,
				Op:    device.OpOr,
			},
		}
		controller := NewController(in.Logger, in.Listener, defaultConfig)
		in.Lifecycle.Append(fx.Hook{
			OnStart: controller.Start,
			OnStop:  controller.Stop,
		})
		return nil
	}

	return fx.Invoke(invoke)
}
