package v1alpha1

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	sample_proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (pt *SampleTrafficRoute) GetObjectMeta() *metav1.ObjectMeta {
	return &pt.ObjectMeta
}

func (pt *SampleTrafficRoute) SetObjectMeta(m *metav1.ObjectMeta) {
	pt.ObjectMeta = *m
}

func (pt *SampleTrafficRoute) GetMesh() string {
	return pt.Mesh
}

func (pt *SampleTrafficRoute) SetMesh(mesh string) {
	pt.Mesh = mesh
}

func (pt *SampleTrafficRoute) GetSpec() proto.Message {
	spec := pt.Spec
	m := sample_proto.TrafficRoute{}

	if spec == nil || len(spec.Raw) == 0 {
		return &m
	}

	return util_proto.MustUnmarshalJSON(spec.Raw, &m)
}

func (pt *SampleTrafficRoute) SetSpec(spec proto.Message) {
	if spec == nil {
		pt.Spec = nil
		return
	}

	if _, ok := spec.(*sample_proto.TrafficRoute); !ok {
		panic(fmt.Sprintf("unexpected protobuf message type %T", spec))
	}

	pt.Spec = &apiextensionsv1.JSON{Raw: util_proto.MustMarshalJSON(spec)}
}

func (pt *SampleTrafficRoute) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *SampleTrafficRouteList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}
