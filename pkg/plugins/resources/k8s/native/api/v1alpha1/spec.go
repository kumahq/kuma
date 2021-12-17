package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

// ToSpec marshals a protobuf message into a Kubernetes JSON compatible format.
func ToSpec(p proto.Message) *apiextensionsv1.JSON {
	return &apiextensionsv1.JSON{
		Raw: util_proto.MustMarshalJSON(p),
	}
}
