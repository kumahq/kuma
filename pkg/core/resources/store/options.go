package store

import (
	"github.com/Kong/kuma/pkg/core/resources/model"
)

type CreateOptions struct {
	Namespace string
	Name      string
	Mesh      string
}

type CreateOptionsFunc func(*CreateOptions)

func NewCreateOptions(fs ...CreateOptionsFunc) *CreateOptions {
	opts := &CreateOptions{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

func CreateBy(key model.ResourceKey) CreateOptionsFunc {
	return CreateByKey(key.Namespace, key.Name, key.Mesh)
}

func CreateByKey(ns, name, mesh string) CreateOptionsFunc {
	return func(opts *CreateOptions) {
		opts.Namespace = ns
		opts.Name = name
		opts.Mesh = mesh
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
	Mesh      string
}

type DeleteOptionsFunc func(*DeleteOptions)

func NewDeleteOptions(fs ...DeleteOptionsFunc) *DeleteOptions {
	opts := &DeleteOptions{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

func DeleteBy(key model.ResourceKey) DeleteOptionsFunc {
	return DeleteByKey(key.Namespace, key.Name, key.Mesh)
}

func DeleteByKey(ns, name, mesh string) DeleteOptionsFunc {
	return func(opts *DeleteOptions) {
		opts.Namespace = ns
		opts.Name = name
		opts.Mesh = mesh
	}
}

type DeleteAllOptions struct {
	Mesh string
}

type DeleteAllOptionsFunc func(*DeleteAllOptions)

func DeleteAllByMesh(mesh string) DeleteAllOptionsFunc {
	return func(opts *DeleteAllOptions) {
		opts.Mesh = mesh
	}
}

func NewDeleteAllOptions(fs ...DeleteAllOptionsFunc) *DeleteAllOptions {
	opts := &DeleteAllOptions{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

type GetOptions struct {
	Namespace string
	Name      string
	Mesh      string
}

type GetOptionsFunc func(*GetOptions)

func NewGetOptions(fs ...GetOptionsFunc) *GetOptions {
	opts := &GetOptions{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

func GetBy(key model.ResourceKey) GetOptionsFunc {
	return GetByKey(key.Namespace, key.Name, key.Mesh)
}

func GetByKey(ns, name, mesh string) GetOptionsFunc {
	return func(opts *GetOptions) {
		opts.Namespace = ns
		opts.Name = name
		opts.Mesh = mesh
	}
}

type ListOptions struct {
	Namespace string
	Mesh      string
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

func ListByMesh(mesh string) ListOptionsFunc {
	return func(opts *ListOptions) {
		opts.Mesh = mesh
	}
}
