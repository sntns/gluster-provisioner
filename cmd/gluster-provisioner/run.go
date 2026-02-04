package main

import (
	"log"

	"github.com/sntns/gluster-provisioner/pkg/application"
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/device/udev"
	"github.com/sntns/gluster-provisioner/pkg/disk"
	"github.com/sntns/gluster-provisioner/pkg/gluster"
	"github.com/sntns/gluster-provisioner/pkg/metadata/openstack"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the gluster-provisioner listener",
	Run: func(cmd *cobra.Command, args []string) {
		err := Run(
			capability.WithZapLogger(),
			capability.WithViperConfigurationLoader(
				capability.ViperLoaderWithPath(&path),
				capability.ViperLoaderWithFileType("yaml"),
			),
			application.WithApplication(),
			udev.WithUdev(),
			openstack.WithOpenstack(),
			disk.WithManager(),
			gluster.WithManager(),
		)
		if err != nil {
			log.Fatalf("Error running listener: %v", err)
		}
	},
}
