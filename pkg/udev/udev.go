//go:build linux
// +build linux

package udev

import (
	"context"
	"log"

	"github.com/jochenvg/go-udev"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

type Controller struct {
	config Configuration
}

func NewController(config Configuration) *Controller {
	return &Controller{config: config}
}

// Watch listens to udev events according to the config and sends filtered DeviceInfo on the channel. Stops if ctx is done.
func (c *Controller) Watch(ctx context.Context, ch chan<- model.DeviceInfo) {
	deviceChan, err := udev.NewMonitorFromNetlink("udev").
		FilterAddMatchSubsystem(c.config.Subsystem).
		FilterAddMatchAction("add").
		DeviceChan(nil)
	if err != nil {
		log.Fatalf("udev monitor error: %v", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case dev, ok := <-deviceChan:
			if !ok {
				return
			}
			if dev.Devnode() != "" {
				info := model.DeviceInfo{
					Name:       dev.Sysname(),
					Num:        dev.Sysnum(),
					Path:       dev.Syspath(),
					DevNode:    dev.Devnode(),
					DevType:    dev.Devtype(),
					Driver:     dev.Driver(),
					Properties: dev.Properties(),
				}
				if c.match(info) {
					ch <- info
				}
			}
		}
	}
}

// match checks if any filter matches one of the device's properties
func (c *Controller) match(info model.DeviceInfo) bool {
	for _, filter := range c.config.Filters {
		if filter.Match(info) {
			return true
		}
	}
	return false
}
