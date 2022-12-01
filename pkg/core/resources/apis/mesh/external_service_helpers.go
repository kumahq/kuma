package mesh

import (
	"github.com/mitchellh/copystructure"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

func (es *ExternalServiceResource) IsReachableFromZone(zone string) bool {
	return es.Spec.Tags[mesh_proto.ZoneTag] == "" || es.Spec.Tags[mesh_proto.ZoneTag] == zone
}

func (esl *ExternalServiceResourceList) MarshalLog() interface{} {
	maskedList := make([]*ExternalServiceResource, len(esl.Items))
	for _, es := range esl.Items {
		maskedList = append(maskedList, es.MarshalLog().(*ExternalServiceResource))
	}
	return ExternalServiceResourceList{
		Items:      maskedList,
		Pagination: esl.Pagination,
	}
}

func (es *ExternalServiceResource) MarshalLog() interface{} {
	c, err := copystructure.Copy(es)
	if err != nil {
		return nil
	}
	esCopy := c.(*ExternalServiceResource)
	spec := esCopy.Spec
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
	return esCopy
}
