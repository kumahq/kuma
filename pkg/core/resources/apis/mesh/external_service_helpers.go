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
		c, err := copystructure.Copy(es)
		if err != nil {
			return nil
		}
		esCopy := c.(*ExternalServiceResource)
		spec := esCopy.Spec
		if spec == nil {
			maskedList = append(maskedList, es)
			continue
		}
		net := spec.GetNetworking()
		if net == nil {
			maskedList = append(maskedList, es)
			continue
		}
		tls := net.GetTls()
		if tls == nil {
			maskedList = append(maskedList, es)
			continue
		}
		tls.CaCert = tls.CaCert.MaskInlineDatasource()
		tls.ClientCert = tls.ClientCert.MaskInlineDatasource()
		tls.ClientKey = tls.ClientKey.MaskInlineDatasource()
		maskedList = append(maskedList, esCopy)
	}
	return ExternalServiceResourceList{
		Items:      maskedList,
		Pagination: esl.Pagination,
	}
}
