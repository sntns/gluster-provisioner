package main

import (
	"context"
	"log"

	"go.uber.org/fx"

	"github.com/sntns/gluster-provisioner-listener/pkg/contextapi"
	"github.com/sntns/gluster-provisioner-listener/pkg/disk"
	"github.com/sntns/gluster-provisioner-listener/pkg/udev"
)

func RunListener(lc fx.Lifecycle, watcher udev.Watcher, fetchContext contextapi.Fetcher, prepareAndMount disk.Preparer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Starting gluster-provisioner-listener (fx)...")
			udevChan := make(chan string)
			go watcher.WatchDisks(udevChan)
			go func() {
				for dev := range udevChan {
					log.Printf("New disk detected: %s", dev)
					ctxData, err := fetchContext.FetchContext(dev)
					if err != nil {
						log.Printf("Failed to fetch context: %v", err)
						continue
					}
					log.Printf("Context for %s: name=%s tags=%v", dev, ctxData.Name, ctxData.Tags)
					if err := prepareAndMount.PrepareAndMount(dev, ctxData.Name); err != nil {
						log.Printf("Failed to prepare/mount disk: %v", err)
					}
				}
			}()
			return nil
		},
	})
}

func main() {
	fx.New(
		fx.Provide(
			udev.NewWatcherFX,
			contextapi.NewFetcher,
			disk.NewPreparer,
		),
		fx.Invoke(RunListener),
	).Run()
}
