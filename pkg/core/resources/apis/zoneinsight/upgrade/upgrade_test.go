package upgrade_test

import (
	"context"
	"encoding/json"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/yaml"

	zoneinsight_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zoneinsight/api/v1alpha1"
	zoneinsight_k8s "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zoneinsight/k8s/v1alpha1"
)

// The stored object is written while the 2.14 definition is installed, then the 3.0
// definition replaces it in place, the way an upgrade does. Nothing may be pruned or
// rejected: the definition is opaque on both sides precisely so the API server does not
// touch a spec it cannot interpret.
const (
	crdFrom2_14 = "testdata/zoneinsights.2.14.crd.yaml"
	crdShipped  = "../../../../../../deployments/charts/kuma/crds/kuma.io_zoneinsights.yaml"
)

var _ = Describe("ZoneInsight across a definition upgrade", Ordered, func() {
	var (
		env    *envtest.Environment
		cfg    *rest.Config
		cl     client.Client
		stored = "zone-1"
	)

	spec := map[string]any{
		"subscriptions": []any{map[string]any{
			"id":               "sub-1",
			"globalInstanceId": "global-01",
			"connectTime":      "2026-03-14T09:05:30.100Z",
			"config":           `{"store":{"type":"postgres"}}`,
			"status": map[string]any{
				"total": map[string]any{"responsesSent": "18446744073709551615"},
			},
			"fieldRemovedIn30": "still here",
		}},
		"kdsStreams": map[string]any{
			"clusters": map[string]any{"globalInstanceId": "c"},
		},
	}

	applyCRD := func(path string) {
		raw, err := os.ReadFile(path)
		Expect(err).ToNot(HaveOccurred())
		crd := &apiextensionsv1.CustomResourceDefinition{}
		Expect(yaml.Unmarshal(raw, crd)).To(Succeed())

		existing := &apiextensionsv1.CustomResourceDefinition{}
		err = cl.Get(context.Background(), types.NamespacedName{Name: crd.Name}, existing)
		if err == nil {
			crd.ResourceVersion = existing.ResourceVersion
			Expect(cl.Update(context.Background(), crd)).To(Succeed())
			return
		}
		Expect(cl.Create(context.Background(), crd)).To(Succeed())
	}

	BeforeAll(func() {
		env = &envtest.Environment{}
		var err error
		cfg, err = env.Start()
		Expect(err).ToNot(HaveOccurred())

		Expect(apiextensionsv1.AddToScheme(scheme.Scheme)).To(Succeed())
		Expect(zoneinsight_k8s.AddToScheme(scheme.Scheme)).To(Succeed())
		cl, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		Expect(env.Stop()).To(Succeed())
	})

	It("stores an object under the 2.14 definition", func() {
		applyCRD(crdFrom2_14)

		raw, err := json.Marshal(spec)
		Expect(err).ToNot(HaveOccurred())

		obj := &zoneinsight_k8s.ZoneInsight{Spec: &apiextensionsv1.JSON{Raw: raw}}
		obj.SetName(stored)

		Eventually(func() error {
			return cl.Create(context.Background(), obj)
		}, "30s", "500ms").Should(Succeed())
	})

	It("keeps every stored byte when the 3.0 definition replaces it", func() {
		applyCRD(crdShipped)

		fetched := &zoneinsight_k8s.ZoneInsight{}
		Eventually(func() error {
			return cl.Get(context.Background(), types.NamespacedName{Name: stored}, fetched)
		}, "30s", "500ms").Should(Succeed())

		var got map[string]any
		Expect(json.Unmarshal(fetched.Spec.Raw, &got)).To(Succeed())
		Expect(got).To(Equal(spec))
	})

	It("reads the stored spec into the Go type without losing the counter or the timestamp", func() {
		fetched := &zoneinsight_k8s.ZoneInsight{}
		Expect(cl.Get(context.Background(), types.NamespacedName{Name: stored}, fetched)).To(Succeed())

		coreSpec, err := fetched.GetSpec()
		Expect(err).ToNot(HaveOccurred())

		insight := coreSpec.(*zoneinsight_api.ZoneInsight)
		Expect(insight.Subscriptions).To(HaveLen(1))
		Expect(insight.Subscriptions[0].ID).To(Equal("sub-1"))
		Expect(insight.Subscriptions[0].Status.Total.ResponsesSent).To(Equal(uint64(18446744073709551615)))
		Expect(insight.Subscriptions[0].ConnectTime.Time).
			To(Equal(time.Date(2026, 3, 14, 9, 5, 30, 100000000, time.UTC)))
	})

	It("accepts an update under the 3.0 definition, so a stored object is never wedged", func() {
		fetched := &zoneinsight_k8s.ZoneInsight{}
		Expect(cl.Get(context.Background(), types.NamespacedName{Name: stored}, fetched)).To(Succeed())

		fetched.Labels = map[string]string{"touched": "yes"}
		Expect(cl.Update(context.Background(), fetched)).To(Succeed())

		again := &zoneinsight_k8s.ZoneInsight{}
		Expect(cl.Get(context.Background(), types.NamespacedName{Name: stored}, again)).To(Succeed())

		var got map[string]any
		Expect(json.Unmarshal(again.Spec.Raw, &got)).To(Succeed())
		Expect(got).To(Equal(spec), "the update must not prune the unknown field")
	})
})
