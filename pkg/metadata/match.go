package metadata

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sntns/gluster-provisioner/pkg/model"
)

func Match(meta Meta, device model.DeviceInfo) (*model.DiskMetadata, error) {
	// Depending on the udev source, the kernel object path may be provided as
	// "/devices/..." (kobject) instead of a full sysfs path "/sys/devices/...".
	// OpenStack metadata matching uses sysfs paths, so normalize here.
	path := device.Path
	if strings.HasPrefix(path, "/devices/") {
		path = "/sys" + path
	}

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
		if matched, err := regexp.MatchString(pattern, path); err != nil {
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
