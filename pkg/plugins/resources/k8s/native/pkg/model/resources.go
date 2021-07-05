package model

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Scope string

const (
	ScopeNamespace Scope = "namespace"
	ScopeCluster   Scope = "cluster"
)

type KubernetesObject interface {
	runtime.Object
	metav1.Object

	GetObjectMeta() *metav1.ObjectMeta
	SetObjectMeta(*metav1.ObjectMeta)
	GetMesh() string
	SetMesh(string)
	GetSpec() map[string]interface{}
	SetSpec(map[string]interface{})
	Scope() Scope
}

type KubernetesList interface {
	runtime.Object
	GetItems() []KubernetesObject
	GetContinue() string
}

// RawMessage is a carrier for an untyped JSON payload.
type RawMessage map[string]interface{}

// DeepCopy ...
func (in RawMessage) DeepCopy() RawMessage {
	return runtime.DeepCopyJSON(in)
}
