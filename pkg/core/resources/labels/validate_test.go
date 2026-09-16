package labels_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var _ = Describe("Validate", func() {
	timeout := func() core_model.Resource {
		return builders.MeshTimeout().
			WithMesh("mesh-1").
			WithTargetRef(builders.TargetRefMesh()).
			AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
				IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
			}).
			Build()
	}
	dataplane := func() core_model.Resource {
		return builders.Dataplane().WithMesh("mesh-1").WithServices("backend").Build()
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
		trusted  bool
		cp       resource_labels.ControlPlane
		expected []validators.Violation
	}

	violation := func(key, msg string) validators.Violation {
		return validators.Violation{Field: validators.Root().Key(key).String(), Message: msg}
	}

	DescribeTable("should report the violations of today's rules",
		func(given testCase) {
			w := resource_labels.Write{
				Descriptor:    given.r.Descriptor(),
				Spec:          given.r.GetSpec(),
				Namespace:     given.ns,
				Mesh:          given.r.GetMeta().GetMesh(),
				Labels:        given.labels,
				TrustedWriter: given.trusted,
			}
			err := resource_labels.Validate(w, given.cp)
			if given.expected == nil {
				Expect(err.HasViolations()).To(BeFalse(), err.Error())
			} else {
				Expect(err.Violations).To(Equal(given.expected))
			}
		},
		// origin, Universal
		Entry("origin: global on a global CP", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: universalGlobal,
		}),
		Entry("origin: zone on a global CP", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: "zone"}, cp: universalGlobal,
			expected: []validators.Violation{violation(mesh_proto.ResourceOriginLabel, "the origin label must be set to 'global'")},
		}),
		Entry("origin: global on a federated zone", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: universalFederated,
			expected: []validators.Violation{violation(mesh_proto.ResourceOriginLabel, "the origin label must be set to 'zone'")},
		}),
		Entry("origin: global on a non-federated zone", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: universalNonFederated,
		}),
		Entry("origin: empty is not present", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: ""}, cp: universalGlobal,
		}),
		Entry("origin: unknown value fails the format rule", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: "unknownvalue"}, cp: universalNonFederated,
			expected: []validators.Violation{violation(mesh_proto.ResourceOriginLabel, `unknown resource origin "unknownvalue"`)},
		}),
		Entry("origin: unknown value fails the format rule for a trusted writer", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: "unknownvalue"}, trusted: true, cp: universalGlobal,
			expected: []validators.Violation{violation(mesh_proto.ResourceOriginLabel, `unknown resource origin "unknownvalue"`)},
		}),
		Entry("origin: format violation is listed before the ownership violation", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ResourceOriginLabel: "unknownvalue"}, cp: universalGlobal,
			expected: []validators.Violation{
				violation(mesh_proto.ResourceOriginLabel, `unknown resource origin "unknownvalue"`),
				violation(mesh_proto.ResourceOriginLabel, "the origin label must be set to 'global'"),
			},
		}),
		// origin, k8s
		Entry("origin: zone on a k8s global CP", testCase{
			r: timeout(), ns: systemNamespace, labels: map[string]string{mesh_proto.ResourceOriginLabel: "zone"}, cp: k8sGlobal,
			expected: []validators.Violation{violation(mesh_proto.ResourceOriginLabel, "'kuma.io/origin' label should have 'global' value, got 'zone'")},
		}),
		Entry("origin: global in the system namespace of a k8s federated zone", testCase{
			r: timeout(), ns: systemNamespace, labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: k8sFederated,
			expected: []validators.Violation{violation(mesh_proto.ResourceOriginLabel, "'kuma.io/origin' label should have 'zone' value, got 'global'")},
		}),
		Entry("origin: global in an app namespace of a k8s federated zone", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: k8sFederated,
		}),
		Entry("origin: global on a k8s non-federated zone", testCase{
			r: timeout(), ns: systemNamespace, labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: k8sNonFederated,
		}),
		Entry("origin: a non-plugin type is not checked on k8s", testCase{
			r: dataplane(), ns: systemNamespace, labels: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: k8sFederated,
		}),
		Entry("origin: unknown value fails the format rule on k8s", testCase{
			r: timeout(), ns: systemNamespace, labels: map[string]string{mesh_proto.ResourceOriginLabel: "unknownvalue"}, cp: k8sNonFederated,
			expected: []validators.Violation{violation(mesh_proto.ResourceOriginLabel, `unknown resource origin "unknownvalue"`)},
		}),
		// zone, Universal
		Entry("zone: any value on a global CP", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ZoneTag: "zone-1"}, cp: universalGlobal,
			expected: []validators.Violation{violation(mesh_proto.ZoneTag, "kuma.io/zone is not allowed on a global control plane")},
		}),
		Entry("zone: the local zone", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ZoneTag: "zone-1"}, cp: universalNonFederated,
		}),
		Entry("zone: another zone", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ZoneTag: "zone-2"}, cp: universalNonFederated,
			expected: []validators.Violation{violation(mesh_proto.ZoneTag, "kuma.io/zone label should have zone-1 value")},
		}),
		Entry("origin is reported before zone", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.ZoneTag: "zone-2", mesh_proto.ResourceOriginLabel: "global"}, cp: universalFederated,
			expected: []validators.Violation{
				violation(mesh_proto.ResourceOriginLabel, "the origin label must be set to 'zone'"),
				violation(mesh_proto.ZoneTag, "kuma.io/zone label should have zone-1 value"),
			},
		}),
		// zone, k8s
		Entry("zone: another zone with a zone origin on a k8s federated zone", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.ResourceOriginLabel: "zone", mesh_proto.ZoneTag: "zone-2"}, cp: k8sFederated,
			expected: []validators.Violation{violation(mesh_proto.ZoneTag, "'kuma.io/zone' label should have 'zone-1' value, got 'zone-2'")},
		}),
		Entry("zone: another zone without an origin on a k8s federated zone", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.ZoneTag: "zone-2"}, cp: k8sFederated,
		}),
		Entry("zone: another zone on a k8s non-federated zone", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.ResourceOriginLabel: "zone", mesh_proto.ZoneTag: "zone-2"}, cp: k8sNonFederated,
		}),
		Entry("zone: any value on a k8s global CP", testCase{
			r: timeout(), ns: systemNamespace, labels: map[string]string{mesh_proto.ZoneTag: "zone-2"}, cp: k8sGlobal,
		}),
		// mesh
		Entry("mesh: the resource's mesh", testCase{
			r: timeout(), labels: map[string]string{metadata.KumaMeshLabel: "mesh-1"}, cp: universalNonFederated,
		}),
		Entry("mesh: another mesh", testCase{
			r: timeout(), labels: map[string]string{metadata.KumaMeshLabel: "mesh-2"}, cp: universalNonFederated,
			expected: []validators.Violation{violation(metadata.KumaMeshLabel, "kuma.io/mesh label must not differ from mesh set on resource")},
		}),
		Entry("mesh: another mesh on k8s", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{metadata.KumaMeshLabel: "mesh-2"}, cp: k8sNonFederated,
		}),
		// policy-role
		Entry("policy-role: system", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.PolicyRoleLabel: "system"}, cp: universalNonFederated,
		}),
		Entry("policy-role: empty reads as system", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.PolicyRoleLabel: ""}, cp: universalNonFederated,
		}),
		Entry("policy-role: consumer", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.PolicyRoleLabel: "consumer"}, cp: universalNonFederated,
			expected: []validators.Violation{violation(mesh_proto.PolicyRoleLabel, "kuma.io/policy-role label should have system value, got consumer")},
		}),
		Entry("policy-role: consumer on a non-policy", testCase{
			r: dataplane(), labels: map[string]string{mesh_proto.PolicyRoleLabel: "consumer"}, cp: universalNonFederated,
		}),
		Entry("policy-role: invalid on k8s", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{mesh_proto.PolicyRoleLabel: "invalid"}, cp: k8sFederated,
		}),
		// service-account
		Entry("service-account: on a Dataplane on Universal", testCase{
			r: dataplane(), labels: map[string]string{metadata.KumaServiceAccount: "victim-sa"}, cp: universalNonFederated,
		}),
		Entry("service-account: on a Dataplane on k8s", testCase{
			r: dataplane(), ns: appNamespace, labels: map[string]string{metadata.KumaServiceAccount: "victim-sa"}, cp: k8sGlobal,
			expected: []validators.Violation{violation(metadata.KumaServiceAccount, `Label "k8s.kuma.io/service-account" is managed by Kuma and cannot be set manually.`)},
		}),
		Entry("service-account: on a policy on k8s", testCase{
			r: timeout(), ns: appNamespace, labels: map[string]string{metadata.KumaServiceAccount: "victim-sa"}, cp: k8sFederated,
		}),
		Entry("service-account: on a Dataplane on k8s for a trusted writer", testCase{
			r: dataplane(), ns: appNamespace, labels: map[string]string{metadata.KumaServiceAccount: "victim-sa"}, trusted: true, cp: k8sGlobal,
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
		Entry("syntax: annotation-backed values are names", testCase{
			r: timeout(), labels: map[string]string{mesh_proto.DisplayName: "a.very.long.name.that.would.not.fit.a.label.value.because.it.is.longer.than.sixty-three.characters"}, cp: universalNonFederated,
		}),
		Entry("syntax: annotation-backed values must be DNS subdomains", testCase{
			r: timeout(), labels: map[string]string{metadata.KumaWorkload: "Not_A_Name"}, cp: universalNonFederated,
			expected: []validators.Violation{violation(metadata.KumaWorkload, "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')")},
		}),
	)

	type deleteCase struct {
		r        core_model.Resource
		ns       resource_labels.Namespace
		stored   map[string]string
		cp       resource_labels.ControlPlane
		expected *validators.Violation
	}

	DescribeTable("ValidateDelete should refuse a resource another control plane owns",
		func(given deleteCase) {
			err := resource_labels.ValidateDelete(resource_labels.NewStoredResource(given.r, given.ns, given.stored, given.cp), given.cp)
			if given.expected == nil {
				Expect(err.HasViolations()).To(BeFalse(), err.Error())
			} else {
				Expect(err.Violations).To(Equal([]validators.Violation{*given.expected}))
			}
		},
		Entry("Universal global CP: local", deleteCase{
			r: timeout(), stored: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: universalGlobal,
		}),
		Entry("Universal global CP: synced from a zone", deleteCase{
			r: timeout(), stored: map[string]string{mesh_proto.ResourceOriginLabel: "zone"}, cp: universalGlobal,
			expected: pointer.To(violation(mesh_proto.ResourceOriginLabel, "the origin label must be set to 'global'")),
		}),
		Entry("Universal global CP: unknown origin", deleteCase{
			r: timeout(), stored: map[string]string{mesh_proto.ResourceOriginLabel: "unknownvalue"}, cp: universalGlobal,
			expected: pointer.To(violation(mesh_proto.ResourceOriginLabel, "the origin label must be set to 'global'")),
		}),
		Entry("Universal global CP: no origin", deleteCase{
			r: timeout(), cp: universalGlobal,
		}),
		Entry("Universal global CP: empty origin", deleteCase{
			r: timeout(), stored: map[string]string{mesh_proto.ResourceOriginLabel: ""}, cp: universalGlobal,
		}),
		Entry("Universal federated zone: synced from global", deleteCase{
			r: timeout(), stored: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: universalFederated,
			expected: pointer.To(violation(mesh_proto.ResourceOriginLabel, "the origin label must be set to 'zone'")),
		}),
		Entry("Universal federated zone: local", deleteCase{
			r: timeout(), stored: map[string]string{mesh_proto.ResourceOriginLabel: "zone"}, cp: universalFederated,
		}),
		Entry("Universal non-federated zone owns everything", deleteCase{
			r: timeout(), stored: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: universalNonFederated,
		}),
		Entry("no mode owns everything", deleteCase{
			r: timeout(), stored: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: resource_labels.ControlPlane{},
		}),
		Entry("k8s global CP: synced from a zone", deleteCase{
			r: timeout(), ns: systemNamespace, stored: map[string]string{mesh_proto.ResourceOriginLabel: "zone"}, cp: k8sGlobal,
			expected: pointer.To(violation(mesh_proto.ResourceOriginLabel, "'kuma.io/origin' label should have 'global' value, got 'zone'")),
		}),
		Entry("k8s global CP: unknown origin", deleteCase{
			r: timeout(), ns: systemNamespace, stored: map[string]string{mesh_proto.ResourceOriginLabel: "unknownvalue"}, cp: k8sGlobal,
			expected: pointer.To(violation(mesh_proto.ResourceOriginLabel, "'kuma.io/origin' label should have 'global' value, got 'unknownvalue'")),
		}),
		Entry("k8s global CP: a non-plugin type synced from a zone", deleteCase{
			r: dataplane(), ns: systemNamespace, stored: map[string]string{mesh_proto.ResourceOriginLabel: "zone"}, cp: k8sGlobal,
			expected: pointer.To(violation(mesh_proto.ResourceOriginLabel, "'kuma.io/origin' label should have 'global' value, got 'zone'")),
		}),
		Entry("k8s federated zone: synced from global into the system namespace", deleteCase{
			r: timeout(), ns: systemNamespace, stored: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: k8sFederated,
			expected: pointer.To(violation(mesh_proto.ResourceOriginLabel, "'kuma.io/origin' label should have 'zone' value, got 'global'")),
		}),
		Entry("k8s federated zone: an app namespace is local whatever the origin says", deleteCase{
			r: timeout(), ns: appNamespace, stored: map[string]string{mesh_proto.ResourceOriginLabel: "global"}, cp: k8sFederated,
		}),
		Entry("k8s federated zone: a mismatching zone label does not change who owns it", deleteCase{
			r: timeout(), ns: systemNamespace, stored: map[string]string{mesh_proto.ResourceOriginLabel: "zone", mesh_proto.ZoneTag: "zone-2"}, cp: k8sFederated,
		}),
		Entry("k8s federated zone: a service-account label does not change who owns it", deleteCase{
			r: dataplane(), ns: appNamespace, stored: map[string]string{metadata.KumaServiceAccount: "victim-sa"}, cp: k8sFederated,
		}),
	)
})
