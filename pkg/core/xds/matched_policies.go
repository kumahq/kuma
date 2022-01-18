package xds

import (
	"context"
	"fmt"

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

func GroupByAttachment(matchedPolicies *MatchedPolicies) AttachmentMap {
	result := AttachmentMap{}

	addPolicies := func(key Attachment, policies []core_model.Resource) {
		if len(policies) == 0 {
			return
		}
		if _, ok := result[key]; !ok {
			result[key] = PolicyMap{}
		}
		for _, policy := range policies {
			resType := policy.Descriptor().Name
			result[key][resType] = append(result[key][resType], policy)
		}
	}

	for inbound, policies := range getInboundMatchedPolicies(matchedPolicies) {
		addPolicies(Attachment{Type: Inbound, Name: inbound.String()}, policies)
	}

	for outbound, policies := range getOutboundMatchedPolicies(matchedPolicies) {
		addPolicies(Attachment{Type: Outbound, Name: outbound.String()}, policies)
	}

	for service, policies := range getServiceMatchedPolicies(matchedPolicies) {
		addPolicies(Attachment{Type: Service, Name: service}, policies)
	}

	addPolicies(Attachment{Type: Dataplane, Name: ""}, getDataplaneMatchedPolicies(matchedPolicies))

	return result
}

func getInboundMatchedPolicies(matchedPolicies *MatchedPolicies) map[mesh_proto.InboundInterface][]core_model.Resource {
	result := map[mesh_proto.InboundInterface][]core_model.Resource{}

	for inbound, tp := range matchedPolicies.TrafficPermissions {
		result[inbound] = append(result[inbound], tp)
	}
	for inbound, fiList := range matchedPolicies.FaultInjections {
		for _, fi := range fiList {
			result[inbound] = append(result[inbound], fi)
		}
	}
	for inbound, rlList := range matchedPolicies.RateLimitsInbound {
		for _, rl := range rlList {
			result[inbound] = append(result[inbound], rl)
		}
	}

	return result
}

func getOutboundMatchedPolicies(matchedPolicies *MatchedPolicies) map[mesh_proto.OutboundInterface][]core_model.Resource {
	result := map[mesh_proto.OutboundInterface][]core_model.Resource{}

	for outbound, timeout := range matchedPolicies.Timeouts {
		result[outbound] = append(result[outbound], timeout)
	}
	for outbound, rl := range matchedPolicies.RateLimitsOutbound {
		result[outbound] = append(result[outbound], rl)
	}

	return result
}

func getServiceMatchedPolicies(matchedPolicies *MatchedPolicies) map[ServiceName][]core_model.Resource {
	result := map[ServiceName][]core_model.Resource{}

	for service, tl := range matchedPolicies.TrafficLogs {
		result[service] = append(result[service], tl)
	}
	for service, hc := range matchedPolicies.HealthChecks {
		result[service] = append(result[service], hc)
	}
	for service, cb := range matchedPolicies.CircuitBreakers {
		result[service] = append(result[service], cb)
	}
	for service, retry := range matchedPolicies.Retries {
		result[service] = append(result[service], retry)
	}

	return result
}

func getDataplaneMatchedPolicies(matchedPolicies *MatchedPolicies) []core_model.Resource {
	if matchedPolicies.TrafficTrace != nil {
		return []core_model.Resource{matchedPolicies.TrafficTrace}
	}
	return nil
}
