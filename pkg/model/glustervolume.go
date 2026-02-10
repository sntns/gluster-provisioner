package model

// GlusterVolumeManager defines the interface for managing Gluster volumes
type GlusterVolumeManager interface {
	CreateVolume(volumeName string, brickPath string) error
	StartVolume(volumeName string) error
	EnsureMounted(volumeName string, mountPoint string) error
	VolumeExists(volumeName string) bool
	VolumeStarted(volumeName string) bool
}

// GlusterVolume represents a created Gluster volume
type GlusterVolume struct {
	Name      string
	BrickPath string
	Started   bool
}

// GlusterVolumes is a collection of Gluster volumes
type GlusterVolumes []GlusterVolume
