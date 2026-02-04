package file

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

	provide := func(in provideIn) (out provideOut) {
		out.Fetcher = NewFileFetcher(path, in.Logger)
		return
	}

	return fx.Provide(provide)
}
