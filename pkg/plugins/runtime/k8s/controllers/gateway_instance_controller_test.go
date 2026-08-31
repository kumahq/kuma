package controllers_test

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	mesh_k8s "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/controllers"
)

// recordingSink records the values every log entry is written with, so that we
// can assert the logger doesn't accumulate them across invocations.
type recordingSink struct {
	values  []any
	entries *[][]any
}

func (s *recordingSink) Init(logr.RuntimeInfo) {}

func (s *recordingSink) Enabled(int) bool { return true }

func (s *recordingSink) Info(_ int, _ string, keysAndValues ...any) {
	*s.entries = append(*s.entries, s.merged(keysAndValues))
}

func (s *recordingSink) Error(_ error, _ string, keysAndValues ...any) {
	*s.entries = append(*s.entries, s.merged(keysAndValues))
}

func (s *recordingSink) WithValues(keysAndValues ...any) logr.LogSink {
	return &recordingSink{values: s.merged(keysAndValues), entries: s.entries}
}

func (s *recordingSink) WithName(string) logr.LogSink { return s }

func (s *recordingSink) merged(keysAndValues []any) []any {
	values := make([]any, 0, len(s.values)+len(keysAndValues))
	values = append(values, s.values...)
	return append(values, keysAndValues...)
}

// emptyClient lists no objects, which is all the mapper needs from a client.
type emptyClient struct {
	kube_client.Client
}

func (emptyClient) List(context.Context, kube_client.ObjectList, ...kube_client.ListOption) error {
	return nil
}

var _ = Describe("GatewayToInstanceMapper", func() {
	gateway := func(name string, serviceTags ...string) *mesh_k8s.MeshGateway {
		selectors := ""
		for i, tag := range serviceTags {
			if i > 0 {
				selectors += ","
			}
			selectors += `{"match":{"kuma.io/service":"` + tag + `"}}`
		}
		return &mesh_k8s.MeshGateway{
			Name: name,
			Mesh: "default",
			Spec: &apiextensionsv1.JSON{
				Raw: []byte(`{"selectors":[` + selectors + `]}`),
			},
		}
	}

	newMapper := func(entries *[][]any) kube_handler.MapFunc {
		return controllers.GatewayToInstanceMapper(logr.New(&recordingSink{entries: entries}), emptyClient{})
	}

	It("should map service tags to MeshGatewayInstance names", func() {
		// given
		var entries [][]any
		mapper := newMapper(&entries)

		// when
		requests := mapper(context.Background(), gateway("edge-gateway", "gateway_kuma-demo_svc_8080"))

		// then
		Expect(requests).To(Equal([]kube_reconcile.Request{{
			Name: "gateway", Namespace: "kuma-demo",
		}}))
		Expect(entries).To(BeEmpty())
	})

	It("should not accumulate log values across invocations", func() {
		// given
		var entries [][]any
		mapper := newMapper(&entries)
		// a service tag that isn't in the <name>_<namespace>_svc_<port> form makes
		// the mapper log an error, which is what surfaces accumulated values
		obj := gateway("edge-gateway", "edge-gateway")

		// when
		for range 3 {
			Expect(mapper(context.Background(), obj)).To(BeEmpty())
		}

		// then
		Expect(entries).To(Equal([][]any{
			{"gateway", "edge-gateway", "mesh", "default"},
			{"gateway", "edge-gateway", "mesh", "default"},
			{"gateway", "edge-gateway", "mesh", "default"},
		}))
	})

	It("should not leak values of one MeshGateway into another", func() {
		// given
		var entries [][]any
		mapper := newMapper(&entries)

		// when
		Expect(mapper(context.Background(), gateway("edge-gateway", "edge-gateway"))).To(BeEmpty())
		Expect(mapper(context.Background(), gateway("second-gateway", "second-gateway"))).To(BeEmpty())

		// then
		Expect(entries).To(Equal([][]any{
			{"gateway", "edge-gateway", "mesh", "default"},
			{"gateway", "second-gateway", "mesh", "default"},
		}))
	})
})
