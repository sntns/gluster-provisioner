package model

import (
	"fmt"
	"strings"
)

type DeviceListener interface {
	DeviceChan() chan DeviceInfo
}

// DeviceInfo groups relevant udev device information for matching
type DeviceInfo struct {
	Name  string
	Path  string
	Type  string
	Seq   uint64
	Major uint32
	Minor uint32
}

func (d DeviceInfo) DevName() string {
	if strings.HasPrefix(d.Name, "/dev/") {
		return d.Name
	}
	return fmt.Sprintf("/dev/%s", d.Name)
}

func (d DeviceInfo) String() string {
	return d.DevName()
}
