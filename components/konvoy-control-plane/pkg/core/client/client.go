package client

import (
	"context"
	"fmt"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/model"
)

type ResourceClient interface {
	Create(context.Context, model.Resource, ...CreateOptionsFunc) error
	Update(context.Context, model.Resource, ...UpdateOptionsFunc) error
	Delete(context.Context, model.Resource, ...DeleteOptionsFunc) error
	Get(context.Context, model.Resource, ...GetOptionsFunc) error
	List(context.Context, model.ResourceList, ...ListOptionsFunc) error
}

func NewStrictResourceClient(c ResourceClient) ResourceClient {
	return &strictResourceClient{delegate: c}
}

var _ ResourceClient = &strictResourceClient{}

// strictResourceClient encapsulates a contract between ResourceClient and its users.
type strictResourceClient struct {
	delegate ResourceClient
}

func (s *strictResourceClient) Create(ctx context.Context, r model.Resource, fs ...CreateOptionsFunc) error {
	if r == nil {
		return fmt.Errorf("ResourceClient.Create() requires a non-nil resource")
	}
	if r.GetMeta() != nil {
		return fmt.Errorf("ResourceClient.Create() ignores resource.GetMeta() but the argument has a non-nil value")
	}
	opts := NewCreateOptions(fs...)
	if opts.Name == "" {
		return fmt.Errorf("ResourceClient.Create() requires options.Name to be a non-empty value")
	}
	return s.delegate.Create(ctx, r, fs...)
}
func (s *strictResourceClient) Update(ctx context.Context, r model.Resource, fs ...UpdateOptionsFunc) error {
	if r == nil {
		return fmt.Errorf("ResourceClient.Update() requires a non-nil resource")
	}
	if r.GetMeta() == nil {
		return fmt.Errorf("ResourceClient.Update() requires resource.GetMeta() to be a non-nil value previously returned by ResourceClient.Get()")
	}
	return s.delegate.Update(ctx, r, fs...)
}
func (s *strictResourceClient) Delete(ctx context.Context, r model.Resource, fs ...DeleteOptionsFunc) error {
	if r == nil {
		return fmt.Errorf("ResourceClient.Delete() requires a non-nil resource")
	}
	opts := NewDeleteOptions(fs...)
	if opts.Name == "" {
		return fmt.Errorf("ResourceClient.Delete() requires options.Name to be a non-empty value")
	}
	if r.GetMeta() != nil {
		if opts.Name != r.GetMeta().GetName() {
			return fmt.Errorf("ResourceClient.Delete() requires resource.GetMeta() either to be a nil or resource.GetMeta().GetName() == options.Name")
		}
		if opts.Namespace != r.GetMeta().GetNamespace() {
			return fmt.Errorf("ResourceClient.Delete() requires resource.GetMeta() either to be a nil or resource.GetMeta().GetNamespace() == options.Namespace")
		}
	}
	return s.delegate.Delete(ctx, r, fs...)
}
func (s *strictResourceClient) Get(ctx context.Context, r model.Resource, fs ...GetOptionsFunc) error {
	if r == nil {
		return fmt.Errorf("ResourceClient.Get() requires a non-nil resource")
	}
	if r.GetMeta() != nil {
		return fmt.Errorf("ResourceClient.Get() ignores resource.GetMeta() but the argument has a non-nil value")
	}
	opts := NewGetOptions(fs...)
	if opts.Name == "" {
		return fmt.Errorf("ResourceClient.Get() requires options.Name to be a non-empty value")
	}
	return s.delegate.Get(ctx, r, fs...)
}
func (s *strictResourceClient) List(ctx context.Context, rs model.ResourceList, fs ...ListOptionsFunc) error {
	if rs == nil {
		return fmt.Errorf("ResourceClient.List() requires a non-nil resource list")
	}
	return s.delegate.List(ctx, rs, fs...)
}

func ErrorResourceNotFound(rt model.ResourceType, namespace, name string) error {
	return fmt.Errorf("Resource not found: type=%q namespace=%q name=%q", rt, namespace, name)
}

func ErrorResourceAlreadyExists(rt model.ResourceType, namespace, name string) error {
	return fmt.Errorf("Resource already exists: type=%q namespace=%q name=%q", rt, namespace, name)
}
