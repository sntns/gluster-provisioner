//go:build linux
// +build linux

package udev

import (
	"context"
	"strconv"
	"sync"

	"github.com/pilebones/go-udev/crawler"
	"github.com/pilebones/go-udev/netlink"
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

type Controller struct {
	sync.Mutex // to protect cancel
	logger     capability.Logger
	config     Configuration
	listener   model.DeviceListener
	cancel     context.CancelFunc
}

func NewController(logger capability.Logger, config Configuration, listener model.DeviceListener) (*Controller, error) {
	return &Controller{
		logger:   logger,
		config:   config,
		listener: listener,
		cancel:   nil,
	}, nil
}

func (c *Controller) Start(ctx context.Context) error {
	c.Lock()
	defer c.Unlock()
	if c.cancel != nil {
		// already started
		return nil
	}
	var cancelCtx context.Context
	cancelCtx, c.cancel = context.WithCancel(ctx)
	// start watching in a separate goroutine
	go c.watch(cancelCtx, true)

	return nil
}

func (c *Controller) Stop(ctx context.Context) error {
	c.Lock()
	defer c.Unlock()
	if c.cancel == nil {
		// not started
		return nil
	}
	c.cancel()
	c.cancel = nil
	return nil
}

func (c *Controller) pushDevice(_ context.Context, device crawler.Device) {
	fields := map[string]any{
		"env":  device.Env,
		"kobj": device.KObj,
	}
	c.logger.Debug("Udev device event received", fields)

	seq, _ := strconv.ParseUint(device.Env["DISKSEQ"], 10, 64)
	major, _ := strconv.ParseUint(device.Env["MAJOR"], 10, 32)
	minor, _ := strconv.ParseUint(device.Env["MINOR"], 10, 32)

	info := &model.DeviceInfo{
		Path:  device.KObj,
		Name:  device.Env["DEVNAME"],
		Type:  device.Env["DEVTYPE"],
		Seq:   seq,
		Major: uint32(major),
		Minor: uint32(minor),
	}
	if c.match(*info) {
		ch := c.listener.DeviceChan()
		ch <- *info
	}
}

// watch listens to udev events according to the config and sends filtered DeviceInfo on the channel. Stops if ctx is done.
func (c *Controller) watch(ctx context.Context, enumerate bool) {
	fields := map[string]any{
		"subsystem": c.config.Subsystem,
		"enumerate": enumerate,
	}
	c.logger.Info("Starting udev monitor", fields)

	errors := make(chan error)
	crawlerQueue := make(chan crawler.Device)
	var crawlerQuit chan struct{}
	matcher := &netlink.RuleDefinitions{
		Rules: []netlink.RuleDefinition{
			{
				Action: &c.config.Action,
				Env: map[string]string{
					"SUBSYSTEM": c.config.Subsystem,
				},
			},
		},
	}

	if enumerate {
		c.logger.Info("Starting udev enumeration", fields)
		crawlerQuit = crawler.ExistingDevices(crawlerQueue, errors, matcher)
		defer close(crawlerQuit)
	}

	// Starting netlink UEvent monitor
	conn := new(netlink.UEventConn)
	if err := conn.Connect(netlink.UdevEvent); err != nil {
		c.logger.Error("udev monitor connect error", map[string]any{
			"error": err,
		})
		return
	}
	defer conn.Close()

	eventQueue := make(chan netlink.UEvent)
	monitorQuit := conn.Monitor(eventQueue, errors, matcher)
	defer close(monitorQuit)

	c.logger.Info("Udev monitor started", map[string]any{})
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case uevent, ok := <-eventQueue:
			if !ok {
				eventQueue = nil
				continue
			}
			fields := map[string]any{
				"action": uevent.Action,
				"kobj":   uevent.KObj,
				"env":    uevent.Env,
			}
			device := crawler.Device{
				KObj: uevent.KObj,
				Env:  uevent.Env,
			}
			c.logger.Debug("Udev device event received", fields)
			c.pushDevice(ctx, device)
		case device, ok := <-crawlerQueue:
			if !ok {
				crawlerQueue = nil
				continue
			}
			fields := map[string]any{
				"kobj": device.KObj,
				"env":  device.Env,
			}
			c.logger.Debug("Udev device crawled", fields)
			c.pushDevice(ctx, device)
		case err := <-errors:
			c.logger.Error("Netlink error", map[string]interface{}{
				"error": err,
			})
		}
	}

	c.logger.Info("Udev monitor stopped", map[string]any{})
}

// match délègue à Filters.Match la logique de filtrage
func (c *Controller) match(info model.DeviceInfo) bool {
	return c.config.Filters.Match(info)
}
