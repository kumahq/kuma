package model

import (
	"fmt"

	"github.com/pkg/errors"
)

type Spec interface {
	Descriptor() ResourceTypeDescriptor
}

type Res[T Spec] struct {
	Meta ResourceMeta
	Spec T

	ValidateFn     func(*Res[T]) error
	DeprecationsFn func(*Res[T]) []string
}

func (ri *Res[T]) GetMeta() ResourceMeta {
	return ri.Meta
}

func (ri *Res[T]) GetSpec() ResourceSpec {
	return ri.Spec
}

func (ri *Res[T]) GetStatus() ResourceStatus {
	return nil
}

func (ri *Res[T]) SetMeta(m ResourceMeta) {
	ri.Meta = m
}

func (ri *Res[T]) SetSpec(spec ResourceSpec) error {
	_, ok := spec.(T)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	}
	ri.Spec = spec.(T)
	return nil
}

func (ri *Res[T]) SetStatus(ResourceStatus) error {
	return errors.New("status not supported")
}

func (ri *Res[T]) Descriptor() ResourceTypeDescriptor {
	return ri.Spec.Descriptor()
}

func (ri *Res[T]) Validate() error {
	if ri.ValidateFn != nil {
		return ri.ValidateFn(ri)
	}
	return nil
}

func (ri *Res[T]) Deprecations() []string {
	if ri.DeprecationsFn != nil {
		return ri.DeprecationsFn(ri)
	}
	return nil
}

type ResStatus[T Spec, S ResourceStatus] struct {
	Meta   ResourceMeta
	Spec   T
	Status S

	ValidateFn     func(*ResStatus[T, S]) error
	DeprecationsFn func(*ResStatus[T, S]) []string
}

func (rsi *ResStatus[T, S]) GetMeta() ResourceMeta {
	return rsi.Meta
}

func (rsi *ResStatus[T, S]) GetSpec() ResourceSpec {
	return rsi.Spec
}

func (rsi *ResStatus[T, S]) GetStatus() ResourceStatus {
	return rsi.Status
}

func (rsi *ResStatus[T, S]) SetMeta(m ResourceMeta) {
	rsi.Meta = m
}

func (rsi *ResStatus[T, S]) SetSpec(spec ResourceSpec) error {
	_, ok := spec.(T)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	}
	rsi.Spec = spec.(T)
	return nil
}

func (rsi *ResStatus[T, S]) SetStatus(status ResourceStatus) error {
	_, ok := status.(S)
	if !ok {
		return fmt.Errorf("invalid type %T for Status", status)
	}
	rsi.Status = status.(S)
	return nil
}

func (rsi *ResStatus[T, S]) Descriptor() ResourceTypeDescriptor {
	return rsi.Spec.Descriptor()
}

func (rsi *ResStatus[T, S]) Validate() error {
	if rsi.ValidateFn != nil {
		return rsi.ValidateFn(rsi)
	}
	return nil
}

func (rsi *ResStatus[T, S]) Deprecations() []string {
	if rsi.DeprecationsFn != nil {
		return rsi.DeprecationsFn(rsi)
	}
	return nil
}

type ResList[T Resource] struct {
	Items      []T
	Pagination Pagination
	NewFn      func() T
}

func (rl *ResList[T]) AddItem(r Resource) error {
	if trr, ok := r.(T); ok {
		rl.Items = append(rl.Items, trr)
		return nil
	} else {
		var zero T
		return ErrorInvalidItemType(zero, r)
	}
}

func (rl *ResList[T]) GetItemType() ResourceType {
	var zero T
	return zero.Descriptor().Name
}

func (rl *ResList[T]) GetItems() []Resource {
	res := make([]Resource, len(rl.Items))
	for i, elem := range rl.Items {
		res[i] = elem
	}
	return res
}

func (rl *ResList[T]) GetPagination() *Pagination {
	return &rl.Pagination
}

func (rl *ResList[T]) NewItem() Resource {
	return rl.NewFn()
}

func (rl *ResList[T]) SetPagination(pagination Pagination) {
	rl.Pagination = pagination
}
