package xds

import (
	"reflect"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

func ValidateMatchedPoliciesType(t reflect.Type) error {
	// isValidKeyType checks the type is mesh_proto.InboundInterface, mesh_proto.OutboundInterface
	// or string
	isValidKeyType := func(t reflect.Type) error {
		switch t.Name() {
		case reflect.TypeOf(mesh_proto.InboundInterface{}).Name():
		case reflect.TypeOf(mesh_proto.OutboundInterface{}).Name():
		case reflect.String.String():
		default:
			return errors.Errorf("key has wrong type %s", t.Name())
		}
		return nil
	}

	// isValidValueType checks if the type implements core_model.Resource interface or
	// is a slice of elements that implement core_model.Resource interface
	isValidValueType := func(t reflect.Type) error {
		var valueType reflect.Type
		if t.Kind() == reflect.Slice {
			valueType = t.Elem()
		} else {
			valueType = t
		}

		interfaceType := reflect.TypeOf((*core_model.Resource)(nil)).Elem()
		if !valueType.Implements(interfaceType) {
			return errors.Errorf("value doesn't implement %s", interfaceType.Name())
		}
		return nil
	}

	isValidMap := func(t reflect.Type) error {
		var errs error
		if err := isValidKeyType(t.Key()); err != nil {
			errs = multierr.Append(errs, err)
		}
		if err := isValidValueType(t.Elem()); err != nil {
			errs = multierr.Append(errs, err)
		}
		return errs
	}

	isValidField := func(t reflect.Type) error {
		switch t.Kind() {
		case reflect.Map:
			return isValidMap(t)
		case reflect.Slice, reflect.Ptr:
			return isValidValueType(t)
		}
		return errors.Errorf("wrong kind %s", t.Kind())
	}

	var errs error
	for i := 0; i < t.NumField(); i++ {
		errs = multierr.Append(errs, errors.Wrapf(isValidField(t.Field(i).Type), "field %s", t.Field(i).Name))
	}
	return errs
}
