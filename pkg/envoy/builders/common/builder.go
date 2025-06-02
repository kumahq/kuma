package common

type Builder[R any] struct {
	configurers []Configurer[R]
}

type Configurer[R any] func(*R) error

func (b *Builder[R]) Configure(configurer Configurer[R]) *Builder[R] {
	b.configurers = append(b.configurers, configurer)
	return b
}

func (b *Builder[R]) ConfigureIf(condition bool, fn func() Configurer[R]) *Builder[R] {
	if !condition {
		return b
	}
	b.configurers = append(b.configurers, fn())
	return b
}

func (b *Builder[R]) Build() (*R, error) {
	var r R
	for _, configurer := range b.configurers {
		if err := configurer(&r); err != nil {
			return nil, err
		}
	}
	return &r, nil
}
