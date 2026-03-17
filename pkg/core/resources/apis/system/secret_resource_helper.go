package system

import (
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

type secretResource struct {
	Meta model.ResourceMeta
	Spec string
}

func (l *SecretResourceList) MarshalLog() any {
	list := make([]any, 0, len(l.Items))
	for _, res := range l.Items {
		list = append(list, res.MarshalLog())
	}
	return list
}

func (sr *SecretResource) MarshalLog() any {
	return secretResource{
		Meta: sr.Meta,
		Spec: "***",
	}
}

func (l *GlobalSecretResourceList) MarshalLog() any {
	list := make([]any, 0, len(l.Items))
	for _, res := range l.Items {
		list = append(list, res.MarshalLog())
	}
	return list
}

func (gs *GlobalSecretResource) MarshalLog() any {
	return secretResource{
		Meta: gs.Meta,
		Spec: "***",
	}
}
