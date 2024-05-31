package hostname_test

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/hostname"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
)

var _ = Describe("Hostname Generator", func() {
	var stopChSend chan<- struct{}
	var resManager manager.ResourceManager

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		resManager = manager.NewResourceManager(memory.NewStore())
		allocator, err := hostname.NewGenerator(logr.Discard(), m, resManager, 50*time.Millisecond)
		Expect(err).ToNot(HaveOccurred())
		ch := make(chan struct{})
		var stopChRecv <-chan struct{}
		stopChSend, stopChRecv = ch, ch
		go func() {
			defer GinkgoRecover()
			Expect(allocator.Start(stopChRecv)).To(Succeed())
		}()

		Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())

		generator := hostnamegenerator_api.NewHostnameGeneratorResource()
		generator.Meta = &test_model.ResourceMeta{
			Mesh: core_model.DefaultMesh,
			Name: "backend",
		}
		generator.Spec = &hostnamegenerator_api.HostnameGenerator{
			Template: "{{ .Name }}.mesh",
			Selector: hostnamegenerator_api.Selector{
				MeshService: hostnamegenerator_api.LabelSelector{
					MatchLabels: map[string]string{
						"label": "value",
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), generator, store.CreateBy(core_model.MetaToResourceKey(generator.GetMeta())))).To(Succeed())
	})

	AfterEach(func() {
		close(stopChSend)
		Expect(resManager.Delete(context.Background(), &meshservice_api.MeshServiceResource{}, store.DeleteByKey("backend", "default"))).To(Succeed())
	})

	vipOfMeshService := func(name string) *meshservice_api.MeshServiceStatus {
		ms := meshservice_api.NewMeshServiceResource()
		err := resManager.Get(context.Background(), ms, store.GetByKey(name, model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		return ms.Status
	}

	It("should not generate hostname if no generator selects a given MeshService", func() {
		// when
		err := samples.MeshServiceBackendBuilder().WithoutVIP().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			status := vipOfMeshService("backend")
			g.Expect(status.Addresses).Should(BeEmpty())
			g.Expect(status.HostnameGenerators).Should(BeEmpty())
		}, "10s", "100ms").Should(Succeed())
	})

	It("should generate hostname if a generator selects a given MeshService", func() {
		// when
		err := samples.MeshServiceBackendBuilder().WithoutVIP().WithLabels(map[string]string{
			"label": "value",
		}).Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			status := vipOfMeshService("backend")
			g.Expect(status.Addresses).Should(Not(BeEmpty()))
			g.Expect(status.HostnameGenerators).Should(Not(BeEmpty()))
		}, "2s", "100ms").Should(Succeed())
	})
})
