package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/disk"
	"github.com/sntns/gluster-provisioner/pkg/model"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var partitionCmd = &cobra.Command{
	Use:   "partition",
	Short: "Partition a disk",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := Run(
			capability.WithZapLogger(),
			capability.WithViperConfigurationLoader(
				capability.ViperLoaderWithPath(&path),
				capability.ViperLoaderWithFileType("yaml"),
			),
			disk.WithManager(),
			fx.Invoke(func(lc fx.Lifecycle, disk model.DiskManager, logger capability.Logger) {
				logger.Info("Starting device enumeration", map[string]any{})
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						device := model.Disk{
							Path: args[0],
							Layout: model.DiskLayout{
								Type: model.PartitionTableTypeGPT,
								Partitions: []model.PartitionLayout{
									{
										Label: "obss",
									},
								},
							},
						}
						partitions, err := disk.Partition(ctx, device)
						if err != nil {
							return err
						}

						output, _ := json.MarshalIndent(partitions, "", " ")
						fmt.Printf("Partitions:\n%s\n", output)
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
