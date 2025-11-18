package application

import (
	"context"
	"sync"
	"time"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

type Application struct {
	sync.Mutex
	logger         capability.Logger
	diskManager    model.DiskManager
	fetcher        model.DiskFetcher
	glusterManager model.GlusterVolumeManager
	deviceChan     chan model.DeviceInfo
	cancel         context.CancelFunc
}

func NewApplication(logger capability.Logger, diskManager model.DiskManager, fetcher model.DiskFetcher, glusterManager model.GlusterVolumeManager) *Application {
	return &Application{
		logger:         logger,
		diskManager:    diskManager,
		fetcher:        fetcher,
		glusterManager: glusterManager,
		deviceChan:     make(chan model.DeviceInfo, 1),
		cancel:         nil,
	}
}

func (a *Application) Start(ctx context.Context) error {
	a.Lock()
	defer a.Unlock()
	if a.cancel != nil {
		// already started
		return nil
	}

	// create a cancellable context
	cancelContext, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	// start processing in a separate goroutine
	go a.start(cancelContext)
	return nil
}

func (a *Application) Stop(ctx context.Context) error {
	a.Lock()
	defer a.Unlock()
	if a.cancel == nil {
		// not started
		return nil
	}
	a.cancel()
	a.cancel = nil
	return nil
}

func (a *Application) DeviceChan() chan model.DeviceInfo {
	if a.cancel == nil {
		_ = a.Start(context.Background())
	}
	return a.deviceChan
}

func (a *Application) start(ctx context.Context) error {
	a.logger.Info("Application started", map[string]any{})
	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Application stopped", map[string]any{})
			return nil
		case dev := <-a.deviceChan:
			ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
			// process device event
			err := a.handleDevice(ctx, dev)
			if err != nil {
				a.logger.Error("Failed to handle device", map[string]any{
					"device": dev.Path,
					"error":  err,
				})
			}
			cancel()
		}
	}
}
