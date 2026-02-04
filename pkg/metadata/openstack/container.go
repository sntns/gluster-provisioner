package openstack

import (
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
	"go.uber.org/fx"
)

// WithOpenstack returns a Fetcher for OpenStack metadata
func WithOpenstack() fx.Option {
	type provideIn struct {
		fx.In
		Logger capability.Logger
	}

	type provideOut struct {
		fx.Out
		Fetcher model.DiskFetcher
	}

	provide := func(in provideIn) (out provideOut) {
		out.Fetcher = NewOpenStackFetcher(in.Logger)
		return
	}

	return fx.Provide(provide)
}
