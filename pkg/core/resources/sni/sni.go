package sni

import (
	"fmt"
	"strings"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
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
//	sni.<short>.<mesh>.<name>.<sectionName>                          (5 segments) — global-originated
//	sni.<short>.<mesh>.<zone>.<name>.<sectionName>                   (6 segments) — zone-originated resource on universal
//	sni.<short>.<mesh>.<zone>.<namespace>.<name>.<sectionName>       (7 segments) — zone-originated resource on k8s
func FromKRI(id kri.Identifier) string {
	return strings.Join(buildSegments(id), ".")
}

// ValidateKRI returns every reason the SNI that FromKRI would produce does
// not satisfy MADR-101 / DNS-1035 naming rules:
//
//   - Mesh, Name and SectionName are non-empty
//   - if Namespace is non-empty, Zone must also be non-empty
//   - Mesh, Name, Zone and Namespace conform to RFC 1035 ([a-z]([-a-z0-9]*[a-z0-9])?, max 63 chars)
//   - SectionName conforms to RFC 1035 OR is an all-digit port number (1-5 digits, ≤63 chars)
//   - total length ≤ 253 (DNS hostname limit)
func ValidateKRI(id kri.Identifier) []error {
	desc, err := registry.Global().DescriptorFor(id.ResourceType)
	if err != nil {
		return nil
	}
	if _, ok := sniCapableShortNames[desc.ShortName]; !ok {
		return nil
	}
	if id.Mesh == "" || id.Name == "" || id.SectionName == "" {
		return []error{fmt.Errorf("mesh, name and sectionName must be non-empty")}
	}

	var errs []error
	if id.Namespace != "" && id.Zone == "" {
		errs = append(errs, fmt.Errorf("namespace %q is set without a zone", id.Namespace))
	}

	type field struct {
		name, value string
	}
	fields := []field{
		{"mesh", id.Mesh},
		{"name", id.Name},
	}
	if id.Zone != "" {
		fields = append(fields, field{"zone", id.Zone})
	}
	if id.Namespace != "" {
		fields = append(fields, field{"namespace", id.Namespace})
	}
	for _, f := range fields {
		if msgs := apimachineryvalidation.NameIsDNS1035Label(f.value, false); len(msgs) > 0 {
			errs = append(errs, fmt.Errorf("%s %q does not conform to RFC 1035: %s", f.name, f.value, strings.Join(msgs, "; ")))
		}
	}
	if msgs := validateSectionName(id.SectionName); len(msgs) > 0 {
		errs = append(errs, fmt.Errorf("port %q does not conform to RFC 1035: %s", id.SectionName, strings.Join(msgs, "; ")))
	}

	if sni := FromKRI(id); len(sni) > dnsHostnameLimit {
		errs = append(errs, fmt.Errorf("computed SNI for port %q is %d characters which exceeds the DNS hostname limit (%d)", id.SectionName, len(sni), dnsHostnameLimit))
	}
	return errs
}

// validateSectionName accepts either a valid DNS-1035 label or an all-digit
// numeric port (1-5 digits). The numeric carve-out preserves backwards
// compatibility for ports without an explicit name (the default formatting
// is the decimal port number, which would otherwise fail RFC 1035 because
// it starts with a digit).
func validateSectionName(s string) []string {
	if s != "" && len(s) <= 5 && isAllDigits(s) {
		return nil
	}
	return apimachineryvalidation.NameIsDNS1035Label(s, false)
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func buildSegments(id kri.Identifier) []string {
	desc, err := registry.Global().DescriptorFor(id.ResourceType)
	if err != nil {
		panic("unknown resource type " + string(id.ResourceType))
	}
	if _, ok := sniCapableShortNames[desc.ShortName]; !ok {
		panic("resource type not supported for SNI: " + string(id.ResourceType))
	}

	segments := []string{sniFormatPrefix, desc.ShortName, id.Mesh}
	if id.Zone != "" {
		segments = append(segments, id.Zone)
	}
	if id.Namespace != "" {
		segments = append(segments, id.Namespace)
	}
	segments = append(segments, id.Name, id.SectionName)
	return segments
}
