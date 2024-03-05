package k8s

import (
	"fmt"

	"google.golang.org/protobuf/types/known/wrapperspb"
	v1 "k8s.io/api/core/v1"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

// Secret is a KubernetesObject for Kuma's Secret and GlobalSecret.
// Note that it's not registered in TypeRegistry because we cannot multiply KubernetesObject
// for a single Spec (both Secret and GlobalSecret has same Spec).
type Secret struct {
	v1.Secret
}

func NewSecret(typ v1.SecretType) *Secret {
	return &Secret{
		Secret: v1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1.SchemeGroupVersion.String(),
				Kind:       "Secret",
			},
			Type: typ,
		},
	}
}

var _ model.KubernetesObject = &Secret{}

func (s *Secret) GetObjectMeta() *k8s.ObjectMeta {
	return &s.ObjectMeta
}

func (s *Secret) SetObjectMeta(meta *k8s.ObjectMeta) {
	s.ObjectMeta = *meta
}

func (s *Secret) GetMesh() string {
	if mesh, ok := s.ObjectMeta.Labels[metadata.KumaMeshLabel]; ok {
		return mesh
	} else {
		return core_model.DefaultMesh
	}
}

func (s *Secret) SetMesh(mesh string) {
	if s.ObjectMeta.Labels == nil {
		s.ObjectMeta.Labels = map[string]string{}
	}
	s.ObjectMeta.Labels[metadata.KumaMeshLabel] = mesh
}

func (s *Secret) GetSpec() (core_model.ResourceSpec, error) {
	bytes, ok := s.Data["value"]
	if !ok {
		return nil, nil
	}
	return &system_proto.Secret{
		Data: &wrapperspb.BytesValue{
			Value: bytes,
		},
	}, nil
}

func (s *Secret) SetSpec(spec core_model.ResourceSpec) {
	if _, ok := spec.(*system_proto.Secret); !ok {
		panic(fmt.Sprintf("unexpected protobuf message type %T", spec))
	}
	s.Data = map[string][]byte{
		"value": spec.(*system_proto.Secret).GetData().GetValue(),
	}
}

func (s *Secret) Scope() model.Scope {
	return model.ScopeNamespace
}
