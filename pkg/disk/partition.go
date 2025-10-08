package disk

import (
	"fmt"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/gpt"
)

// CreatePartitionTable creates a GPT partition table and a single partition
func CreatePartitionTable(dev string) error {
	disk, err := diskfs.Open(dev)
	if err != nil {
		return fmt.Errorf("open disk: %w", err)
	}
	// Create GPT partition table if not present
	err = disk.Partition(&diskfs.PartitionTable{
		Type: diskfs.PartitionTableTypeGPT,
		Table: &gpt.Table{
			Partitions: []gpt.Partition{
				{
					Start: 2048,
					End:   disk.Size/512 - 1,
					Type:  gpt.LinuxFilesystem,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("partition: %w", err)
	}
	return nil
}
