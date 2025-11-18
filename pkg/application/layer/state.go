package layer

import "errors"

var (
	ErrInvalidState     = errors.New("invalid state for layer operation")
	ErrMetadataNotFound = errors.New("disk metadata not found")
)

type State struct {
	Discovered  *DiscoveredState
	Ready       *ReadyState
	Partitioned *PartitionedState
	Formatted   *FormattedState
	Mounted     *MountedState
	Glusterd    *GlusterdState
}
