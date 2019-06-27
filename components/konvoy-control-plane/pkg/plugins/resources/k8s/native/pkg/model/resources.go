package model

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type KubernetesObject interface {
	runtime.Object
	GetObjectMeta() *metav1.ObjectMeta
	SetObjectMeta(*metav1.ObjectMeta)
	GetSpec() map[string]interface{}
	SetSpec(map[string]interface{})
}

type KubernetesList interface {
	runtime.Object
	GetItems() []KubernetesObject
}
