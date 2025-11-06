package disk

// PrepareAndMount is a high-level helper that calls all steps in order
import (
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

var _ model.DiskManager = (*Manager)(nil)

type Manager struct {
	capability.Logger
}

func NewManager(logger capability.Logger) *Manager {
	return &Manager{
		Logger: logger,
	}
}
