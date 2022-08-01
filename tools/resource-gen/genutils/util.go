package genutils

import (
	"fmt"
	"strings"
	"unicode"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/kumahq/kuma/api/mesh"
)

// KumaResourceForMessage fetches the Kuma resource option out of a message.
func KumaResourceForMessage(desc protoreflect.MessageDescriptor) *mesh.KumaResourceOptions {
	ext := proto.GetExtension(desc.Options(), mesh.E_Resource)
	var resOption *mesh.KumaResourceOptions
	if r, ok := ext.(*mesh.KumaResourceOptions); ok {
		resOption = r
	}

	return resOption
}

// SelectorsForMessage finds all the top-level fields in the message are
// repeated selectors. We want to generate convenience accessors for these.
func SelectorsForMessage(m protoreflect.MessageDescriptor) []string {
	var selectors []string
	fields := m.Fields()

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		m := field.Message()
		if m != nil && m.FullName() == "kuma.mesh.v1alpha1.Selector" {
			fieldName := string(field.Name())
			selectors = append(selectors, strings.Title(fieldName))
		}
	}

	return selectors
}

type ResourceInfo struct {
	ResourceName           string
	ResourceType           string
	ProtoType              string
	Selectors              []string
	SkipRegistration       bool
	SkipValidation         bool
	SkipKubernetesWrappers bool
	ScopeNamespace         bool
	Global                 bool
	KumactlSingular        string
	KumactlPlural          string
	WsReadOnly             bool
	WsAdminOnly            bool
	WsPath                 string
	KdsDirection           string
	AllowToInspect         bool
	StorageVersion         bool
	IsPolicy               bool
	DisplayName            string
}

func ToResourceInfo(desc protoreflect.MessageDescriptor) ResourceInfo {
	r := KumaResourceForMessage(desc)

	out := ResourceInfo{
		ResourceType:           r.Type,
		ResourceName:           r.Name,
		ProtoType:              string(desc.Name()),
		Selectors:              SelectorsForMessage(desc),
		SkipRegistration:       r.SkipRegistration,
		SkipKubernetesWrappers: r.SkipKubernetesWrappers,
		SkipValidation:         r.SkipValidation,
		Global:                 r.Global,
		ScopeNamespace:         r.ScopeNamespace,
		AllowToInspect:         r.AllowToInspect,
		StorageVersion:         r.StorageVersion,
		DisplayName:            r.DisplayName,
	}
	if r.Ws != nil {
		pluralResourceName := r.Ws.Plural
		if pluralResourceName == "" {
			pluralResourceName = r.Ws.Name + "s"
		}
		out.WsReadOnly = r.Ws.ReadOnly
		out.WsAdminOnly = r.Ws.AdminOnly
		out.WsPath = pluralResourceName
		if !r.Ws.ReadOnly {
			out.KumactlSingular = r.Ws.Name
			out.KumactlPlural = pluralResourceName
			// Keep the typo to preserve backward compatibility
			if out.KumactlSingular == "health-check" {
				out.KumactlSingular = "healthcheck"
				out.KumactlPlural = "healthchecks"
			}
		}
	}
	// Working around the fact we don't really differentiate policies from the rest of resources:
	// Anything global can't be a policy as it need to be on a mesh. Anything with locked Ws config is something internal and therefore not a policy
	out.IsPolicy = !out.SkipRegistration && !out.Global && !out.WsAdminOnly && !out.WsReadOnly && out.ResourceType != "Dataplane"
	if out.DisplayName == "" && out.IsPolicy {
		for i, c := range out.ResourceType {
			if unicode.IsUpper(c) && i != 0 {
				out.DisplayName += " "
			}
			out.DisplayName += string(c)
		}
		out.DisplayName += "s"
	}
	switch {
	case r.Kds == nil || (!r.Kds.SendToZone && !r.Kds.SendToGlobal):
		out.KdsDirection = ""
	case r.Kds.SendToGlobal && r.Kds.SendToZone:
		out.KdsDirection = "model.FromZoneToGlobal | model.FromGlobalToZone"
	case r.Kds.SendToGlobal:
		out.KdsDirection = "model.FromZoneToGlobal"
	case r.Kds.SendToZone:
		out.KdsDirection = "model.FromGlobalToZone"
	}

	if p := desc.Parent(); p != nil {
		if _, ok := p.(protoreflect.MessageDescriptor); ok {
			out.ProtoType = fmt.Sprintf("%s_%s", p.Name(), desc.Name())
		}
	}
	return out
}
