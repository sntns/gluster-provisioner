package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/device/udev"
	"github.com/sntns/gluster-provisioner/pkg/model"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type simpleListener struct {
	ch chan model.DeviceInfo
}

func NewSimpleListener() model.DeviceListener {
	return &simpleListener{
		ch: make(chan model.DeviceInfo, 100),
	}
}

func (l *simpleListener) DeviceChan() chan model.DeviceInfo {
	return l.ch
}

var enumerateCmd = &cobra.Command{
	Use:   "enumerate",
	Short: "Enumerate all devices on the system",
	Run: func(cmd *cobra.Command, args []string) {
		err := Run(
			capability.WithZapLogger(),
			capability.WithViperConfigurationLoader(
				capability.ViperLoaderWithPath(&path),
				capability.ViperLoaderWithFileType("yaml"),
			),
			fx.Provide(NewSimpleListener),
			udev.WithUdev(),
			fx.Invoke(func(lc fx.Lifecycle, listener model.DeviceListener, logger capability.Logger) {
				logger.Info("Starting device enumeration", map[string]any{})
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() {
							fmt.Printf("Devices:\n")
							for info := range listener.DeviceChan() {
								value, _ := json.MarshalIndent(info, "", " ")
								fmt.Printf("DeviceInfo: %+v\n", info.Name)
								fmt.Printf("%s\n", value)
							}
							fmt.Printf("Enumeration complete.\n")
						}()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						return nil
					},
				})

			}),
		)
		if err != nil {
			log.Fatalf("Error running listener: %v", err)
		}
	},
}
