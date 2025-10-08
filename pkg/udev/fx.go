package udev

type Watcher interface {
	WatchDisks(ch chan<- string)
}

// NewWatcher returns a default implementation of Watcher
func NewWatcher() Watcher {
	return &defaultWatcher{}
}

// NewWatcherFX fournit un Watcher FX avec configuration par défaut (block/vdb)
func NewWatcherFX() Watcher {
	cfg := DefaultConfiguration()
	return NewController(cfg)
}

type defaultWatcher struct{}

func (w *defaultWatcher) WatchDisks(ch chan<- string) {
	WatchDisks(ch)
}
