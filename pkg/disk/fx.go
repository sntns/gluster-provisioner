package disk

type Preparer interface {
	PrepareAndMount(dev, name string) error
}

func NewPreparer() Preparer {
	return &defaultPreparer{}
}

type defaultPreparer struct{}

func (p *defaultPreparer) PrepareAndMount(dev, name string) error {
	return PrepareAndMount(dev, name)
}
