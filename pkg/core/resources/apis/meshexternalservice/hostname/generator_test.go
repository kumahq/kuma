package hostname_test

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/hostname"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	mes_hostname "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/hostname"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
)

var _ = Describe("MeshExternalService Hostname Generator", func() {
	var stopChSend chan<- struct{}
	var resManager manager.ResourceManager

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		resManager = manager.NewResourceManager(memory.NewStore())
		allocator, err := hostname.NewGenerator(
			logr.Discard(), m, resManager, "", 50*time.Millisecond,
			[]hostname.HostnameGenerator{mes_hostname.NewMeshExternalServiceHostnameGenerator(resManager)},
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

		err = builders.HostnameGenerator().
			WithName("display-name").
			WithTemplate("{{ .DisplayName }}.dn.mesh").
			WithMeshExternalServiceMatchLabels(map[string]string{"dn": "true"}).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		err = builders.HostnameGenerator().
			WithName("example").
			WithTemplate("{{ .Name }}.mesh").
			WithMeshExternalServiceMatchLabels(map[string]string{"label": "value"}).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		err = builders.HostnameGenerator().
			WithName("static").
			WithTemplate("static.mesh").
			WithMeshExternalServiceMatchLabels(map[string]string{"generate": "static"}).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		close(stopChSend)
		Expect(resManager.DeleteAll(context.Background(), &meshexternalservice_api.MeshExternalServiceResourceList{})).To(Succeed())
	})

	meshExternalServiceStatus := func(name string) *meshexternalservice_api.MeshExternalServiceStatus {
		ms := meshexternalservice_api.NewMeshExternalServiceResource()
		err := resManager.Get(context.Background(), ms, store.GetByKey(name, model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		return ms.Status
	}

	It("should not generate hostname if no generator selects a given MeshExternalService", func() {
		// when
		err := samples.MeshExternalServiceExampleBuilder().WithoutVIP().WithName("example").Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Consistently(func(g Gomega) {
			status := meshExternalServiceStatus("example")
			g.Expect(status.Addresses).Should(BeEmpty())
			g.Expect(status.HostnameGenerators).Should(BeEmpty())
		}, "1s", "100ms").Should(Succeed())
	})

	It("should generate hostname if a generator selects a given MeshExternalService", func() {
		// when
		err := samples.MeshExternalServiceExampleBuilder().WithoutVIP().WithLabels(map[string]string{
			"label": "value",
		}).Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			status := meshExternalServiceStatus("example")
			g.Expect(status.Addresses).Should(Not(BeEmpty()))
			g.Expect(status.HostnameGenerators).Should(Not(BeEmpty()))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should generate hostname if for MeshExternalService using .DisplayName taken from label", func() {
		// when
		err := samples.MeshExternalServiceExampleBuilder().
			WithoutVIP().
			WithName("test-external-svc").
			WithLabels(map[string]string{"kuma.io/display-name": "other-name", "dn": "true"}).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			status := meshExternalServiceStatus("test-external-svc")
			g.Expect(status.Addresses).Should(Not(BeEmpty()))
			g.Expect(status.Addresses[0].Hostname).Should(Equal("other-name.dn.mesh"))
			g.Expect(status.HostnameGenerators).Should(Not(BeEmpty()))
		}, "2000s", "100ms").Should(Succeed())
	})

	It("should generate hostname if for MeshExternalService using .DisplayName taken from name", func() {
		// when
		err := samples.MeshExternalServiceExampleBuilder().
			WithoutVIP().
			WithName("test-external-svc").
			WithLabels(map[string]string{"dn": "true"}).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			status := meshExternalServiceStatus("test-external-svc")
			g.Expect(status.Addresses).Should(Not(BeEmpty()))
			g.Expect(status.Addresses[0].Hostname).Should(Equal("test-external-svc.dn.mesh"))
			g.Expect(status.HostnameGenerators).Should(Not(BeEmpty()))
		}, "2000s", "100ms").Should(Succeed())
	})

	It("should set an error if there's a collision", func() {
		// when
		Expect(
			samples.MeshExternalServiceExampleBuilder().WithoutVIP().WithLabels(map[string]string{
				"generate": "static",
			}).Create(resManager),
		).To(Succeed())
		Expect(
			samples.MeshExternalServiceExampleBuilder().WithoutVIP().WithLabels(map[string]string{
				"generate": "static",
			}).WithName("other").Create(resManager),
		).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			otherStatus := meshExternalServiceStatus("other")
			exampleStatus := meshExternalServiceStatus("example")
			g.Expect(otherStatus.Addresses).Should(BeEmpty())
			g.Expect(otherStatus.HostnameGenerators).Should(ConsistOf(
				hostnamegenerator_api.HostnameGeneratorStatus{
					HostnameGeneratorRef: hostnamegenerator_api.HostnameGeneratorRef{CoreName: "static"},
					Conditions: []hostnamegenerator_api.Condition{{
						Type:    hostnamegenerator_api.GeneratedCondition,
						Status:  kube_meta.ConditionFalse,
						Reason:  hostnamegenerator_api.CollisionReason,
						Message: "Hostname collision with MeshExternalService: other",
					}},
				},
			))
			g.Expect(exampleStatus.Addresses).Should(Not(BeEmpty()))
			g.Expect(exampleStatus.HostnameGenerators).Should(ConsistOf(
				hostnamegenerator_api.HostnameGeneratorStatus{
					HostnameGeneratorRef: hostnamegenerator_api.HostnameGeneratorRef{CoreName: "static"},
					Conditions: []hostnamegenerator_api.Condition{{
						Type:   hostnamegenerator_api.GeneratedCondition,
						Status: kube_meta.ConditionTrue,
						Reason: hostnamegenerator_api.GeneratedReason,
					}},
				},
			))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should not generate hostname when selector is not MeshExternalService", func() {
		// when
		err := builders.HostnameGenerator().
			WithName("ms-generator").
			WithTemplate("{{ .DisplayName }}.mesh").
			WithMeshServiceMatchLabels(map[string]string{"test": "true"}).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		err = samples.MeshExternalServiceExampleBuilder().
			WithLabels(map[string]string{"test": "true"}).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			status := meshExternalServiceStatus("example")
			g.Expect(status.Addresses).Should(BeEmpty())
			g.Expect(status.HostnameGenerators).Should(BeEmpty())
		}, "2s", "100ms").Should(Succeed())
	})
})
