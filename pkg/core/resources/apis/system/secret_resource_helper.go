package system

import (
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type secretResource struct {
	Meta model.ResourceMeta
	Spec string
}

func (l *SecretResourceList) MarshalLog() interface{} {
	list := make([]interface{}, 0, len(l.Items))
	for _, res := range l.Items {
		list = append(list, res.MarshalLog())
	}
	return list
}

func (sr *SecretResource) MarshalLog() interface{} {
	return secretResource{
		Meta: sr.Meta,
		Spec: "***",
	}
}

func (l *GlobalSecretResourceList) MarshalLog() interface{} {
	list := make([]interface{}, 0, len(l.Items))
	for _, res := range l.Items {
		list = append(list, res.MarshalLog())
	}
	return list
}

func (gs *GlobalSecretResource) MarshalLog() interface{} {
	return secretResource{
		Meta: gs.Meta,
		Spec: "***",
	}
}
