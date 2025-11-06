package application

import (
	"context"

	"github.com/sntns/gluster-provisioner/pkg/model"
)

var _ model.DeviceListener = &Application{}

func (a *Application) handleDevice(ctx context.Context, dev model.DeviceInfo) error {
	fields := map[string]any{
		"name": dev.Name,
		"path": dev.Path,
	}
	a.logger.Info("New device detected", fields)

	go func() {
		ctx := context.Background()
		stack := NewDeviceStack(a, dev)
		for {
			fields["previous"] = stack.CurrentName()
			err := stack.MoveUp(ctx)
			fields["layer"] = stack.CurrentName()
			if err != nil {
				if err == ErrInvalidTransition {
					a.logger.Info("Device reached the top layer", fields)
					break
				}
				fields["error"] = err
				a.logger.Error("Failed Device transition", fields)
				return
			}
			a.logger.Info("Device transitioned to new layer", fields)
		}
	}()
	return nil
}
