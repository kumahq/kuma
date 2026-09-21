package labels_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
)

var _ = Describe("EnforcedReadLabels", func() {
	idleTimeout := meshtimeout_api.Conf{
		IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
	}
	meshWideTimeout := func() core_model.Resource {
		return builders.MeshTimeout().
			WithTargetRef(builders.TargetRefMesh()).
			AddTo(builders.TargetRefMesh(), idleTimeout).
			Build()
	}
	appNamespace := resource_labels.NewNamespace("kuma-demo", false)
	systemNamespace := resource_labels.NewNamespace("kuma-system", true)
	universal := resource_labels.UnsetNamespace
	zoneCP := resource_labels.ControlPlane{Mode: config_core.Zone, Zone: "zone-1"}
	globalCP := resource_labels.ControlPlane{Mode: config_core.Global}
	noCP := resource_labels.ControlPlane{}

	type testCase struct {
		r        core_model.Resource
		ns       resource_labels.Namespace
		isLocal  bool
		cp       resource_labels.ControlPlane
		expected map[string]string
	}

	DescribeTable("should recompute the control-plane-owned labels",
		func(given testCase) {
			Expect(resource_labels.EnforcedReadLabels(resource_labels.StoredResource{
				Descriptor: given.r.Descriptor(),
				Spec:       given.r.GetSpec(),
				Namespace:  given.ns,
				IsLocal:    given.isLocal,
			}, given.cp)).To(Equal(given.expected))
		},
		Entry("policy in an app namespace gets the namespace", testCase{
			r:       meshWideTimeout(),
			ns:      appNamespace,
			isLocal: true,
			cp:      noCP,
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			},
		}),
		Entry("a local policy in the system namespace gets no role on read", testCase{
			r:       meshWideTimeout(),
			ns:      systemNamespace,
			isLocal: true,
			cp:      zoneCP,
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
				mesh_proto.ZoneTag:             "zone-1",
			},
		}),
		Entry("nothing is enforced in the system namespace without a mode", testCase{
			r:        meshWideTimeout(),
			ns:       systemNamespace,
			isLocal:  true,
			cp:       noCP,
			expected: nil,
		}),
		Entry("nothing is enforced on Universal without a mode", testCase{
			r:        meshWideTimeout(),
			ns:       universal,
			isLocal:  false,
			cp:       noCP,
			expected: nil,
		}),
		Entry("a non-policy resource gets the namespace", testCase{
			r:       meshservice_api.NewMeshServiceResource(),
			ns:      appNamespace,
			isLocal: true,
			cp:      noCP,
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			},
		}),
		Entry("local resource on a zone gets origin and zone", testCase{
			r:       meshWideTimeout(),
			ns:      universal,
			isLocal: true,
			cp:      zoneCP,
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
				mesh_proto.ZoneTag:             "zone-1",
			},
		}),
		Entry("local resource in a k8s app namespace on a zone gets everything", testCase{
			r:       meshWideTimeout(),
			ns:      appNamespace,
			isLocal: true,
			cp:      zoneCP,
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag:    "kuma-demo",
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
				mesh_proto.ZoneTag:             "zone-1",
			},
		}),
		Entry("import on a zone gets the global origin and no zone", testCase{
			r:       meshWideTimeout(),
			ns:      universal,
			isLocal: false,
			cp:      zoneCP,
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
			},
		}),
		Entry("local resource on global gets the global origin and no zone", testCase{
			r:       meshWideTimeout(),
			ns:      universal,
			isLocal: true,
			cp:      globalCP,
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
			},
		}),
		Entry("import on global gets the zone origin and no zone", testCase{
			r:       meshWideTimeout(),
			ns:      universal,
			isLocal: false,
			cp:      globalCP,
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			},
		}),
		Entry("zone is not enforced on a type the zone does not provide", testCase{
			r:       builders.Mesh().Build(),
			ns:      universal,
			isLocal: true,
			cp:      zoneCP,
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			},
		}),
		Entry("zone is not enforced when the zone has no name", testCase{
			r:       meshWideTimeout(),
			ns:      universal,
			isLocal: true,
			cp:      resource_labels.ControlPlane{Mode: config_core.Zone},
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			},
		}),
	)

	fromGlobal := map[string]string{mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin)}
	fromZone := map[string]string{mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin)}

	DescribeTable("NewStoredResource should derive whether the resource is local",
		func(ns resource_labels.Namespace, stored map[string]string, cp resource_labels.ControlPlane, expected bool) {
			r := resource_labels.NewStoredResource(meshWideTimeout(), ns, stored, cp)

			Expect(r.IsLocal).To(Equal(expected))
			Expect(r.Descriptor.Name).To(Equal(meshtimeout_api.MeshTimeoutType))
			Expect(r.Spec).To(Equal(meshWideTimeout().GetSpec()))
			Expect(r.Namespace).To(Equal(ns))
		},
		Entry("an app namespace is local whatever the stored origin says", appNamespace, fromGlobal, zoneCP, true),
		Entry("the system namespace follows the stored origin: import", systemNamespace, fromGlobal, zoneCP, false),
		Entry("the system namespace follows the stored origin: local", systemNamespace, fromZone, zoneCP, true),
		Entry("Universal follows the stored origin: import", universal, fromZone, globalCP, false),
		Entry("Universal follows the stored origin: local", universal, fromGlobal, globalCP, true),
		Entry("no stored origin is local", universal, nil, zoneCP, true),
		Entry("without a mode everything is local", systemNamespace, fromGlobal, noCP, true),
	)
})
