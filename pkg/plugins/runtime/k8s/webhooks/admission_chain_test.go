package webhooks_test

import (
	"context"
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	core_registry "github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	k8s_resources "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/webhooks"
)

// The API server runs the defaulting webhook first and hands its output to the
// validating webhook, so every label the defaulter computed is "supplied" by the
// time the validating webhook checks ownership, and an update carries the labels
// stored by the previous write.
var _ = Describe("Defaulting then validating webhook", func() {
	zoneChecker := webhooks.ResourceAdmissionChecker{
		AllowedUsers:    []string{"system:serviceaccount:kuma-system:kuma-control-plane"},
		ControlPlane:    resource_labels.ControlPlane{Mode: core.Zone, Zone: "kuma-2", IsK8s: true, FederatedZone: true},
		SystemNamespace: "kuma-system",
	}

	namespaceOf := func(raw []byte) string {
		var obj struct {
			Metadata struct {
				Namespace string `json:"namespace"`
			} `json:"metadata"`
		}
		Expect(json.Unmarshal(raw, &obj)).To(Succeed())
		return obj.Metadata.Namespace
	}

	request := func(op admissionv1.Operation, obj, oldObj []byte) kube_admission.Request {
		return kube_admission.Request{
			UID:       kube_types.UID("12345"),
			Kind:      kube_meta.GroupVersionKind{Group: "kuma.io", Version: "v1alpha1", Kind: "MeshTimeout"},
			Namespace: namespaceOf(obj),
			Operation: op,
			UserInfo:  authenticationv1.UserInfo{Username: "cli-user"},
			Object:    kube_runtime.RawExtension{Raw: obj},
			OldObject: kube_runtime.RawExtension{Raw: oldObj},
		}
	}

	// admit returns the defaulting response, the object as the defaulter left it, and
	// the validating response on that object.
	admit := func(checker webhooks.ResourceAdmissionChecker, op admissionv1.Operation, obj, oldObj []byte) (kube_admission.Response, []byte, kube_admission.Response) {
		converter := k8s_resources.NewSimpleConverter(checker.SystemNamespace, resource_labels.ControlPlane{})
		defaulter := webhooks.DefaultingWebhookFor(scheme, converter, checker)
		validator := webhooks.NewValidatingWebhook(converter, core_registry.Global(), k8s_registry.Global(), checker)
		validator.InjectDecoder(kube_admission.NewDecoder(scheme))

		defaulted := defaulter.Handle(context.Background(), request(op, obj, oldObj))
		patched := obj
		if len(defaulted.Patch) > 0 {
			patch, err := jsonpatch.DecodePatch(defaulted.Patch)
			Expect(err).ToNot(HaveOccurred())
			patched, err = patch.Apply(obj)
			Expect(err).ToNot(HaveOccurred())
		}
		validated := validator.Handle(context.Background(), request(op, patched, oldObj))
		return defaulted, patched, validated
	}

	labelsOf := func(raw []byte) map[string]string {
		var obj struct {
			Metadata struct {
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
		}
		Expect(json.Unmarshal(raw, &obj)).To(Succeed())
		return obj.Metadata.Labels
	}

	edit := func(raw []byte, fn func(obj map[string]any)) []byte {
		obj := map[string]any{}
		Expect(json.Unmarshal(raw, &obj)).To(Succeed())
		fn(obj)
		out, err := json.Marshal(obj)
		Expect(err).ToNot(HaveOccurred())
		return out
	}

	producerTimeout := func() []byte {
		raw, err := yaml.YAMLToJSON([]byte(`
kind: MeshTimeout
apiVersion: kuma.io/v1alpha1
metadata:
  name: to-test-server
  namespace: producer-policy-flow-ns
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: producer-policy-flow
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        labels:
          kuma.io/display-name: test-server
      default:
        http:
          requestTimeout: 2s
`))
		Expect(err).ToNot(HaveOccurred())
		return raw
	}

	It("should accept a create and an update that flips the computed policy role", func() {
		// when
		defaulted, stored, validated := admit(zoneChecker, admissionv1.Create, producerTimeout(), nil)

		// then
		Expect(defaulted.Allowed).To(BeTrue(), defaulted.Result.Message)
		Expect(validated.Allowed).To(BeTrue(), validated.Result.Message)
		Expect(labelsOf(stored)).To(HaveKeyWithValue("kuma.io/policy-role", "producer"))

		// when the stored object is re-applied with a to-item in another namespace
		updated := edit(stored, func(obj map[string]any) {
			to := obj["spec"].(map[string]any)["to"].([]any)[0].(map[string]any)
			to["targetRef"].(map[string]any)["labels"].(map[string]any)["k8s.kuma.io/namespace"] = "random-ns-name"
		})
		defaulted, stored, validated = admit(zoneChecker, admissionv1.Update, updated, stored)

		// then
		Expect(defaulted.Allowed).To(BeTrue(), defaulted.Result.Message)
		Expect(validated.Allowed).To(BeTrue(), validated.Result.Message)
		Expect(labelsOf(stored)).To(HaveKeyWithValue("kuma.io/policy-role", "consumer"))
	})

	It("should reject an update that changes a control-plane-owned label", func() {
		// given
		_, stored, _ := admit(zoneChecker, admissionv1.Create, producerTimeout(), nil)

		// when
		updated := edit(stored, func(obj map[string]any) {
			obj["metadata"].(map[string]any)["labels"].(map[string]any)["kuma.io/zone"] = "other-zone"
		})
		defaulted, _, _ := admit(zoneChecker, admissionv1.Update, updated, stored)

		// then
		Expect(defaulted.Allowed).To(BeFalse())
		Expect(defaulted.Result.Message).To(Equal(`Operation not allowed. label "kuma.io/zone" is managed by the control plane: got "other-zone", expected "kuma-2"`))
	})

	It("should reject an update of a zone-originated resource on Global once the defaulter recomputes its origin", func() {
		// given a resource synced from a zone, stored with the origin the zone gave it
		globalChecker := zoneChecker
		globalChecker.ControlPlane = resource_labels.ControlPlane{Mode: core.Global, IsK8s: true}
		stored := edit(producerTimeout(), func(obj map[string]any) {
			obj["metadata"].(map[string]any)["namespace"] = "kuma-system"
		})

		// when
		defaulted, patched, validated := admit(globalChecker, admissionv1.Update, stored, stored)

		// then the stored origin is not a value the writer chose, so ownership passes
		// and the defaulter recomputes it; immutability then refuses the change
		Expect(defaulted.Allowed).To(BeTrue(), defaulted.Result.Message)
		Expect(labelsOf(patched)).To(HaveKeyWithValue("kuma.io/origin", "global"))
		Expect(validated.Allowed).To(BeFalse())
		Expect(validated.Result.Message).To(Equal("Operation not allowed. 'kuma.io/origin' label is immutable, cannot be changed from 'zone' to 'global'"))
	})
})
