package metadata

import (
	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
	"go.uber.org/fx"
)

// WithFile returns a Fetcher for local file-based metadata
func WithFile(path string) fx.Option {
	type provideIn struct {
		fx.In
		Logger capability.Logger
	}

	type provideOut struct {
		fx.Out
		Fetcher model.DiskFetcher
	}

	provide := func(in provideIn) provideOut {
		return provideOut{
			Fetcher: NewFileFetcher(path, in.Logger),
		}
	}

	return fx.Provide(provide)
}

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

	provide := func(in provideIn) provideOut {
		return provideOut{
			Fetcher: NewOpenStackFetcher(in.Logger),
		}
	}

	return fx.Provide(provide)
}
