package client

type CreateOptions struct {
	Namespace string
	Name      string
}

type CreateOptionsFunc func(*CreateOptions)

func NewCreateOptions(fs ...CreateOptionsFunc) *CreateOptions {
	opts := &CreateOptions{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

func CreateByName(ns, name string) CreateOptionsFunc {
	return func(opts *CreateOptions) {
		opts.Namespace = ns
		opts.Name = name
	}
}

type UpdateOptions struct {
}

type UpdateOptionsFunc func(*UpdateOptions)

func NewUpdateOptions(fs ...UpdateOptionsFunc) *UpdateOptions {
	opts := &UpdateOptions{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

type DeleteOptions struct {
	Namespace string
	Name      string
}

type DeleteOptionsFunc func(*DeleteOptions)

func NewDeleteOptions(fs ...DeleteOptionsFunc) *DeleteOptions {
	opts := &DeleteOptions{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

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

func NewGetOptions(fs ...GetOptionsFunc) *GetOptions {
	opts := &GetOptions{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

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

func NewListOptions(fs ...ListOptionsFunc) *ListOptions {
	opts := &ListOptions{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

func ListByNamespace(ns string) ListOptionsFunc {
	return func(opts *ListOptions) {
		opts.Namespace = ns
	}
}
