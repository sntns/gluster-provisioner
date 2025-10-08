package contextapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sntns/gluster-provisioner-listener/pkg/model"
)

type openstackMeta struct {
	Name    string          `json:"name"`
	Devices []openstackDisk `json:"devices"`
}

type openstackDisk struct {
	Address string   `json:"address"`
	Tags    []string `json:"tags"`
}

// FetchContext fetches disk metadata from the API (generic, calls OpenStack by default)
func FetchContext(dev string) (*model.DiskMetadata, error) {
	return fetchOpenStackContext(dev)
}

// fetchOpenStackContext fetches metadata from OpenStack metadata API
func fetchOpenStackContext(dev string) (*model.DiskMetadata, error) {
	resp, err := http.Get("http://169.254.169.254/openstack/latest/meta_data.json")
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
	var meta openstackMeta
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, err
	}
	var metadata *model.DiskMetadata
	for _, d := range meta.Devices {
		if d.Address == dev {
			if len(d.Tags) > 0 {
				metadata = &model.DiskMetadata{
					Name: d.Tags[0],
					Tags: d.Tags[1:],
				}
			}
			break
		}
	}
	return metadata, nil
}
