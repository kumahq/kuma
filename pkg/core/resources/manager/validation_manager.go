package manager

import (
	"context"
	"time"

	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
)

type CustomValidator interface {
	Validate(model.Resource) error
}

type validationManager struct {
	delegate         ResourceManager
	customValidators map[model.ResourceType][]CustomValidator
}

var _ ResourceManager = &validationManager{}

// NewCustomizableValidationManager creates a customizable validation manager
func NewValidationManager(delegate ResourceManager, customValidators map[model.ResourceType][]CustomValidator) ResourceManager {
	if customValidators == nil {
		customValidators = map[model.ResourceType][]CustomValidator{}
	}
	return &validationManager{
		delegate:         delegate,
		customValidators: customValidators,
	}
}

func (r *validationManager) Get(ctx context.Context, resource model.Resource, fs ...store.GetOptionsFunc) error {
	return r.delegate.Get(ctx, resource, fs...)
}

func (r *validationManager) List(ctx context.Context, list model.ResourceList, fs ...store.ListOptionsFunc) error {
	return r.delegate.List(ctx, list, fs...)
}

func (r *validationManager) Create(ctx context.Context, resource model.Resource, fs ...store.CreateOptionsFunc) error {
	if err := model.Validate(resource); err != nil {
		return err
	}
	if validators, found := r.customValidators[resource.Descriptor().Name]; found {
		for _, validator := range validators {
			if err := validator.Validate(resource); err != nil {
				return err
			}
		}
	}
	return r.delegate.Create(ctx, resource, fs...)
}

func (r *validationManager) Delete(ctx context.Context, resource model.Resource, fs ...store.DeleteOptionsFunc) error {
	return r.delegate.Delete(ctx, resource, fs...)
}

func (r *validationManager) DeleteAll(ctx context.Context, list model.ResourceList, fs ...store.DeleteAllOptionsFunc) error {
	return r.delegate.DeleteAll(ctx, list, fs...)
}

func (r *validationManager) Update(ctx context.Context, resource model.Resource, fs ...store.UpdateOptionsFunc) error {
	if err := model.Validate(resource); err != nil {
		return err
	}
	if validators, found := r.customValidators[resource.Descriptor().Name]; found {
		for _, validator := range validators {
			if err := validator.Validate(resource); err != nil {
				return err
			}
		}
	}
	return r.delegate.Update(ctx, resource, append(fs, store.ModifiedAt(time.Now()))...)
}
