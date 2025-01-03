package genutils

import (
	"fmt"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/kumahq/kuma/api/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

// ProtoMessageFunc ...
type ProtoMessageFunc func(protoreflect.MessageType) bool

// OnKumaResourceMessage ...
func OnKumaResourceMessage(pkg string, f ProtoMessageFunc) ProtoMessageFunc {
	return func(m protoreflect.MessageType) bool {
		r := KumaResourceForMessage(m.Descriptor())
		if r == nil {
			return true
		}

		if r.Package == pkg {
			return f(m)
		}

		return true
	}
}

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
			caser := cases.Title(language.English)
			selectors = append(selectors, caser.String(fieldName))
		}
	}

	return selectors
}

type ResourceInfo struct {
	ResourceName             string
	ResourceType             string
	ProtoType                string
	Selectors                []string
	SkipRegistration         bool
	SkipKubernetesWrappers   bool
	ScopeNamespace           bool
	Global                   bool
	KumactlSingular          string
	KumactlPlural            string
	ShortName                string
	WsReadOnly               bool
	WsAdminOnly              bool
	WsPath                   string
	KdsDirection             string
	AllowToInspect           bool
	StorageVersion           bool
	IsPolicy                 bool
	SingularDisplayName      string
	PluralDisplayName        string
	IsExperimental           bool
	AdditionalPrinterColumns []string
	HasInsights              bool
	IsProxy                  bool
}

func ToResourceInfo(desc protoreflect.MessageDescriptor) ResourceInfo {
	r := KumaResourceForMessage(desc)

	out := ResourceInfo{
		ResourceType:             r.Type,
		ResourceName:             r.Name,
		ProtoType:                string(desc.Name()),
		Selectors:                SelectorsForMessage(desc),
		SkipRegistration:         r.SkipRegistration,
		SkipKubernetesWrappers:   r.SkipKubernetesWrappers,
		Global:                   r.Global,
		ShortName:                r.ShortName,
		ScopeNamespace:           r.ScopeNamespace,
		AllowToInspect:           r.AllowToInspect,
		StorageVersion:           r.StorageVersion,
		SingularDisplayName:      core_model.DisplayName(r.Type),
		PluralDisplayName:        r.PluralDisplayName,
		IsExperimental:           r.IsExperimental,
		AdditionalPrinterColumns: r.AdditionalPrinterColumns,
		HasInsights:              r.HasInsights,
		IsProxy:                  r.IsProxy,
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
	if out.PluralDisplayName == "" {
		out.PluralDisplayName = core_model.PluralType(core_model.DisplayName(r.Type))
	}
	// Working around the fact we don't really differentiate policies from the rest of resources:
	// Anything global can't be a policy as it need to be on a mesh. Anything with locked Ws config is something internal and therefore not a policy
	out.IsPolicy = !out.SkipRegistration && !out.Global && !out.WsAdminOnly && !out.WsReadOnly && out.ResourceType != "Dataplane" && out.ResourceType != "ExternalService" && out.ResourceType != "MeshGateway"
	switch {
	case r.Kds == nil || (!r.Kds.SendToZone && !r.Kds.SendToGlobal):
		out.KdsDirection = ""
	case r.Kds.SendToGlobal && r.Kds.SendToZone:
		out.KdsDirection = "model.ZoneToGlobalFlag | model.GlobalToAllButOriginalZoneFlag"
	case r.Kds.SendToGlobal:
		out.KdsDirection = "model.ZoneToGlobalFlag"
	case r.Kds.SendToZone:
		out.KdsDirection = "model.GlobalToAllZonesFlag"
	}

	if out.ResourceType == "MeshGateway" {
		out.KdsDirection = "model.ZoneToGlobalFlag | model.GlobalToAllZonesFlag"
	}

	if p := desc.Parent(); p != nil {
		if _, ok := p.(protoreflect.MessageDescriptor); ok {
			out.ProtoType = fmt.Sprintf("%s_%s", p.Name(), desc.Name())
		}
	}
	return out
}
