package udev

import (
	"strings"

	"github.com/sntns/gluster-provisioner/pkg/model"
)

type Filter interface {
	Match(info model.DeviceInfo) bool
}

type PrefixFilter struct {
	Prefix string `json:"prefix"`
}

var _ Filter = PrefixFilter{}

func (f PrefixFilter) Match(info model.DeviceInfo) bool {
	return strings.HasPrefix(info.DevNode, f.Prefix)
}
