package client

type CreateOptions struct {
	Namespace string
	Name      string
}

type CreateOptionsFunc func(*CreateOptions)

func CreateByName(ns, name string) CreateOptionsFunc {
	return func(opts *CreateOptions) {
		opts.Namespace = ns
		opts.Name = name
	}
}

type UpdateOptions struct {
}

type UpdateOptionsFunc func(*UpdateOptions)

type DeleteOptions struct {
	Namespace string
	Name      string
}

type DeleteOptionsFunc func(*DeleteOptions)

func DeleteByName(ns, name string) DeleteOptionsFunc {
	return func(opts *DeleteOptions) {
		opts.Namespace = ns
		opts.Name = name
	}
}

type GetOptions struct {
	Namespace string
	Name      string
}

type GetOptionsFunc func(*GetOptions)

func GetByName(ns, name string) GetOptionsFunc {
	return func(opts *GetOptions) {
		opts.Namespace = ns
		opts.Name = name
	}
}

type ListOptions struct {
	Namespace string
}

type ListOptionsFunc func(*ListOptions)

func ListByNamespace(ns string) ListOptionsFunc {
	return func(opts *ListOptions) {
		opts.Namespace = ns
	}
}
