package udev

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/sntns/gluster-provisioner/pkg/device"
)

type Configuration struct {
	Subsystem string         `json:"subsystem"`
	Action    string         `json:"action"`
	Filters   device.Filters `json:"filters"`
}

// Validate checks the Configuration fields using ozzo-validation
func (c Configuration) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Subsystem, validation.Required),
		validation.Field(&c.Action, validation.Required),
	)
}

func DefaultConfiguration() Configuration {
	return Configuration{
		Subsystem: "block",
		Action:    "add",
		Filters: device.Filters{
			Items: []device.Filter{device.PathFilter{Pattern: "/dev/vd*"}},
			Op:    device.OpOr,
		},
	}
}
