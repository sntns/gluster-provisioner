package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

const OPENSTACK_METADATA_URL = "http://169.254.169.254/openstack/latest/meta_data.json"

var (
	maxWait   = 30 * time.Second
	pollEvery = 2000 * time.Millisecond
)

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
	// OpenStack IMDS can lag behind hot-attached volumes (eventual consistency).
	// If udev sees the disk before IMDS lists it, the match will be nil on the first try.
	// Retry for a bounded time to avoid requiring a service restart.
	deadline := time.Now().Add(maxWait)
	var lastErr error
	for {
		metadata, err := f.fetchOnce(ctx, device)
		if err == nil && metadata != nil {
			return metadata, nil
		}
		if err != nil {
			lastErr = err
		}

		if time.Now().After(deadline) {
			if lastErr != nil {
				return nil, fmt.Errorf("openstack metadata lookup timed out after %s: %w", maxWait, lastErr)
			}
			return nil, nil
		}

		time.Sleep(pollEvery)
	}
}

func (f *OpenStackFetcher) fetchOnce(ctx context.Context, device model.DeviceInfo) (*model.DiskMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, OPENSTACK_METADATA_URL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
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
