package labels_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

var _ = Describe("Validate", func() {
	timeoutConf := meshtimeout_api.Conf{IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second}}
	timeout := func() core_model.Resource {
		return builders.MeshTimeout().
			WithMesh("mesh-1").
			WithTargetRef(builders.TargetRefMesh()).
			AddTo(builders.TargetRefMesh(), timeoutConf).
			Build()
	}
	timeoutWithoutMesh := func() core_model.Resource {
		return builders.MeshTimeout().
			WithMesh("").
			WithTargetRef(builders.TargetRefMesh()).
			AddTo(builders.TargetRefMesh(), timeoutConf).
			Build()
	}
	mixedTimeout := func() core_model.Resource {
		return builders.MeshTimeout().
			WithMesh("mesh-1").
			WithTargetRef(builders.TargetRefMesh()).
			AddTo(builders.TargetRefMeshServiceLabels(map[string]string{
				mesh_proto.DisplayName:      "backend-1",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}, ""), timeoutConf).
			AddTo(builders.TargetRefMeshServiceLabels(map[string]string{
				mesh_proto.DisplayName:      "backend-2",
				mesh_proto.KubeNamespaceTag: "other-ns",
			}, ""), timeoutConf).
			Build()
	}
	dataplane := func() core_model.Resource {
		return builders.Dataplane().WithMesh("mesh-1").WithServices("backend").Build()
	}
	zoneIngressDataplane := func() core_model.Resource {
		return builders.Dataplane().WithMesh("mesh-1").
			With(func(dp *core_mesh.DataplaneResource) {
				dp.Spec.Networking.Listeners = []*mesh_proto.Dataplane_Networking_Listener{{
					Type:    mesh_proto.Dataplane_Networking_Listener_ZoneIngress,
					Address: "127.0.0.1",
					Port:    10001,
				}}
			}).
			Build()
	}
	meshResource := func() core_model.Resource {
		return &core_mesh.MeshResource{
			Meta: &test_model.ResourceMeta{Name: "mesh-1"},
			Spec: &mesh_proto.Mesh{},
		}
	}
	appNamespace := resource_labels.NewNamespace("kuma-demo", false)
	systemNamespace := resource_labels.NewNamespace("kuma-system", true)

	universalGlobal := resource_labels.ControlPlane{Mode: config_core.Global}
	universalFederated := resource_labels.ControlPlane{Mode: config_core.Zone, Zone: "zone-1", FederatedZone: true}
	universalNonFederated := resource_labels.ControlPlane{Mode: config_core.Zone, Zone: "zone-1"}
	k8sGlobal := resource_labels.ControlPlane{Mode: config_core.Global, IsK8s: true}
	k8sFederated := resource_labels.ControlPlane{Mode: config_core.Zone, Zone: "zone-1", IsK8s: true, FederatedZone: true}
	k8sNonFederated := resource_labels.ControlPlane{Mode: config_core.Zone, Zone: "zone-1", IsK8s: true}

	type testCase struct {
		r        core_model.Resource
		ns       resource_labels.Namespace
		labels   map[string]string
		previous map[string]string
		trusted  bool
		cp       resource_labels.ControlPlane
		expected []validators.Violation
	}

	violation := func(key, msg string) validators.Violation {
		return validators.Violation{Field: validators.Root().Key(key).String(), Message: msg}
	}
	differs := func(key, got, expected string) validators.Violation {
		return violation(key, `label "`+key+`" is managed by the control plane: got "`+got+`", expected "`+expected+`"`)
	}
	notHere := func(key string) validators.Violation {
		return violation(key, `label "`+key+`" is managed by the control plane and cannot be set here`)
	}

	DescribeTable("should reject a control-plane-owned label unless it carries the value the control plane computes",
		func(given testCase) {
			w := resource_labels.Write{
				Descriptor:    given.r.Descriptor(),
				Spec:          given.r.GetSpec(),
				Namespace:     given.ns,
				Mesh:          given.r.GetMeta().GetMesh(),
				DisplayName:   given.r.GetMeta().GetName(),
				Labels:        given.labels,
				Previous:      given.previous,
				TrustedWriter: given.trusted,
			}
			err := resource_labels.Validate(w, given.cp)
			if given.expected == nil {
				Expect(err.HasViolations()).To(BeFalse(), err.Error())
			} else {
				Expect(err.Violations).To(Equal(given.expected))
			}
		},
		Entry("no labels: nothing to check", testCase{
			r: timeout(), labels: map[string]string{}, cp: universalGlobal,
		}),
		Entry("trusted writer: every mismatch is skipped", testCase{
			r: timeout(), ns: appNamespace, trusted: true, cp: k8sFederated,
			labels: map[string]string{
				mesh_proto.ResourceOriginLabel: "global",
				mesh_proto.ZoneTag:             "zone-2",
				metadata.KumaMeshLabel:         "mesh-2",
				mesh_proto.PolicyRoleLabel:     "invalid",
				mesh_proto.DisplayName:         "other",
				mesh_proto.EnvTag:              "universal",
				mesh_proto.KubeNamespaceTag:    "other-ns",
				metadata.KumaServiceAccount:    "victim-sa",
			},
		}),
		Entry("update: a label unchanged from the stored value is not checked", testCase{
			r: timeout(), ns: appNamespace, cp: k8sFederated,
			previous: map[string]string{mesh_proto.PolicyRoleLabel: "producer", mesh_proto.ZoneTag: "zone-1"},
			labels:   map[string]string{mesh_proto.PolicyRoleLabel: "producer", mesh_proto.ZoneTag: "zone-1"},
		}),
		Entry("update: a label changed from the stored value is checked", testCase{
			r: timeout(), ns: appNamespace, cp: k8sFederated,
			previous: map[string]string{mesh_proto.ZoneTag: "zone-1"},
			labels:   map[string]string{mesh_proto.ZoneTag: "zone-2"},
			expected: []validators.Violation{differs(mesh_proto.ZoneTag, "zone-2", "zone-1")},
		}),
		Entry("update: a label not stored before is checked", testCase{
			r: timeout(), ns: appNamespace, cp: k8sFederated,
			previous: map[string]string{},
			labels:   map[string]string{mesh_proto.ZoneTag: "zone-2"},
			expected: []validators.Violation{differs(mesh_proto.ZoneTag, "zone-2", "zone-1")},
		}),
		Entry("violations are reported in registry order", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ZoneTag: "zone-2", metadata.KumaMeshLabel: "mesh-2", mesh_proto.ResourceOriginLabel: "global"}, cp: universalFederated,
			expected: []validators.Violation{
				differs(mesh_proto.ResourceOriginLabel, "global", "zone"),
				differs(mesh_proto.ZoneTag, "zone-2", "zone-1"),
				differs(metadata.KumaMeshLabel, "mesh-2", "mesh-1"),
			},
		}),
		// origin
		Entry("origin: equal on a global CP", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: universalGlobal,
		}),
		Entry("origin: differs on a global CP", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: "zone"}, cp: universalGlobal,
			expected: []validators.Violation{differs(mesh_proto.ResourceOriginLabel, "zone", "global")},
		}),
		Entry("origin: differs on a federated zone", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: universalFederated,
			expected: []validators.Violation{differs(mesh_proto.ResourceOriginLabel, "global", "zone")},
		}),
		Entry("origin: differs on a non-federated zone", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: universalNonFederated,
			expected: []validators.Violation{differs(mesh_proto.ResourceOriginLabel, "global", "zone")},
		}),
		Entry("origin: unknown value differs, once", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: "unknownvalue"}, cp: universalGlobal,
			expected: []validators.Violation{differs(mesh_proto.ResourceOriginLabel, "unknownvalue", "global")},
		}),
		Entry("origin: empty differs", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: ""}, cp: universalGlobal,
			expected: []validators.Violation{differs(mesh_proto.ResourceOriginLabel, "", "global")},
		}),
		Entry("origin: differs on a k8s global CP", testCase{
			r: timeout(), ns: systemNamespace, labels: map[string]string{mesh_proto.ResourceOriginLabel: "zone"}, cp: k8sGlobal,
			expected: []validators.Violation{differs(mesh_proto.ResourceOriginLabel, "zone", "global")},
		}),
		Entry("origin: differs in an app namespace of a k8s federated zone", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: k8sFederated,
			expected: []validators.Violation{differs(mesh_proto.ResourceOriginLabel, "global", "zone")},
		}),
		Entry("origin: differs on a k8s non-federated zone", testCase{
			r: timeout(), ns: systemNamespace, labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: k8sNonFederated,
			expected: []validators.Violation{differs(mesh_proto.ResourceOriginLabel, "global", "zone")},
		}),
		Entry("origin: differs on a non-plugin type on k8s", testCase{
			r: dataplane(), ns: appNamespace, labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: k8sFederated,
			expected: []validators.Violation{differs(mesh_proto.ResourceOriginLabel, "global", "zone")},
		}),
		// zone
		Entry("zone: equal", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ZoneTag: "zone-1"}, cp: universalNonFederated,
		}),
		Entry("zone: differs", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ZoneTag: "zone-2"}, cp: universalNonFederated,
			expected: []validators.Violation{differs(mesh_proto.ZoneTag, "zone-2", "zone-1")},
		}),
		Entry("zone: on a global CP", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ZoneTag: "zone-1"}, cp: universalGlobal,
			expected: []validators.Violation{notHere(mesh_proto.ZoneTag)},
		}),
		Entry("zone: on a type the zone does not provide", testCase{
			r: meshResource(), labels: map[string]string{mesh_proto.ZoneTag: "zone-1"}, cp: universalNonFederated,
			expected: []validators.Violation{notHere(mesh_proto.ZoneTag)},
		}),
		Entry("zone: equal on k8s", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.ZoneTag: "zone-1"}, cp: k8sFederated,
		}),
		Entry("zone: differs without an origin on a k8s federated zone", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.ZoneTag: "zone-2"}, cp: k8sFederated,
			expected: []validators.Violation{differs(mesh_proto.ZoneTag, "zone-2", "zone-1")},
		}),
		Entry("zone: differs on a k8s non-federated zone", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.ZoneTag: "zone-2"}, cp: k8sNonFederated,
			expected: []validators.Violation{differs(mesh_proto.ZoneTag, "zone-2", "zone-1")},
		}),
		Entry("zone: on a k8s global CP", testCase{
			r: timeout(), ns: systemNamespace, labels: map[string]string{mesh_proto.ZoneTag: "zone-2"}, cp: k8sGlobal,
			expected: []validators.Violation{notHere(mesh_proto.ZoneTag)},
		}),
		// mesh
		Entry("mesh: equal", testCase{
			r: timeout(), labels: map[string]string{metadata.KumaMeshLabel: "mesh-1"}, cp: universalNonFederated,
		}),
		Entry("mesh: differs", testCase{
			r: timeout(), labels: map[string]string{metadata.KumaMeshLabel: "mesh-2"}, cp: universalNonFederated,
			expected: []validators.Violation{differs(metadata.KumaMeshLabel, "mesh-2", "mesh-1")},
		}),
		Entry("mesh: differs on k8s", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{metadata.KumaMeshLabel: "mesh-2"}, cp: k8sNonFederated,
			expected: []validators.Violation{differs(metadata.KumaMeshLabel, "mesh-2", "mesh-1")},
		}),
		Entry("mesh: the default mesh when the write has none", testCase{
			r: timeoutWithoutMesh(), labels: map[string]string{metadata.KumaMeshLabel: "default"}, cp: universalNonFederated,
		}),
		Entry("mesh: on a global-scoped type", testCase{
			r: meshResource(), labels: map[string]string{metadata.KumaMeshLabel: "mesh-1"}, cp: universalNonFederated,
			expected: []validators.Violation{notHere(metadata.KumaMeshLabel)},
		}),
		// policy-role
		Entry("policy-role: system on Universal", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.PolicyRoleLabel: "system"}, cp: universalNonFederated,
		}),
		Entry("policy-role: differs on Universal", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.PolicyRoleLabel: "consumer"}, cp: universalNonFederated,
			expected: []validators.Violation{differs(mesh_proto.PolicyRoleLabel, "consumer", "system")},
		}),
		Entry("policy-role: empty differs", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.PolicyRoleLabel: ""}, cp: universalNonFederated,
			expected: []validators.Violation{differs(mesh_proto.PolicyRoleLabel, "", "system")},
		}),
		Entry("policy-role: on a non-policy", testCase{
			r: dataplane(), labels: map[string]string{mesh_proto.PolicyRoleLabel: "system"}, cp: universalNonFederated,
			expected: []validators.Violation{notHere(mesh_proto.PolicyRoleLabel)},
		}),
		Entry("policy-role: system in the system namespace on k8s", testCase{
			r: timeout(), ns: systemNamespace, labels: map[string]string{mesh_proto.PolicyRoleLabel: "system"}, cp: k8sFederated,
		}),
		Entry("policy-role: equal in an app namespace on k8s", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.PolicyRoleLabel: "consumer"}, cp: k8sFederated,
		}),
		Entry("policy-role: differs in an app namespace on k8s", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.PolicyRoleLabel: "invalid"}, cp: k8sFederated,
			expected: []validators.Violation{differs(mesh_proto.PolicyRoleLabel, "invalid", "consumer")},
		}),
		Entry("policy-role: a policy Compute rejects is left to Compute", testCase{
			r: mixedTimeout(), ns: appNamespace, labels: map[string]string{mesh_proto.PolicyRoleLabel: "invalid"}, cp: k8sFederated,
		}),
		// display-name
		Entry("display-name: equal", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.DisplayName: "mt-1"}, cp: universalNonFederated,
		}),
		Entry("display-name: differs", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.DisplayName: "other"}, cp: universalNonFederated,
			expected: []validators.Violation{differs(mesh_proto.DisplayName, "other", "mt-1")},
		}),
		Entry("display-name: differs on k8s", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.DisplayName: "other"}, cp: k8sNonFederated,
			expected: []validators.Violation{differs(mesh_proto.DisplayName, "other", "mt-1")},
		}),
		// env
		Entry("env: equal on a Universal zone", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.EnvTag: "universal"}, cp: universalNonFederated,
		}),
		Entry("env: differs on a Universal zone", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.EnvTag: "kubernetes"}, cp: universalNonFederated,
			expected: []validators.Violation{differs(mesh_proto.EnvTag, "kubernetes", "universal")},
		}),
		Entry("env: equal on a k8s zone", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.EnvTag: "kubernetes"}, cp: k8sNonFederated,
		}),
		Entry("env: on a global CP", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.EnvTag: "universal"}, cp: universalGlobal,
			expected: []validators.Violation{notHere(mesh_proto.EnvTag)},
		}),
		// namespace
		Entry("namespace: equal on k8s", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.KubeNamespaceTag: "kuma-demo"}, cp: k8sNonFederated,
		}),
		Entry("namespace: differs on k8s", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.KubeNamespaceTag: "other-ns"}, cp: k8sNonFederated,
			expected: []validators.Violation{differs(mesh_proto.KubeNamespaceTag, "other-ns", "kuma-demo")},
		}),
		Entry("namespace: on a cluster-scoped object on k8s", testCase{
			r: meshResource(), labels: map[string]string{mesh_proto.KubeNamespaceTag: "kuma-demo"}, cp: k8sNonFederated,
			expected: []validators.Violation{notHere(mesh_proto.KubeNamespaceTag)},
		}),
		Entry("namespace: on Universal", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.KubeNamespaceTag: "kuma-demo"}, cp: universalNonFederated,
			expected: []validators.Violation{notHere(mesh_proto.KubeNamespaceTag)},
		}),
		// service-account
		Entry("service-account: on a Dataplane on k8s", testCase{
			r: dataplane(), ns: appNamespace, labels: map[string]string{metadata.KumaServiceAccount: "victim-sa"}, cp: k8sGlobal,
			expected: []validators.Violation{notHere(metadata.KumaServiceAccount)},
		}),
		Entry("service-account: on a policy on k8s", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{metadata.KumaServiceAccount: "victim-sa"}, cp: k8sFederated,
			expected: []validators.Violation{notHere(metadata.KumaServiceAccount)},
		}),
		Entry("service-account: on Universal", testCase{
			r: dataplane(), labels: map[string]string{metadata.KumaServiceAccount: "victim-sa"}, cp: universalNonFederated,
			expected: []validators.Violation{notHere(metadata.KumaServiceAccount)},
		}),
		Entry("service-account: for a trusted writer on k8s", testCase{
			r: dataplane(), ns: appNamespace, labels: map[string]string{metadata.KumaServiceAccount: "victim-sa"}, trusted: true, cp: k8sGlobal,
		}),
		// listeners
		Entry("listener-zoneingress: equal on a Dataplane with the listener", testCase{
			r: zoneIngressDataplane(), labels: map[string]string{mesh_proto.ListenerZoneIngressLabel: "enabled"}, cp: universalNonFederated,
		}),
		Entry("listener-zoneingress: on a Dataplane without the listener", testCase{
			r: dataplane(), labels: map[string]string{mesh_proto.ListenerZoneIngressLabel: "enabled"}, cp: universalNonFederated,
			expected: []validators.Violation{notHere(mesh_proto.ListenerZoneIngressLabel)},
		}),
		Entry("listener-zoneegress: on a policy", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ListenerZoneEgressLabel: "enabled"}, cp: universalNonFederated,
			expected: []validators.Violation{notHere(mesh_proto.ListenerZoneEgressLabel)},
		}),
		// user-owned
		Entry("workload: any value", testCase{
			r: timeout(), labels: map[string]string{metadata.KumaWorkload: "anything"}, cp: universalNonFederated,
		}),
		// syntax
		Entry("syntax: violations are sorted by key", testCase{
			r: timeout(), labels: map[string]string{"kuma.io/b": "-bad", "kuma.io/a": "bad-", "bad key": "v"}, cp: universalNonFederated,
			expected: []validators.Violation{
				violation("bad key", "name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')"),
				violation("kuma.io/a", "a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')"),
				violation("kuma.io/b", "a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')"),
			},
		}),
		Entry("syntax: ownership violations are listed before syntax violations", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ZoneTag: "zone-2", "kuma.io/a": "bad-"}, cp: universalNonFederated,
			expected: []validators.Violation{
				differs(mesh_proto.ZoneTag, "zone-2", "zone-1"),
				violation("kuma.io/a", "a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')"),
			},
		}),
		Entry("syntax: annotation-backed values are names", testCase{
			r: timeout(), labels: map[string]string{metadata.KumaWorkload: "a.very.long.name.that.would.not.fit.a.label.value.because.it.is.longer.than.sixty-three.characters"}, cp: universalNonFederated,
		}),
		Entry("syntax: annotation-backed values must be DNS subdomains", testCase{
			r: timeout(), labels: map[string]string{metadata.KumaWorkload: "Not_A_Name"}, cp: universalNonFederated,
			expected: []validators.Violation{violation(metadata.KumaWorkload, "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')")},
		}),
	)
})
