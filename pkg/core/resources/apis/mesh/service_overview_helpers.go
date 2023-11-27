package mesh

import mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

func (t *ServiceOverviewResource) GetStatus() Status {
	switch t.Spec.Status {
	case mesh_proto.ServiceInsight_Service_partially_degraded:
		return PartiallyDegraded
	case mesh_proto.ServiceInsight_Service_online:
		return Online
	case mesh_proto.ServiceInsight_Service_offline:
		fallthrough
	default:
		return Offline
	}
}

func MapProtocol(protocol mesh_proto.ServiceInsight_Service_Protocol) Protocol {
	switch protocol {
	case mesh_proto.ServiceInsight_Service_grpc:
		return ProtocolGRPC
	case mesh_proto.ServiceInsight_Service_http:
		return ProtocolHTTP
	case mesh_proto.ServiceInsight_Service_http2:
		return ProtocolHTTP2
	case mesh_proto.ServiceInsight_Service_tcp:
		return ProtocolTCP
	case mesh_proto.ServiceInsight_Service_kafka:
		return ProtocolKafka
	default:
		return ProtocolUnknown
	}
}

func MapProtocolProto(protocol Protocol) mesh_proto.ServiceInsight_Service_Protocol {
	switch protocol {
	case ProtocolGRPC:
		return mesh_proto.ServiceInsight_Service_grpc
	case ProtocolHTTP:
		return mesh_proto.ServiceInsight_Service_http
	case ProtocolHTTP2:
		return mesh_proto.ServiceInsight_Service_http2
	case ProtocolTCP:
		return mesh_proto.ServiceInsight_Service_tcp
	case ProtocolKafka:
		return mesh_proto.ServiceInsight_Service_kafka
	default:
		return mesh_proto.ServiceInsight_Service_unknown
	}
}
