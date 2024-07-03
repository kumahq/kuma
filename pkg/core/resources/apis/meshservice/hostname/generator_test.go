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
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshservice_hostname "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/hostname"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
)

var _ = Describe("MeshService Hostname Generator", func() {
	var stopChSend chan<- struct{}
	var resManager manager.ResourceManager

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		resManager = manager.NewResourceManager(memory.NewStore())
		allocator, err := hostname.NewGenerator(
			logr.Discard(), m, resManager, 50*time.Millisecond,
			[]hostname.HostnameGenerator{meshservice_hostname.NewMeshServiceHostnameGenerator(resManager)},
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
			WithName("backend").
			WithTemplate("{{ .Name }}.mesh").
			WithMeshServiceMatchLabels(map[string]string{"label": "value"}).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		err = builders.HostnameGenerator().
			WithName("static").
			WithTemplate("static.mesh").
			WithMeshServiceMatchLabels(map[string]string{"generate": "static"}).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		close(stopChSend)
		Expect(resManager.Delete(context.Background(), &meshservice_api.MeshServiceResource{}, store.DeleteByKey("backend", "default"))).To(Succeed())
	})

	meshServiceStatus := func(name string) *meshservice_api.MeshServiceStatus {
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
			status := meshServiceStatus("backend")
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
			status := meshServiceStatus("backend")
			g.Expect(status.Addresses).Should(Not(BeEmpty()))
			g.Expect(status.HostnameGenerators).Should(Not(BeEmpty()))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should set an error if there's a collision", func() {
		// when
		Expect(
			samples.MeshServiceBackendBuilder().WithoutVIP().WithLabels(map[string]string{
				"generate": "static",
			}).Create(resManager),
		).To(Succeed())
		Expect(
			samples.MeshServiceBackendBuilder().WithoutVIP().WithLabels(map[string]string{
				"generate": "static",
			}).WithName("other").Create(resManager),
		).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			otherStatus := meshServiceStatus("other")
			backendStatus := meshServiceStatus("backend")
			g.Expect(otherStatus.Addresses).Should(BeEmpty())
			g.Expect(otherStatus.HostnameGenerators).Should(ConsistOf(
				hostnamegenerator_api.HostnameGeneratorStatus{
					HostnameGeneratorRef: hostnamegenerator_api.HostnameGeneratorRef{CoreName: "static"},
					Conditions: []hostnamegenerator_api.Condition{{
						Type:    hostnamegenerator_api.GeneratedCondition,
						Status:  kube_meta.ConditionFalse,
						Reason:  hostnamegenerator_api.CollisionReason,
						Message: "Hostname collision with MeshService: other",
					}},
				},
			))
			g.Expect(backendStatus.Addresses).Should(Not(BeEmpty()))
			g.Expect(backendStatus.HostnameGenerators).Should(ConsistOf(
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

	It("should not generate hostname when selector is not MeshService", func() {
		// when
		err := builders.HostnameGenerator().
			WithName("mes-generator").
			WithTemplate("{{ .DisplayName }}.mesh").
			WithMeshExternalServiceMatchLabels(map[string]string{"test": "true"}).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		err = samples.MeshServiceBackendBuilder().
			WithLabels(map[string]string{"test": "true"}).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			status := meshServiceStatus("backend")
			g.Expect(status.Addresses).Should(BeEmpty())
			g.Expect(status.HostnameGenerators).Should(BeEmpty())
		}, "2s", "100ms").Should(Succeed())
	})
})
