package device

import (
	"regexp"

	"github.com/sntns/gluster-provisioner/pkg/model"
)

// PathFilter matches devices based on a path pattern.
type PathFilter struct {
	Type    string `mapstructure:"type"`
	Pattern string `mapstructure:"pattern"`
}

var _ Filter = PathFilter{}

// Match returns true if the device path matches the pattern.
func (f PathFilter) Match(info model.DeviceInfo) bool {
	matched, err := regexp.MatchString(f.Pattern, info.DevName())
	if err != nil {
		// If the pattern is invalid, do not match anything
		return false
	}
	return matched
}
