package metadata

import (
	"fmt"
	"regexp"

	"github.com/sntns/gluster-provisioner/pkg/model"
)

func match(meta Meta, device model.DeviceInfo) (*model.DiskMetadata, error) {
	var metadata *model.DiskMetadata
	for _, d := range meta.Devices {
		if d.Type != "disk" {
			continue
		}
		pattern := ""
		switch d.Bus {
		case "pci":
			pattern = fmt.Sprintf("^/sys/devices/%s.*/.*%s/.*/block/.*$", d.Bus, d.Address)
		case "virtual":
			pattern = fmt.Sprintf("^/sys/devices/%s/block/%s$", d.Bus, d.Address)
		default:
			continue
		}
		if matched, err := regexp.MatchString(pattern, device.Path); err != nil {
			return nil, err
		} else if !matched {
			continue
		}
		if len(d.Tags) > 0 {
			metadata = &model.DiskMetadata{
				Name: d.Tags[0],
				Tags: d.Tags[1:],
			}
		}
		break
	}
	return metadata, nil
}
