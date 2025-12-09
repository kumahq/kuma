package meshmultizoneservice_test

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshmultizoneservice"
	meshmzservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/v2/pkg/test/metrics"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
)

var _ = Describe("Updater", func() {
	var ch chan struct{}
	var updaterDone *sync.WaitGroup
	var resManager manager.ResourceManager
	var metrics core_metrics.Metrics

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		metrics = m
		resManager = manager.NewResourceManager(memory.NewStore())
		updaterDone = &sync.WaitGroup{}

		updater, err := meshmultizoneservice.NewStatusUpdater(logr.Discard(), resManager, resManager, 50*time.Millisecond, m)
		Expect(err).ToNot(HaveOccurred())
		ch = make(chan struct{})
		updaterDone.Add(1)
		go func() {
			defer GinkgoRecover()
			defer updaterDone.Done()
			Expect(updater.Start(ch)).To(Succeed())
		}()

		Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
	})

	AfterEach(func() {
		if ch != nil {
			close(ch)
		}
		if updaterDone != nil {
			updaterDone.Wait()
		}
	})

	matchedMeshServices := func() ([]meshmzservice_api.MatchedMeshService, error) {
		mzsvc := meshmzservice_api.NewMeshMultiZoneServiceResource()
		err := resManager.Get(context.Background(), mzsvc, store.GetByKey("backend", model.DefaultMesh))
		if err != nil {
			return nil, err
		}
		return mzsvc.Status.MeshServices, nil
	}

	getCondition := func(conditionType string) (string, error) {
		mzsvc := meshmzservice_api.NewMeshMultiZoneServiceResource()
		err := resManager.Get(context.Background(), mzsvc, store.GetByKey("backend", model.DefaultMesh))
		if err != nil {
			return "", err
		}
		for _, c := range mzsvc.Status.Conditions {
			if c.Type == conditionType {
				return string(c.Status), nil
			}
		}
		return "", nil
	}

	ms1Builder := samples.MeshServiceBackendBuilder().
		WithName("backend").
		WithDataplaneTagsSelectorKV("app", "backend").
		WithLabels(map[string]string{
			mesh_proto.DisplayName: "backend",
			mesh_proto.ZoneTag:     "east",
		})

	ms2Builder := samples.MeshServiceBackendBuilder().
		WithName("backend-syncedhash").
		WithDataplaneTagsSelectorKV("app", "backend").
		WithLabels(map[string]string{
			mesh_proto.DisplayName: "backend",
			mesh_proto.ZoneTag:     "west",
		})

	It("should add mesh services to the status of multizone service", func() {
		// when
		Expect(ms1Builder.Create(resManager)).To(Succeed())
		Expect(samples.MeshServiceWebBuilder().Create(resManager)).To(Succeed()) // to check if we ignore it
		Expect(samples.MeshMultiZoneServiceBackendBuilder().Create(resManager)).To(Succeed())

		// then
		Eventually(matchedMeshServices, "10s", "100ms").Should(Equal([]meshmzservice_api.MatchedMeshService{
			{Name: "backend", Zone: "east", Mesh: "default"},
		}))

		// when new service is added
		Expect(ms2Builder.Create(resManager)).To(Succeed())

		// then
		Eventually(matchedMeshServices, "10s", "100ms").Should(Equal([]meshmzservice_api.MatchedMeshService{
			{Name: "backend", Zone: "east", Mesh: "default"},
			{Name: "backend", Zone: "west", Mesh: "default"},
		}))
	})

	It("should result in the same list when services are added in a different order", func() {
		// when
		Expect(ms2Builder.Create(resManager)).To(Succeed())
		Expect(ms1Builder.Create(resManager)).To(Succeed())
		Expect(samples.MeshMultiZoneServiceBackendBuilder().Create(resManager)).To(Succeed())

		// then
		Eventually(matchedMeshServices, "10s", "100ms").Should(Equal([]meshmzservice_api.MatchedMeshService{
			{Name: "backend", Zone: "east", Mesh: "default"},
			{Name: "backend", Zone: "west", Mesh: "default"},
		}))
	})

	It("should emit metric", func() {
		Eventually(func(g Gomega) {
			g.Expect(test_metrics.FindMetric(metrics, "component_mzms_status_updater")).ToNot(BeNil())
		}, "10s", "100ms").Should(Succeed())
	})

	It("should set MeshServicesMatched condition to False when no matches found", func() {
		// when - create MMZS without any matching MeshServices
		Expect(samples.MeshMultiZoneServiceBackendBuilder().Create(resManager)).To(Succeed())

		// then - condition should be False
		Eventually(func(g Gomega) {
			status, err := getCondition(meshmzservice_api.MeshServicesMatchedCondition)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status).To(Equal(string(kube_meta.ConditionFalse)))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should set MeshServicesMatched condition to True when matches found", func() {
		// when - create matching MeshService
		Expect(ms1Builder.Create(resManager)).To(Succeed())
		Expect(samples.MeshMultiZoneServiceBackendBuilder().Create(resManager)).To(Succeed())

		// then - condition should be True
		Eventually(func(g Gomega) {
			status, err := getCondition(meshmzservice_api.MeshServicesMatchedCondition)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status).To(Equal(string(kube_meta.ConditionTrue)))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should change condition when matches disappear", func() {
		// given - MMZS with matching MeshService
		Expect(ms1Builder.Create(resManager)).To(Succeed())
		Expect(samples.MeshMultiZoneServiceBackendBuilder().Create(resManager)).To(Succeed())

		// then - condition should be True
		Eventually(func(g Gomega) {
			status, err := getCondition(meshmzservice_api.MeshServicesMatchedCondition)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status).To(Equal(string(kube_meta.ConditionTrue)))
		}, "10s", "100ms").Should(Succeed())

		// when - delete the matching MeshService
		ms1 := meshservice_api.NewMeshServiceResource()
		Expect(resManager.Get(context.Background(), ms1, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
		Expect(resManager.Delete(context.Background(), ms1, store.DeleteByKey("backend", model.DefaultMesh))).To(Succeed())

		// then - condition should change to False
		Eventually(func(g Gomega) {
			status, err := getCondition(meshmzservice_api.MeshServicesMatchedCondition)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status).To(Equal(string(kube_meta.ConditionFalse)))
		}, "10s", "100ms").Should(Succeed())
	})
})
