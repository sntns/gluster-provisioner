package capability

type Loader interface {
	Load(name string, data interface{}) error
}
