package model

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type Scope string

const (
	ScopeNamespace Scope = "namespace"
	ScopeCluster   Scope = "cluster"
)

type KubernetesObject interface {
	client.Object

	GetObjectMeta() *metav1.ObjectMeta
	SetObjectMeta(*metav1.ObjectMeta)
	GetMesh() string
	SetMesh(string)
	GetSpec() (model.ResourceSpec, error)
	SetSpec(model.ResourceSpec)
	GetStatus() (model.ResourceStatus, error)
	SetStatus(status model.ResourceStatus) error
	Scope() Scope
}

type KubernetesList interface {
	client.ObjectList

	GetItems() []KubernetesObject
	GetContinue() string
}

// RawMessage is a carrier for an untyped JSON payload.
type RawMessage map[string]interface{}

// DeepCopy ...
func (in RawMessage) DeepCopy() RawMessage {
	return runtime.DeepCopyJSON(in)
}
