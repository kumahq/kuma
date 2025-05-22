package webhooks_test

import (
	"context"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/config/core"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	k8s_resources "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("Validating Webhook", func() {
	DescribeTableSubtree("Handle",
		func(inputFile string) {
			FIt("should validate requests on Global", func() {
				// given
				wh := newValidatingWebhook(core.Global, false)
				req := webhookRequest(inputFile)
				// when
				resp := wh.Handle(context.Background(), req)
				// then
				bytes, err := yaml.Marshal(resp)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(matchers.MatchGoldenYAML(strings.ReplaceAll(inputFile, ".input.yaml", ".global.golden.yaml")))
			})

			It("should validate requests on federated zone", func() {
				// given
				wh := newValidatingWebhook(core.Zone, true)
				req := webhookRequest(inputFile)
				// when
				resp := wh.Handle(context.Background(), req)
				// then
				bytes, err := yaml.Marshal(resp)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(matchers.MatchGoldenYAML(strings.ReplaceAll(inputFile, ".input.yaml", ".federated.golden.yaml")))
			})

			It("should validate requests on non-federated zone", func() {
				// given
				wh := newValidatingWebhook(core.Zone, false)
				req := webhookRequest(inputFile)
				// when
				resp := wh.Handle(context.Background(), req)
				// then
				bytes, err := yaml.Marshal(resp)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(matchers.MatchGoldenYAML(strings.ReplaceAll(inputFile, ".input.yaml", ".non-federated.golden.yaml")))
			})
		},
		test.EntriesForFolder("validation"),
	)
})

func newValidatingWebhook(mode core.CpMode, federatedZone bool) *kube_admission.Webhook {
	checker := webhooks.ResourceAdmissionChecker{
		AllowedUsers: []string{
			"system:serviceaccount:kube-system:generic-garbage-collector",
			"system:serviceaccount:kuma-system:kuma-control-plane",
		},
		Mode:                         mode,
		FederatedZone:                federatedZone,
		DisableOriginLabelValidation: false,
		SystemNamespace:              "kuma-system",
		ZoneName:                     "zone-1",
	}
	handler := webhooks.NewValidatingWebhook(k8s_resources.NewSimpleConverter(), core_registry.Global(), k8s_registry.Global(), checker)
	handler.InjectDecoder(kube_admission.NewDecoder(scheme))
	return &kube_admission.Webhook{
		Handler: handler,
	}
}

func webhookRequest(inputFile string) kube_admission.Request {
	input, err := os.ReadFile(inputFile)
	Expect(err).NotTo(HaveOccurred())

	lines := strings.Split(string(input), "\n")

	var user string
	var op string
	var ns string
	for _, l := range lines {
		if strings.HasPrefix(l, "#") {
			l = strings.Trim(l, "# ")
			keyValuePairs := strings.Split(l, ",")
			for _, kv := range keyValuePairs {
				pair := strings.Split(kv, "=")
				switch pair[0] {
				case "user":
					user = pair[1]
				case "operation":
					op = pair[1]
				case "namespace":
					ns = pair[1]
				}
			}
		} else {
			break
		}
	}

	resources := strings.Split(string(input), "---")
	var obj, oldObj []byte
	switch op {
	case "UPDATE":
		Expect(resources).To(HaveLen(2))
		obj, oldObj = []byte(resources[0]), []byte(resources[1])
	case "DELETE":
		Expect(resources).To(HaveLen(1))
		oldObj = []byte(resources[0])
	default:
		Expect(resources).To(HaveLen(1))
		obj = []byte(resources[0])
	}

	gvk := struct {
		ApiVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
	}{}
	Expect(yaml.Unmarshal([]byte(resources[0]), &gvk)).To(Succeed())
	gv, err := schema.ParseGroupVersion(gvk.ApiVersion)
	Expect(err).ToNot(HaveOccurred())

	req := kube_admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			UID: kube_types.UID("12345"),
			Object: kube_runtime.RawExtension{
				Raw: obj,
			},
			OldObject: kube_runtime.RawExtension{
				Raw: oldObj,
			},
			Kind: kube_meta.GroupVersionKind{
				Group:   gv.Group,
				Version: gv.Version,
				Kind:    gvk.Kind,
			},
			UserInfo: authenticationv1.UserInfo{
				Username: user,
			},
			Operation: admissionv1.Operation(op),
			Namespace: ns,
		},
	}

	return req
}
