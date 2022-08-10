package main

import (
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/kumahq/kuma/api/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type PolicyConfig struct {
	Name                string
	Plural              string
	SkipRegistration    bool
	SingularDisplayName string
	PluralDisplayName   string
	Path                string
	AlternativeNames    []string
}

func NewPolicyConfig(desc protoreflect.MessageDescriptor) PolicyConfig {
	ext := proto.GetExtension(desc.Options(), mesh.E_Policy)
	var resOption *mesh.KumaPolicyOptions
	if r, ok := ext.(*mesh.KumaPolicyOptions); ok {
		resOption = r
	}
	res := PolicyConfig{
		Name:             string(desc.FullName().Name()),
		SkipRegistration: resOption.SkipRegistration,
		Plural:           resOption.Plural,
	}
	if resOption.Plural == "" {
		res.Plural = core_model.PluralType(res.Name)
	}
	res.SingularDisplayName = core_model.DisplayName(res.Name)
	res.PluralDisplayName = core_model.PluralDisplayName(res.Name)
	res.Path = strings.ToLower(res.Plural)
	res.AlternativeNames = []string{strings.ToLower(res.Name)}
	return res
}
