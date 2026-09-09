package k8s

import (
	"fmt"

	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	k8s_common "github.com/kumahq/kuma/v3/pkg/plugins/common/k8s"
	k8s_model "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
)

var _ k8s_common.Converter = &SimpleConverter{}

type SimpleConverter struct {
	KubeFactory     KubeFactory
	SystemNamespace string
	Mode            config_core.CpMode
	Zone            string
}

// ConverterOption configures the parts of a converter that only the control plane
// bootstrap knows, so the constructors stay usable from the webhook and secret
// paths that have no zone of their own.
type ConverterOption func(*SimpleConverter)

// WithLocalZone lets the converter enforce kuma.io/zone on read, see
// labels.EnforcedZoneLabel.
func WithLocalZone(mode config_core.CpMode, zone string) ConverterOption {
	return func(c *SimpleConverter) {
		c.Mode = mode
		c.Zone = zone
	}
}

func NewSimpleConverter(systemNamespace string, opts ...ConverterOption) k8s_common.Converter {
	c := &SimpleConverter{
		KubeFactory:     NewSimpleKubeFactory(),
		SystemNamespace: systemNamespace,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func NewSimpleKubeFactory() KubeFactory {
	return &SimpleKubeFactory{
		KubeTypes: registry.Global(),
	}
}

func (c *SimpleConverter) ToKubernetesObject(r core_model.Resource) (k8s_model.KubernetesObject, error) {
	obj, err := c.KubeFactory.NewObject(r)
	if err != nil {
		return nil, err
	}
	obj.SetSpec(r.GetSpec())
	if r.Descriptor().HasStatus {
		if err := obj.SetStatus(r.GetStatus()); err != nil {
			return nil, err
		}
	}
	if r.GetMeta() != nil {
		if adapter, ok := r.GetMeta().(*KubernetesMetaAdapter); ok {
			obj.SetMesh(adapter.Mesh)
			obj.SetObjectMeta(&adapter.ObjectMeta)
		} else {
			return nil, fmt.Errorf("meta has unexpected type: %#v", r.GetMeta())
		}
	}
	return obj, nil
}

func (c *SimpleConverter) ToKubernetesList(rl core_model.ResourceList) (k8s_model.KubernetesList, error) {
	return c.KubeFactory.NewList(rl)
}

func (c *SimpleConverter) ToCoreResource(obj k8s_model.KubernetesObject, out core_model.Resource) error {
	spec, err := obj.GetSpec()
	if err != nil {
		return err
	}
	// SetSpec first, then derive labels from out.GetSpec(): a stored object with an
	// omitted spec yields a typed-nil here, and SetSpec normalizes it to an empty
	// spec. Deriving from the raw value would call policy methods on a nil pointer.
	if err := out.SetSpec(spec); err != nil {
		return err
	}
	out.SetMeta(newMetaAdapter(obj, c.SystemNamespace, out.Descriptor(), out.GetSpec(), c.Mode, c.Zone))
	if out.Descriptor().HasStatus {
		status, err := obj.GetStatus()
		if err != nil {
			return err
		}
		if err := out.SetStatus(status); err != nil {
			return err
		}
	}
	return nil
}

func (c *SimpleConverter) ToCoreList(in k8s_model.KubernetesList, out core_model.ResourceList, predicate k8s_common.ConverterPredicate) error {
	for _, o := range in.GetItems() {
		r := out.NewItem()
		if err := c.ToCoreResource(o, r); err != nil {
			return err
		}
		if predicate(r) {
			_ = out.AddItem(r)
		}
	}
	return nil
}
