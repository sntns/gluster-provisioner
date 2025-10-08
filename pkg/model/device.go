package model

// DeviceInfo groups relevant udev device information for matching
type DeviceInfo struct {
	Name       string
	Num        uint64
	Path       string
	DevNode    string
	DevType    string
	Driver     string
	Properties map[string]string
}
