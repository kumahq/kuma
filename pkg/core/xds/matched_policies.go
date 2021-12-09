package xds

import (
	"context"
	"fmt"
	"reflect"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type MatchedPolicies struct {
	// Inbound(Listener) -> Policy
	TrafficPermissions TrafficPermissionMap
	FaultInjections    FaultInjectionMap
	RateLimitsInbound  InboundRateLimitsMap

	// Service(Cluster) -> Policy
	TrafficLogs     TrafficLogMap
	HealthChecks    HealthCheckMap
	CircuitBreakers CircuitBreakerMap
	Retries         RetryMap

	// Outbound(Listener) -> Policy
	Timeouts           TimeoutMap
	RateLimitsOutbound OutboundRateLimitsMap

	// Dataplane -> Policy
	TrafficTrace *core_mesh.TrafficTraceResource
}

type MatchedPoliciesGetter interface {
	Get(ctx context.Context, dataplaneKey core_model.ResourceKey) (*MatchedPolicies, error)
}

type AttachmentType int64

const (
	Inbound AttachmentType = iota
	Outbound
	Service
	Dataplane
)

func (s AttachmentType) String() string {
	switch s {
	case Inbound:
		return "inbound"
	case Outbound:
		return "outbound"
	case Service:
		return "service"
	case Dataplane:
		return "dataplane"
	default:
		return "unknown"
	}
}

type Attachment struct {
	Type AttachmentType
	Name string
}

type AttachmentList []Attachment

func (a AttachmentList) Len() int           { return len(a) }
func (a AttachmentList) Less(i, j int) bool { return fmt.Sprintf("%s", a[i]) < fmt.Sprintf("%s", a[j]) }
func (a AttachmentList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type PolicyMap map[core_model.ResourceType][]core_model.Resource
type AttachmentMap map[Attachment]PolicyMap

func (m AttachmentMap) store(key interface{}, resources []core_model.Resource) {
	if len(resources) == 0 {
		return
	}
	resType := resources[0].Descriptor().Name

	attachment := Attachment{}
	if key == nil {
		attachment.Type = Dataplane
	} else {
		switch key.(type) {
		case mesh_proto.InboundInterface:
			attachment.Type = Inbound
		case mesh_proto.OutboundInterface:
			attachment.Type = Outbound
		case string:
			attachment.Type = Service
		}
		attachment.Name = fmt.Sprintf("%s", key)
	}

	if _, ok := m[attachment]; !ok {
		m[attachment] = PolicyMap{}
	}
	m[attachment][resType] = resources
}

func GroupByAttachment(matchedPolicies *MatchedPolicies) AttachmentMap {
	result := AttachmentMap{}

	processField := func(f reflect.Value) {
		if f.IsNil() {
			return
		}

		switch f.Kind() {
		case reflect.Map:
			for _, mapKey := range f.MapKeys() {
				mapValue := f.MapIndex(mapKey)

				key := mapKey.Interface()
				resources := []core_model.Resource{}
				if mapValue.Kind() == reflect.Slice {
					for i := 0; i < mapValue.Len(); i++ {
						resources = append(resources, mapValue.Index(i).Interface().(core_model.Resource))
					}
				} else {
					resources = []core_model.Resource{mapValue.Interface().(core_model.Resource)}
				}

				result.store(key, resources)
			}
		case reflect.Ptr:
			var key interface{}
			resources := []core_model.Resource{f.Interface().(core_model.Resource)}

			result.store(key, resources)
		}
	}

	matchedPoliciesValue := reflect.ValueOf(matchedPolicies).Elem()
	for i := 0; i < matchedPoliciesValue.NumField(); i++ {
		processField(matchedPoliciesValue.Field(i))
	}

	return result
}
