package model

import (
	"context"
)

type DiskFetcher interface {
	DiskFetchContext(ctx context.Context, device DeviceInfo) (*DiskMetadata, error)
}

type DiskManager interface {
	Partition(ctx context.Context, disk Disk) (Partitions, error)
	Format(ctx context.Context, disk Disk, partitions Partitions) (Filesystems, error)
	Mount(ctx context.Context, disk Disk, filesystems Filesystems) (Mountpoints, error)
	Unmount(ctx context.Context, disk Disk, filesystems Filesystems) error
}

type Disk struct {
	Path   string
	Layout DiskLayout
}

func (d Disk) String() string {
	return d.Path
}

type PartitionTableType string

const (
	PartitionTableTypeGPT PartitionTableType = "gpt"
	PartitionTableTypeMBR PartitionTableType = "mbr"
)

type DiskLayout struct {
	Type       PartitionTableType
	Partitions []PartitionLayout
}

type PartitionLayout struct {
	SizeGB     uint64
	Label      string
	Filesystem string
	MountPoint string
	Options    map[string]*string
}

type DiskMetadata struct {
	Name string
	Tags []string
}
