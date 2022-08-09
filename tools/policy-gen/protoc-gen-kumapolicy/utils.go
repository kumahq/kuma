package main

import (
	"strings"
	"unicode"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/kumahq/kuma/api/mesh"
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
		switch {
		case strings.HasSuffix(res.Name, "y"):
			res.Plural = strings.TrimSuffix(res.Name, "y") + "ies"
		case strings.HasSuffix(res.Name, "s"):
			res.Plural = res.Name + "es"
		default:
			res.Plural = res.Name + "s"
		}
	}
	for i, c := range res.Name {
		if unicode.IsUpper(c) && i != 0 {
			res.SingularDisplayName += " "
		}
		res.SingularDisplayName += string(c)
	}
	for i, c := range res.Plural {
		if unicode.IsUpper(c) && i != 0 {
			res.PluralDisplayName += " "
		}
		res.PluralDisplayName += string(c)
	}
	res.Path = strings.ToLower(res.Plural)
	res.AlternativeNames = []string{strings.ToLower(res.Name)}
	return res
}
