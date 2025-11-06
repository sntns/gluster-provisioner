package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

const OPENSTACK_METADATA_URL = "http://169.254.169.254/openstack/latest/meta_data.json"

// Use Meta from local package

// OpenStackFetcher fetches disk metadata from OpenStack metadata API
type OpenStackFetcher struct {
	Logger capability.Logger
}

func NewOpenStackFetcher(logger capability.Logger) *OpenStackFetcher {
	return &OpenStackFetcher{
		Logger: logger,
	}
}

func (f *OpenStackFetcher) DiskFetchContext(ctx context.Context, device model.DeviceInfo) (*model.DiskMetadata, error) {
	resp, err := http.Get(OPENSTACK_METADATA_URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var meta Meta
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, err
	}
	return match(meta, device)
}
