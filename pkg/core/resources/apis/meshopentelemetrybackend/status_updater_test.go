package meshopentelemetrybackend_test

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	meshaccesslog_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshmetric_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/api/v1alpha1"
	meshtrace_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/v2/pkg/test/metrics"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
)

var _ = Describe("StatusUpdater", func() {
	var stopCh chan struct{}
	var resManager manager.ResourceManager
	var metrics core_metrics.Metrics

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		metrics = m
		resManager = manager.NewResourceManager(memory.NewStore())

		updater, err := meshopentelemetrybackend.NewStatusUpdater(logr.Discard(), resManager, resManager, 50*time.Millisecond, m)
		Expect(err).ToNot(HaveOccurred())
		stopCh = make(chan struct{})
		go func(stopCh chan struct{}) {
			defer GinkgoRecover()
			Expect(updater.Start(stopCh)).To(Succeed())
		}(stopCh)

		Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
	})

	AfterEach(func() {
		close(stopCh)
	})

	createMOTB := func(name string) {
		motb := motb_api.NewMeshOpenTelemetryBackendResource()
		motb.Spec = &motb_api.MeshOpenTelemetryBackend{
			Endpoint: &motb_api.Endpoint{
				Address: "otel-collector.observability",
				Port:    4317,
			},
			Protocol: motb_api.ProtocolGRPC,
		}
		Expect(resManager.Create(context.Background(), motb, store.CreateByKey(name, model.DefaultMesh))).To(Succeed())
	}

	getConditions := func(name string) func() ([]common_api.Condition, error) {
		return func() ([]common_api.Condition, error) {
			motb := motb_api.NewMeshOpenTelemetryBackendResource()
			err := resManager.Get(context.Background(), motb, store.GetByKey(name, model.DefaultMesh))
			if err != nil {
				return nil, err
			}
			return motb.Status.Conditions, nil
		}
	}

	getMeshMetricConditions := func(name string) func() ([]common_api.Condition, error) {
		return func() ([]common_api.Condition, error) {
			mm := meshmetric_api.NewMeshMetricResource()
			err := resManager.Get(context.Background(), mm, store.GetByKey(name, model.DefaultMesh))
			if err != nil {
				return nil, err
			}
			if mm.Status == nil {
				return nil, nil
			}
			return mm.Status.Conditions, nil
		}
	}

	getMeshTraceConditions := func(name string) func() ([]common_api.Condition, error) {
		return func() ([]common_api.Condition, error) {
			mt := meshtrace_api.NewMeshTraceResource()
			err := resManager.Get(context.Background(), mt, store.GetByKey(name, model.DefaultMesh))
			if err != nil {
				return nil, err
			}
			if mt.Status == nil {
				return nil, nil
			}
			return mt.Status.Conditions, nil
		}
	}

	getMeshAccessLogConditions := func(name string) func() ([]common_api.Condition, error) {
		return func() ([]common_api.Condition, error) {
			mal := meshaccesslog_api.NewMeshAccessLogResource()
			err := resManager.Get(context.Background(), mal, store.GetByKey(name, model.DefaultMesh))
			if err != nil {
				return nil, err
			}
			if mal.Status == nil {
				return nil, nil
			}
			return mal.Status.Conditions, nil
		}
	}

	It("should set NotReferenced condition when no policies reference the backend", func() {
		createMOTB("main-collector")

		Eventually(getConditions("main-collector"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    motb_api.ReferencedByPoliciesCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  motb_api.NotReferencedReason,
			Message: "Not referenced by any observability policy",
		}))
	})

	It("should set Referenced condition when MeshMetric references the backend", func() {
		createMOTB("main-collector")

		mm := meshmetric_api.NewMeshMetricResource()
		mm.Spec = &meshmetric_api.MeshMetric{
			Default: meshmetric_api.Conf{
				Backends: &[]meshmetric_api.Backend{
					{
						Type: meshmetric_api.OpenTelemetryBackendType,
						OpenTelemetry: &meshmetric_api.OpenTelemetryBackend{
							BackendRef: &common_api.BackendResourceRef{
								Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
								Name: "main-collector",
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), mm, store.CreateByKey("mm-1", model.DefaultMesh))).To(Succeed())

		Eventually(getConditions("main-collector"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    motb_api.ReferencedByPoliciesCondition,
			Status:  kube_meta.ConditionTrue,
			Reason:  motb_api.ReferencedReason,
			Message: "Referenced by 1 policy backend(s)",
		}))
	})

	It("should set Referenced condition when MeshTrace references the backend", func() {
		createMOTB("trace-collector")

		mt := meshtrace_api.NewMeshTraceResource()
		mt.Spec = &meshtrace_api.MeshTrace{
			Default: meshtrace_api.Conf{
				Backends: &[]meshtrace_api.Backend{
					{
						Type: meshtrace_api.OpenTelemetryBackendType,
						OpenTelemetry: &meshtrace_api.OpenTelemetryBackend{
							BackendRef: &common_api.BackendResourceRef{
								Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
								Name: "trace-collector",
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), mt, store.CreateByKey("mt-1", model.DefaultMesh))).To(Succeed())

		Eventually(getConditions("trace-collector"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    motb_api.ReferencedByPoliciesCondition,
			Status:  kube_meta.ConditionTrue,
			Reason:  motb_api.ReferencedReason,
			Message: "Referenced by 1 policy backend(s)",
		}))
	})

	It("should set Referenced condition when MeshAccessLog references the backend", func() {
		createMOTB("log-collector")

		mal := meshaccesslog_api.NewMeshAccessLogResource()
		mal.Spec = &meshaccesslog_api.MeshAccessLog{
			To: &[]meshaccesslog_api.To{
				{
					TargetRef: common_api.TargetRef{Kind: "Mesh"},
					Default: meshaccesslog_api.Conf{
						Backends: &[]meshaccesslog_api.Backend{
							{
								Type: meshaccesslog_api.OtelTelemetryBackendType,
								OpenTelemetry: &meshaccesslog_api.OtelBackend{
									BackendRef: &common_api.BackendResourceRef{
										Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
										Name: "log-collector",
									},
								},
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), mal, store.CreateByKey("mal-1", model.DefaultMesh))).To(Succeed())

		Eventually(getConditions("log-collector"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    motb_api.ReferencedByPoliciesCondition,
			Status:  kube_meta.ConditionTrue,
			Reason:  motb_api.ReferencedReason,
			Message: "Referenced by 1 policy backend(s)",
		}))
	})

	It("should count multiple policy references", func() {
		createMOTB("shared-collector")

		// MeshMetric referencing shared-collector
		mm := meshmetric_api.NewMeshMetricResource()
		mm.Spec = &meshmetric_api.MeshMetric{
			Default: meshmetric_api.Conf{
				Backends: &[]meshmetric_api.Backend{
					{
						Type: meshmetric_api.OpenTelemetryBackendType,
						OpenTelemetry: &meshmetric_api.OpenTelemetryBackend{
							BackendRef: &common_api.BackendResourceRef{
								Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
								Name: "shared-collector",
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), mm, store.CreateByKey("mm-1", model.DefaultMesh))).To(Succeed())

		// MeshTrace referencing shared-collector
		mt := meshtrace_api.NewMeshTraceResource()
		mt.Spec = &meshtrace_api.MeshTrace{
			Default: meshtrace_api.Conf{
				Backends: &[]meshtrace_api.Backend{
					{
						Type: meshtrace_api.OpenTelemetryBackendType,
						OpenTelemetry: &meshtrace_api.OpenTelemetryBackend{
							BackendRef: &common_api.BackendResourceRef{
								Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
								Name: "shared-collector",
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), mt, store.CreateByKey("mt-1", model.DefaultMesh))).To(Succeed())

		Eventually(getConditions("shared-collector"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    motb_api.ReferencedByPoliciesCondition,
			Status:  kube_meta.ConditionTrue,
			Reason:  motb_api.ReferencedReason,
			Message: "Referenced by 2 policy backend(s)",
		}))
	})

	It("should change condition when references are removed", func() {
		createMOTB("main-collector")

		mm := meshmetric_api.NewMeshMetricResource()
		mm.Spec = &meshmetric_api.MeshMetric{
			Default: meshmetric_api.Conf{
				Backends: &[]meshmetric_api.Backend{
					{
						Type: meshmetric_api.OpenTelemetryBackendType,
						OpenTelemetry: &meshmetric_api.OpenTelemetryBackend{
							BackendRef: &common_api.BackendResourceRef{
								Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
								Name: "main-collector",
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), mm, store.CreateByKey("mm-1", model.DefaultMesh))).To(Succeed())

		// wait for Referenced condition
		Eventually(getConditions("main-collector"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    motb_api.ReferencedByPoliciesCondition,
			Status:  kube_meta.ConditionTrue,
			Reason:  motb_api.ReferencedReason,
			Message: "Referenced by 1 policy backend(s)",
		}))

		// delete the referencing policy
		Expect(resManager.Delete(context.Background(), meshmetric_api.NewMeshMetricResource(), store.DeleteByKey("mm-1", model.DefaultMesh))).To(Succeed())

		// condition should change to NotReferenced
		Eventually(getConditions("main-collector"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    motb_api.ReferencedByPoliciesCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  motb_api.NotReferencedReason,
			Message: "Not referenced by any observability policy",
		}))
	})

	It("should not count refs from a different mesh", func() {
		// MOTB is in default mesh; policy is in other-mesh with same backendRef name.
		// The status updater must not count the cross-mesh ref.
		createMOTB("main-collector")
		Expect(samples.MeshDefaultBuilder().WithName("other-mesh").Create(resManager)).To(Succeed())
		mmOtherMesh := meshmetric_api.NewMeshMetricResource()
		mmOtherMesh.Spec = &meshmetric_api.MeshMetric{
			Default: meshmetric_api.Conf{
				Backends: &[]meshmetric_api.Backend{
					{
						Type: meshmetric_api.OpenTelemetryBackendType,
						OpenTelemetry: &meshmetric_api.OpenTelemetryBackend{
							BackendRef: &common_api.BackendResourceRef{
								Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
								Name: "main-collector",
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(
			context.Background(),
			mmOtherMesh,
			store.CreateByKey("mm-other-mesh", "other-mesh"),
			store.CreateWithLabels(map[string]string{mesh_proto.MeshTag: "other-mesh"}),
		)).To(Succeed())

		// MOTB in default mesh should stay NotReferenced
		Eventually(getConditions("main-collector"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    motb_api.ReferencedByPoliciesCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  motb_api.NotReferencedReason,
			Message: "Not referenced by any observability policy",
		}))
		Consistently(getConditions("main-collector"), "1s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    motb_api.ReferencedByPoliciesCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  motb_api.NotReferencedReason,
			Message: "Not referenced by any observability policy",
		}))
	})

	It("should mark MeshMetric backendRefs unresolved when MOTB is missing", func() {
		mm := meshmetric_api.NewMeshMetricResource()
		mm.Spec = &meshmetric_api.MeshMetric{
			Default: meshmetric_api.Conf{
				Backends: &[]meshmetric_api.Backend{
					{
						Type: meshmetric_api.OpenTelemetryBackendType,
						OpenTelemetry: &meshmetric_api.OpenTelemetryBackend{
							BackendRef: &common_api.BackendResourceRef{
								Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
								Name: "missing-collector",
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), mm, store.CreateByKey("mm-missing", model.DefaultMesh))).To(Succeed())

		Eventually(getMeshMetricConditions("mm-missing"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    meshmetric_api.BackendRefsResolvedCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  meshmetric_api.UnresolvedBackendRefsReason,
			Message: "Unresolved MeshOpenTelemetryBackend references: missing-collector",
		}))
	})

	It("should mark MeshTrace backendRefs unresolved when MOTB is missing", func() {
		mt := meshtrace_api.NewMeshTraceResource()
		mt.Spec = &meshtrace_api.MeshTrace{
			Default: meshtrace_api.Conf{
				Backends: &[]meshtrace_api.Backend{
					{
						Type: meshtrace_api.OpenTelemetryBackendType,
						OpenTelemetry: &meshtrace_api.OpenTelemetryBackend{
							BackendRef: &common_api.BackendResourceRef{
								Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
								Name: "missing-trace-collector",
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), mt, store.CreateByKey("mt-missing", model.DefaultMesh))).To(Succeed())

		Eventually(getMeshTraceConditions("mt-missing"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    meshtrace_api.BackendRefsResolvedCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  meshtrace_api.UnresolvedBackendRefsReason,
			Message: "Unresolved MeshOpenTelemetryBackend references: missing-trace-collector",
		}))
	})

	It("should mark MeshAccessLog backendRefs unresolved when MOTB is missing", func() {
		mal := meshaccesslog_api.NewMeshAccessLogResource()
		mal.Spec = &meshaccesslog_api.MeshAccessLog{
			To: &[]meshaccesslog_api.To{
				{
					TargetRef: common_api.TargetRef{Kind: "Mesh"},
					Default: meshaccesslog_api.Conf{
						Backends: &[]meshaccesslog_api.Backend{
							{
								Type: meshaccesslog_api.OtelTelemetryBackendType,
								OpenTelemetry: &meshaccesslog_api.OtelBackend{
									BackendRef: &common_api.BackendResourceRef{
										Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
										Name: "missing-log-collector",
									},
								},
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), mal, store.CreateByKey("mal-missing", model.DefaultMesh))).To(Succeed())

		Eventually(getMeshAccessLogConditions("mal-missing"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    meshaccesslog_api.BackendRefsResolvedCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  meshaccesslog_api.UnresolvedBackendRefsReason,
			Message: "Unresolved MeshOpenTelemetryBackend references: missing-log-collector",
		}))
	})

	It("should flip MeshMetric backendRefs condition to resolved when backend appears", func() {
		mm := meshmetric_api.NewMeshMetricResource()
		mm.Spec = &meshmetric_api.MeshMetric{
			Default: meshmetric_api.Conf{
				Backends: &[]meshmetric_api.Backend{
					{
						Type: meshmetric_api.OpenTelemetryBackendType,
						OpenTelemetry: &meshmetric_api.OpenTelemetryBackend{
							BackendRef: &common_api.BackendResourceRef{
								Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
								Name: "appearing-collector",
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), mm, store.CreateByKey("mm-appearing", model.DefaultMesh))).To(Succeed())

		Eventually(getMeshMetricConditions("mm-appearing"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    meshmetric_api.BackendRefsResolvedCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  meshmetric_api.UnresolvedBackendRefsReason,
			Message: "Unresolved MeshOpenTelemetryBackend references: appearing-collector",
		}))

		createMOTB("appearing-collector")

		Eventually(getMeshMetricConditions("mm-appearing"), "10s", "100ms").Should(ContainElement(common_api.Condition{
			Type:    meshmetric_api.BackendRefsResolvedCondition,
			Status:  kube_meta.ConditionTrue,
			Reason:  meshmetric_api.AllBackendRefsResolvedReason,
			Message: "All MeshOpenTelemetryBackend references are resolved",
		}))
	})

	It("should emit metric", func() {
		Eventually(func(g Gomega) {
			g.Expect(test_metrics.FindMetric(metrics, "component_motb_status_updater")).ToNot(BeNil())
		}, "10s", "100ms").Should(Succeed())
	})
})
