package metadata

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

// FileFetcher fetches disk metadata from a local file (for testing)
type FileFetcher struct {
	Logger capability.Logger
	Path   string
}

func NewFileFetcher(path string, logger capability.Logger) *FileFetcher {
	return &FileFetcher{
		Logger: logger,
		Path:   path,
	}
}

func (f *FileFetcher) DiskFetchContext(ctx context.Context, device model.DeviceInfo) (*model.DiskMetadata, error) {
	fields := map[string]any{
		"address": device,
		"path":    f.Path,
	}
	f.Logger.Debug("Fetching disk metadata from file", fields)
	file, err := os.Open(f.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	body, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var meta Meta
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, err
	}
	metadata, err := match(meta, device)
	if err != nil {
		return nil, err
	}
	fields["metadata"] = metadata
	f.Logger.Debug("Fetched disk metadata from file", fields)
	return metadata, nil
}
