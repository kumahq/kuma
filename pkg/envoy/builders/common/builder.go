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

func If[R any](condition bool, configurer Configurer[R]) Configurer[R] {
	return func(c *R) error {
		if condition {
			return configurer(c)
		}
		return nil
	}
}

func IfNotNil[R, P any](ptr *P, fn func(P) Configurer[R]) Configurer[R] {
	return func(c *R) error {
		if ptr != nil {
			return fn(*ptr)(c)
		}
		return nil
	}
}

func AsBuilder[R any](fn func() (*R, error)) *Builder[R] {
	return &Builder[R]{configurers: []Configurer[R]{func(r *R) error {
		ptr, err := fn()
		if err != nil {
			return err
		}
		*r = *ptr
		return nil
	}}}
}
