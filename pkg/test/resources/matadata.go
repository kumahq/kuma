package resources

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/kds/hash"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/k8s"
	"maps"
)

type BuildMeta func(name, mesh string, labels map[string]string) core_model.ResourceMeta

func GlobalUni(name, mesh string, labels map[string]string) core_model.ResourceMeta {
	globalLabels := map[string]string{
		"kuma.io/origin":       "global",
		"kuma.io/display-name": name,
	}
	maps.Copy(globalLabels, labels)
	return &model.ResourceMeta{
		Name:   name,
		Mesh:   mesh,
		Labels: globalLabels,
	}
}

func GlobalK8s(name, mesh string, labels map[string]string) core_model.ResourceMeta {
	globalLabels := map[string]string{
		"kuma.io/origin":        "global",
		"k8s.kuma.io/namespace": "ns-k8s",
		"kuma.io/mesh":          mesh,
		"kuma.io/display-name":  name,
	}
	maps.Copy(globalLabels, labels)
	return &model.ResourceMeta{
		Name:   k8s.K8sNamespacedNameToCoreName(name, "ns-k8s"),
		Mesh:   mesh,
		Labels: globalLabels,
		NameExtensions: map[string]string{
			"k8s.kuma.io/namespace": "ns-k8s",
			"k8s.kuma.io/name":      name,
		},
	}
}

func ZoneUni(name, mesh string, labels map[string]string) core_model.ResourceMeta {
	zoneLabels := map[string]string{
		"kuma.io/origin":       "zone",
		"kuma.io/zone":         "zone-uni",
		"kuma.io/display-name": name,
	}
	maps.Copy(zoneLabels, labels)
	return &model.ResourceMeta{
		Name:   name,
		Mesh:   mesh,
		Labels: zoneLabels,
	}
}

func ZoneK8s(name, mesh string, labels map[string]string) core_model.ResourceMeta {
	zoneLabels := map[string]string{
		"kuma.io/origin":        "zone",
		"kuma.io/zone":          "zone-k8s",
		"k8s.kuma.io/namespace": "ns-k8s",
		"kuma.io/mesh":          mesh,
		"kuma.io/display-name":  name,
	}
	maps.Copy(zoneLabels, labels)
	return &model.ResourceMeta{
		Name:   k8s.K8sNamespacedNameToCoreName(name, "ns-k8s"),
		Mesh:   mesh,
		Labels: zoneLabels,
		NameExtensions: map[string]string{
			"k8s.kuma.io/namespace": "ns-k8s",
			"k8s.kuma.io/name":      name,
		},
	}
}

func SystemPolicy(fn BuildMeta) BuildMeta {
	return WithNamespace(WithPolicyRole(fn, mesh_proto.SystemPolicyRole), "kuma-system")
}

func ProducerPolicy(fn BuildMeta) BuildMeta {
	return WithPolicyRole(fn, mesh_proto.ProducerPolicyRole)
}

func WithPolicyRole(fn BuildMeta, policyRole mesh_proto.PolicyRole) BuildMeta {
	return func(name, mesh string, labels map[string]string) core_model.ResourceMeta {
		meta := fn(name, mesh, labels)
		meta.GetLabels()[mesh_proto.PolicyRoleLabel] = string(policyRole)
		return meta
	}
}

func WithNamespace(fn BuildMeta, namespace string) BuildMeta {
	return func(name, mesh string, labels map[string]string) core_model.ResourceMeta {
		meta := fn(name, mesh, labels)
		meta.GetLabels()[mesh_proto.KubeNamespaceTag] = namespace
		return meta
	}
}

func SyncToUni(fn BuildMeta) BuildMeta {
	return func(name, mesh string, labels map[string]string) core_model.ResourceMeta {
		m := fn(name, mesh, labels)
		var values []string
		if v, ok := m.GetLabels()[mesh_proto.ZoneTag]; ok {
			values = append(values, v)
		}
		if v, ok := m.GetLabels()[mesh_proto.KubeNamespaceTag]; ok {
			values = append(values, v)
		}
		return &model.ResourceMeta{
			Name:   hash.HashedName(m.GetMesh(), core_model.GetDisplayName(m), values...),
			Mesh:   m.GetMesh(),
			Labels: m.GetLabels(),
		}
	}
}

func SyncToK8s(fn BuildMeta) BuildMeta {
	return func(name, mesh string, labels map[string]string) core_model.ResourceMeta {
		m := fn(name, mesh, labels)
		var values []string
		if v, ok := m.GetLabels()[mesh_proto.ZoneTag]; ok {
			values = append(values, v)
		}
		if v, ok := m.GetLabels()[mesh_proto.KubeNamespaceTag]; ok {
			values = append(values, v)
		}
		newName := hash.HashedName(m.GetMesh(), core_model.GetDisplayName(m), values...)
		return &model.ResourceMeta{
			Name:   k8s.K8sNamespacedNameToCoreName(newName, "kuma-system"),
			Mesh:   m.GetMesh(),
			Labels: m.GetLabels(),
			NameExtensions: map[string]string{
				"k8s.kuma.io/namespace": "kuma-system",
				"k8s.kuma.io/name":      newName,
			},
		}
	}
}

func UpdateResourcesMeta(fn BuildMeta, rs core_model.ResourceList) {
	for _, r := range rs.GetItems() {
		if r.Descriptor().Name == mesh.MeshType {
			continue
		}
		UpdateResourceMeta(fn, r)
	}
}

func UpdateResourceMeta(fn BuildMeta, r core_model.Resource) {
	r.SetMeta(fn(r.GetMeta().GetName(), r.GetMeta().GetMesh(), r.GetMeta().GetLabels()))
}
