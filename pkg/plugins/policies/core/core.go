package core

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kumahq/kuma/v2/pkg/core/plugins"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
)

var AllSchemes []func(*runtime.Scheme) error

func Register(resType model.ResourceTypeDescriptor, fn func(scheme *runtime.Scheme) error, p plugins.Plugin) {
	plugins.Register(plugins.PluginName(resType.KumactlArg), p)
	registry.RegisterType(resType)
	AllSchemes = append(AllSchemes, fn)
}
