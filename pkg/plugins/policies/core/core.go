package core

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

var AllSchemes []func(*runtime.Scheme) error

func Register(resType model.ResourceTypeDescriptor, fn func(scheme *runtime.Scheme) error) {
	registry.RegisterType(resType)
	AllSchemes = append(AllSchemes, fn)
}
