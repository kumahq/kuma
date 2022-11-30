package system

import (
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type secretResource struct {
	Meta model.ResourceMeta
	Spec string
}

func (l *SecretResourceList) MarshalLog() interface{} {
	list := make([]secretResource, len(l.Items))
	for _, res := range l.Items {
		list = append(list, secretResource{
			Meta: res.GetMeta(),
			Spec: "***",
		})
	}
	return list
}

func (l *GlobalSecretResourceList) MarshalLog() interface{} {
	list := make([]secretResource, len(l.Items))
	for _, res := range l.Items {
		list = append(list, secretResource{
			Meta: res.GetMeta(),
			Spec: "***",
		})
	}
	return list
}
