package mesh

import (
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

func (es *ExternalServiceResource) IsReachableFromZone(zone string) bool {
	return es.Spec.Tags[mesh_proto.ZoneTag] == "" || es.Spec.Tags[mesh_proto.ZoneTag] == zone
}

func (esl *ExternalServiceResourceList) MarshalLog() interface{} {
	maskedList := make([]*ExternalServiceResource, 0, len(esl.Items))
	for _, es := range esl.Items {
		maskedList = append(maskedList, es.MarshalLog().(*ExternalServiceResource))
	}
	return ExternalServiceResourceList{
		Items:      maskedList,
		Pagination: esl.Pagination,
	}
}

func (es *ExternalServiceResource) MarshalLog() interface{} {
	spec := proto.Clone(es.Spec).(*mesh_proto.ExternalService)
	if spec == nil {
		return es
	}
	net := spec.GetNetworking()
	if net == nil {
		return es
	}
	tls := net.GetTls()
	if tls == nil {
		return es
	}
	tls.CaCert = tls.CaCert.MaskInlineDatasource()
	tls.ClientCert = tls.ClientCert.MaskInlineDatasource()
	tls.ClientKey = tls.ClientKey.MaskInlineDatasource()
	return &ExternalServiceResource{
		Meta: es.Meta,
		Spec: spec,
	}
}
