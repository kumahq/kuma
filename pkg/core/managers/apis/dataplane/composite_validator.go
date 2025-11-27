package dataplane

import (
	"context"

	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

type ValidatorPredicate func() bool

type conditionalValidator struct {
	validator Validator
	predicate ValidatorPredicate
}

type compositeValidator struct {
	validators []conditionalValidator
}

var _ Validator = &compositeValidator{}

func NewCompositeValidator(validators ...conditionalValidator) Validator {
	return &compositeValidator{
		validators: validators,
	}
}

// Always returns a conditional validator that always runs
func Always(v Validator) conditionalValidator {
	return conditionalValidator{
		validator: v,
		predicate: func() bool { return true },
	}
}

// When returns a conditional validator that runs only when predicate returns true
func When(predicate ValidatorPredicate, v Validator) conditionalValidator {
	return conditionalValidator{
		validator: v,
		predicate: predicate,
	}
}

func (c *compositeValidator) ValidateCreate(ctx context.Context, key model.ResourceKey, newDp *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) error {
	for _, cv := range c.validators {
		if cv.predicate() {
			if err := cv.validator.ValidateCreate(ctx, key, newDp, mesh); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *compositeValidator) ValidateUpdate(ctx context.Context, newDp *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) error {
	for _, cv := range c.validators {
		if cv.predicate() {
			if err := cv.validator.ValidateUpdate(ctx, newDp, mesh); err != nil {
				return err
			}
		}
	}
	return nil
}
