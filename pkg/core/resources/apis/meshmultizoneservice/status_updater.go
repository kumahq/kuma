package meshmultizoneservice

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	meshmzservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	util_time "github.com/kumahq/kuma/v2/pkg/util/time"
)

type StatusUpdater struct {
	roResManager manager.ReadOnlyResourceManager
	resManager   manager.ResourceManager
	logger       logr.Logger
	metric       prometheus.Histogram
	interval     time.Duration
}

var _ component.Component = &StatusUpdater{}

func NewStatusUpdater(
	logger logr.Logger,
	roResManager manager.ReadOnlyResourceManager,
	resManager manager.ResourceManager,
	interval time.Duration,
	metrics core_metrics.Metrics,
) (component.Component, error) {
	metric := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "component_mzms_status_updater",
		Help: "Summary of MeshMultizoneService Updater component",
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &StatusUpdater{
		roResManager: roResManager,
		resManager:   resManager,
		logger:       logger,
		metric:       metric,
		interval:     interval,
	}, nil
}

func (s *StatusUpdater) Start(stop <-chan struct{}) error {
	// sleep to mitigate update conflicts with other components
	util_time.SleepUpTo(s.interval)
	s.logger.Info("starting")
	ticker := time.NewTicker(s.interval)
	ctx := user.Ctx(context.Background(), user.ControlPlane)

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			if err := s.updateStatus(ctx); err != nil {
				s.logger.Error(err, "could not update status of mesh multizone services")
			}
			s.metric.Observe(float64(time.Since(start).Milliseconds()))
		case <-stop:
			s.logger.Info("stopping")
			return nil
		}
	}
}

func (s *StatusUpdater) updateStatus(ctx context.Context) error {
	mzSvcList := &meshmzservice_api.MeshMultiZoneServiceResourceList{}
	if err := s.roResManager.List(ctx, mzSvcList); err != nil {
		return errors.Wrap(err, "could not list of MeshMultiZoneServices")
	}
	if len(mzSvcList.Items) == 0 {
		// skip fetching other resources if MeshMultiZoneService is not used
		return nil
	}

	msList := &meshservice_api.MeshServiceResourceList{}
	if err := s.roResManager.List(ctx, msList); err != nil {
		return errors.Wrap(err, "could not list of MeshServices")
	}

	for _, mzSvc := range mzSvcList.Items {
		var matched []meshmzservice_api.MatchedMeshService
		ports := map[int32]meshservice_api.Port{}
		for _, svc := range msList.Items {
			if matchesService(mzSvc, svc) {
				ri := kri.From(svc)
				matched = append(matched, meshmzservice_api.MatchedMeshService{
					Name:      ri.Name,
					Namespace: ri.Namespace,
					Zone:      ri.Zone,
					Mesh:      ri.Mesh,
				})
				for _, port := range svc.Spec.Ports {
					ports[port.Port] = port
				}
			}
		}

		sort.Slice(matched, func(i, j int) bool {
			return matched[i].FullyQualifiedName() < matched[j].FullyQualifiedName()
		})

		condition := buildMeshServicesMatchedCondition(matched)
		statusChanged := !reflect.DeepEqual(mzSvc.Status.MeshServices, matched) ||
			!conditionEquals(mzSvc.Status.Conditions, condition)

		if statusChanged {
			log := s.logger.WithValues("meshmultizoneservice", mzSvc.Meta.GetName())
			mzSvc.Status.MeshServices = matched
			mzSvc.Status.Conditions = updateConditions(mzSvc.Status.Conditions, condition)
			log.V(1).Info("updating matched mesh services", "matchedMeshServices", matched, "condition", condition.Type)
			if err := s.resManager.Update(ctx, mzSvc); err != nil {
				if store.IsConflict(err) {
					log.Info("couldn't update mesh multi zone service, because it was modified in another place. Will try again in the next interval", "interval", s.interval)
				} else {
					log.Error(err, "could not update matched mesh services mesh services")
				}
			}
		}
	}
	return nil
}

func matchesService(mzSvc *meshmzservice_api.MeshMultiZoneServiceResource, svc *meshservice_api.MeshServiceResource) bool {
	for label, value := range pointer.Deref(mzSvc.Spec.Selector.MeshService.MatchLabels) {
		if svc.Meta.GetLabels()[label] != value {
			return false
		}
	}
	return true
}

func (s *StatusUpdater) NeedLeaderElection() bool {
	return true
}

func buildMeshServicesMatchedCondition(matched []meshmzservice_api.MatchedMeshService) common_api.Condition {
	if len(matched) == 0 {
		return common_api.Condition{
			Type:    meshmzservice_api.MeshServicesMatchedCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  meshmzservice_api.NoMatchesFoundReason,
			Message: "Selector does not match any MeshService in this zone",
		}
	}
	return common_api.Condition{
		Type:    meshmzservice_api.MeshServicesMatchedCondition,
		Status:  kube_meta.ConditionTrue,
		Reason:  meshmzservice_api.MatchesFoundReason,
		Message: fmt.Sprintf("Matched %d MeshService(s)", len(matched)),
	}
}

func conditionEquals(conditions []common_api.Condition, newCondition common_api.Condition) bool {
	for _, c := range conditions {
		if c.Type == newCondition.Type {
			return c.Status == newCondition.Status &&
				c.Reason == newCondition.Reason &&
				c.Message == newCondition.Message
		}
	}
	return false
}

func updateConditions(conditions []common_api.Condition, newCondition common_api.Condition) []common_api.Condition {
	for i, c := range conditions {
		if c.Type == newCondition.Type {
			conditions[i] = newCondition
			return conditions
		}
	}
	return append(conditions, newCondition)
}
