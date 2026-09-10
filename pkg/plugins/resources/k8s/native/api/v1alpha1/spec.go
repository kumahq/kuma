package v1alpha1

import (
	"encoding/json"

	"google.golang.org/protobuf/proto"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

// ToSpec marshals a protobuf message into a Kubernetes JSON compatible format.
// ToSpec takes either form a resource spec comes in, since the conversion away from
// protobuf leaves the two side by side.
func ToSpec(p any) *apiextensionsv1.JSON {
	if msg, ok := p.(proto.Message); ok {
		return &apiextensionsv1.JSON{Raw: util_proto.MustMarshalJSON(msg)}
	}
	raw, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return &apiextensionsv1.JSON{Raw: raw}
}
