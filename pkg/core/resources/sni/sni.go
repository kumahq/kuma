package sni

import (
	"fmt"
	"strings"

	k8s_validation "k8s.io/apimachinery/pkg/util/validation"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
)

const (
	dnsHostnameLimit = 253
	sniFormatPrefix  = "sni"
)

// sniCapableShortNames lists resource type ShortNames whose KRI is rendered
// as an SNI by xDS generators.
var sniCapableShortNames = map[string]struct{}{
	"msvc":   {}, // MeshService
	"extsvc": {}, // MeshExternalService
	"mzsvc":  {}, // MeshMultiZoneService
}

// FromKRI builds an SNI in the KRI-derived format described in MADR-101.
//
// The format is:
//
//	sni.<short>.<mesh>.<name>.<sectionName>                          (5 segments) — global-originated (k8s or universal)
//	sni.<short>.<mesh>.<zone>.<name>.<sectionName>                   (6 segments) — zone-originated on universal
//	sni.<short>.<mesh>.<zone>.<namespace>.<name>.<sectionName>       (7 segments) — zone-originated on k8s
//
// Namespace is only emitted when the resource is also zone-scoped: a
// global-originated resource (zone == "") drops the namespace segment even
// if the k8s label is present. A global-originated resource can only
// originate from Global CP, so namespace collisions across zones aren't
// possible — universal has no namespace, and k8s would always be
// kuma-system on Global.
func FromKRI(id kri.Identifier) string {
	return strings.Join(buildSegments(id), ".")
}

// FromKRIE is the non-panicking variant of FromKRI. It returns an error when
// the resource type is not SNI-capable or the descriptor cannot be resolved,
// so callers running in marshaling code paths do not have to defend against
// programmer errors with recover.
func FromKRIE(id kri.Identifier) (string, error) {
	desc, err := registry.Global().DescriptorFor(id.ResourceType)
	if err != nil {
		return "", err
	}
	if _, ok := sniCapableShortNames[desc.ShortName]; !ok {
		return "", fmt.Errorf("resource type not supported for SNI: %s", id.ResourceType)
	}
	return strings.Join(buildSegmentsFor(desc, id), "."), nil
}

// Section describes a single (port, sectionName) pair that contributes one
// SNI for a destination. SectionName is what gets placed into the trailing
// segment of the SNI and matches kri.Identifier.SectionName.
type Section struct {
	Port        int32
	SectionName string
}

// SectionLister is implemented by resource spec types that contribute SNI
// sections.
type SectionLister interface {
	SNIs() []Section
}

// ValidateKRI returns every reason the SNI that FromKRI would produce does
// not satisfy MADR-101 / DNS-1123 naming rules:
//
//   - Mesh, Name and SectionName are non-empty
//   - Mesh, Name, Zone and Namespace conform to RFC 1123 ([a-z0-9]([-a-z0-9]*[a-z0-9])?, max 63 chars)
//   - SectionName conforms to RFC 1123
//   - total length ≤ 253 (DNS hostname limit)
func ValidateKRI(id kri.Identifier) []error {
	desc, err := registry.Global().DescriptorFor(id.ResourceType)
	if err != nil {
		return []error{fmt.Errorf("unknown resource type %q: %w", id.ResourceType, err)}
	}
	if _, ok := sniCapableShortNames[desc.ShortName]; !ok {
		return nil
	}
	if id.Mesh == "" || id.Name == "" || id.SectionName == "" {
		return []error{fmt.Errorf("mesh, name and sectionName must be non-empty")}
	}

	var errs []error

	type field struct {
		name, value string
	}
	fields := []field{
		{"mesh", id.Mesh},
		{"name", id.Name},
	}
	if id.Zone != "" {
		fields = append(fields, field{"zone", id.Zone})
		if id.Namespace != "" {
			fields = append(fields, field{"namespace", id.Namespace})
		}
	}
	for _, f := range fields {
		if msgs := k8s_validation.IsDNS1123Label(f.value); len(msgs) > 0 {
			errs = append(errs, fmt.Errorf("%s %q does not conform to RFC 1123: %s", f.name, f.value, strings.Join(msgs, "; ")))
		}
	}
	if msgs := k8s_validation.IsDNS1123Label(id.SectionName); len(msgs) > 0 {
		errs = append(errs, fmt.Errorf("%s", strings.Join(msgs, "; ")))
	}

	if sni := FromKRI(id); len(sni) > dnsHostnameLimit {
		errs = append(errs, fmt.Errorf("computed SNI is %d characters which exceeds the DNS hostname limit (%d)", len(sni), dnsHostnameLimit))
	}
	return errs
}

func buildSegments(id kri.Identifier) []string {
	desc, err := registry.Global().DescriptorFor(id.ResourceType)
	if err != nil {
		panic("unknown resource type " + string(id.ResourceType))
	}
	if _, ok := sniCapableShortNames[desc.ShortName]; !ok {
		panic("resource type not supported for SNI: " + string(id.ResourceType))
	}
	return buildSegmentsFor(desc, id)
}

func buildSegmentsFor(desc core_model.ResourceTypeDescriptor, id kri.Identifier) []string {
	segments := []string{sniFormatPrefix, desc.ShortName, id.Mesh}
	if id.Zone != "" {
		segments = append(segments, id.Zone)
		if id.Namespace != "" {
			segments = append(segments, id.Namespace)
		}
	}
	segments = append(segments, id.Name, id.SectionName)
	return segments
}
