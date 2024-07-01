package hostname_test

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/hostname"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	mzms_hostname "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/hostname"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
)

var _ = Describe("MeshMultiZoneService Hostname Generator", func() {
	var stopChSend chan<- struct{}
	var resManager manager.ResourceManager

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		resManager = manager.NewResourceManager(memory.NewStore())
		allocator, err := hostname.NewGenerator(
			logr.Discard(), m, resManager, 50*time.Millisecond,
			[]hostname.HostnameGenerator{mzms_hostname.NewMeshMultiZoneServiceHostnameGenerator(resManager)},
		)
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
			Name: "example",
		}
		generator.Spec = &hostnamegenerator_api.HostnameGenerator{
			Template: "{{ .Name }}.mesh",
			Selector: hostnamegenerator_api.Selector{
				MeshMultiZoneService: hostnamegenerator_api.LabelSelector{
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
	})

	multiZoneServiceStatus := func(name string) *meshmzservice_api.MeshMultiZoneServiceStatus {
		mzms := meshmzservice_api.NewMeshMultiZoneServiceResource()
		err := resManager.Get(context.Background(), mzms, store.GetByKey(name, model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		return mzms.Status
	}

	It("should not generate hostname if no generator selects a given MeshMultiZoneService", func() {
		// when
		err := samples.MeshMultiZoneServiceBackendBuilder().WithName("example").Create(resManager)

		// then
		Expect(err).ToNot(HaveOccurred())
		Consistently(func(g Gomega) {
			status := multiZoneServiceStatus("example")
			g.Expect(status.Addresses).Should(BeEmpty())
			g.Expect(status.HostnameGenerators).Should(BeEmpty())
		}, "2s", "100ms").Should(Succeed())
	})

	It("should generate hostname if a generator selects a given MeshMultiZoneService", func() {
		// when
		err := samples.MeshMultiZoneServiceBackendBuilder().
			WithLabels(map[string]string{"label": "value"}).
			WithName("example").
			Create(resManager)

		// then
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			status := multiZoneServiceStatus("example")
			g.Expect(status.Addresses).Should(Not(BeEmpty()))
			g.Expect(status.HostnameGenerators).Should(Not(BeEmpty()))
		}, "2s", "100ms").Should(Succeed())
	})
})
