package util

import (
	"strings"
	"time"

	"golang.org/x/exp/maps"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_store "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

// KDS ResourceMeta only contains name and mesh.
// The rest is managed by the receiver of resources anyways. See ResourceSyncer#Sync
type resourceMeta struct {
	name           string
	mesh           string
	labels         map[string]string
	nameExtensions model.ResourceNameExtensions
}

type CloneResourceMetaOpt func(*resourceMeta)

func WithName(name string) CloneResourceMetaOpt {
	return func(m *resourceMeta) {
		if m.labels[mesh_proto.DisplayName] == "" {
			m.labels[mesh_proto.DisplayName] = m.name
		}
		m.name = name
	}
}

func WithLabel(key, value string) CloneResourceMetaOpt {
	return func(m *resourceMeta) {
		m.labels[key] = value
	}
}

// PopulateNamespaceLabelFromNameExtension on Kubernetes zones adds 'k8s.kuma.io/namespace' label to the resources
// before syncing them to Global.
//
// In 2.7.x method 'GetMeta().GetLabels()' on Kubernetes returned a label map with 'k8s.kuma.io/namespace' added
// dynamically. This behavior was changed in 2.9.x by https://github.com/kumahq/kuma/pull/11020, the namespace label is now
// supposed to be set in ComputeLabels function. But this functions is called only on Create/Update of the resources.
// This means policies that were created on 2.7.x won't have 'k8s.kuma.io/namespace' label when synced to Global.
// Even though the lack of namespace labels affects only how resource looks in GUI on Global it's still worth setting it.
func PopulateNamespaceLabelFromNameExtension() CloneResourceMetaOpt {
	return func(m *resourceMeta) {
		namespace := m.nameExtensions[model.K8sNamespaceComponent]
		if _, ok := m.labels[mesh_proto.KubeNamespaceTag]; !ok && namespace != "" {
			m.labels[mesh_proto.KubeNamespaceTag] = namespace
		}
	}
}

func WithoutLabel(key string) CloneResourceMetaOpt {
	return func(m *resourceMeta) {
		delete(m.labels, key)
	}
}

func WithoutLabelPrefixes(prefixes ...string) CloneResourceMetaOpt {
	return func(m *resourceMeta) {
		for label := range m.labels {
			for _, prefix := range prefixes {
				if strings.HasPrefix(label, prefix) {
					delete(m.labels, label)
				}
			}
		}
	}
}

func If(condition func(resource model.ResourceMeta) bool, fn CloneResourceMetaOpt) CloneResourceMetaOpt {
	return func(meta *resourceMeta) {
		if condition(meta) {
			fn(meta)
		}
	}
}

func IsKubernetes(storeType config_store.StoreType) func(model.ResourceMeta) bool {
	return func(_ model.ResourceMeta) bool {
		return storeType == config_store.KubernetesStore
	}
}

func CloneResourceMeta(m model.ResourceMeta, fs ...CloneResourceMetaOpt) model.ResourceMeta {
	labels := maps.Clone(m.GetLabels())
	if labels == nil {
		labels = map[string]string{}
	}
	ne := maps.Clone(m.GetNameExtensions())
	if ne == nil {
		ne = model.ResourceNameExtensions{}
	}
	meta := &resourceMeta{
		name:           m.GetName(),
		mesh:           m.GetMesh(),
		labels:         labels,
		nameExtensions: ne,
	}
	for _, f := range fs {
		f(meta)
	}
	if len(meta.labels) == 0 {
		meta.labels = nil
	}
	return meta
}

func kumaResourceMetaToResourceMeta(meta *mesh_proto.KumaResource_Meta) model.ResourceMeta {
	return &resourceMeta{
		name:   meta.Name,
		mesh:   meta.Mesh,
		labels: meta.GetLabels(),
	}
}

func (r *resourceMeta) GetName() string {
	return r.name
}

func (r *resourceMeta) GetNameExtensions() model.ResourceNameExtensions {
	return model.ResourceNameExtensionsUnsupported
}

func (r *resourceMeta) GetVersion() string {
	return ""
}

func (r *resourceMeta) GetMesh() string {
	return r.mesh
}

func (r *resourceMeta) GetCreationTime() time.Time {
	return time.Unix(0, 0)
}

func (r *resourceMeta) GetModificationTime() time.Time {
	return time.Unix(0, 0)
}

func (r *resourceMeta) GetLabels() map[string]string {
	return r.labels
}
