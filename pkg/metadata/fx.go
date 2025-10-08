package contextapi

import "github.com/sntns/gluster-provisioner-listener/pkg/model"

type Fetcher interface {
	FetchContext(dev string) (*model.DiskMetadata, error)
}

func NewFetcher() Fetcher {
	return &defaultFetcher{}
}

type defaultFetcher struct{}

func (f *defaultFetcher) FetchContext(dev string) (*model.DiskMetadata, error) {
	return FetchContext(dev)
}
