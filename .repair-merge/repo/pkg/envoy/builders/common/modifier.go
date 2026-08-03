package common

type Modifier[R any] struct {
	resource    *R
	configurers []Configurer[R]
}

func NewModifier[R any](r *R) *Modifier[R] {
	return &Modifier[R]{
		resource: r,
	}
}

func (m *Modifier[R]) Configure(configurer Configurer[R]) *Modifier[R] {
	m.configurers = append(m.configurers, configurer)
	return m
}

func (m *Modifier[R]) Modify() error {
	for _, configurer := range m.configurers {
		if err := configurer(m.resource); err != nil {
			return err
		}
	}
	return nil
}
