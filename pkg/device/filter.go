package device

import "github.com/sntns/gluster-provisioner/pkg/model"

// Filter is the interface for all device filters.
type Filter interface {
	Match(info model.DeviceInfo) bool
}
