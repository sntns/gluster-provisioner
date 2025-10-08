package disk

// PrepareAndMount is a high-level helper that calls all steps in order
func PrepareAndMount(dev, name string) error {
	if err := CreatePartitionTable(dev); err != nil {
		return err
	}
	part := dev + "1"
	if err := FormatExt4(part); err != nil {
		return err
	}
	if err := MountPartition(part, name); err != nil {
		return err
	}
	return nil
}
