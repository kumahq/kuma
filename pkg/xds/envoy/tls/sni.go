package tls

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	meshexternalservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v2/pkg/util/maps"
	envoy_tags "github.com/kumahq/kuma/v2/pkg/xds/envoy/tags"
)

func SNIFromTags(tags envoy_tags.Tags) string {
	extraTags := tags.WithoutTags(mesh_proto.ServiceTag).String()
	service := tags[mesh_proto.ServiceTag]
	if extraTags == "" {
		return service
	}
	return fmt.Sprintf("%s{%s}", service, extraTags)
}

func TagsFromSNI(sni string) (envoy_tags.Tags, error) {
	parts := strings.Split(sni, "{")
	if len(parts) > 2 {
		return nil, errors.New(fmt.Sprintf("cannot parse tags from sni: %s", sni))
	}
	if len(parts) == 1 {
		return envoy_tags.Tags{mesh_proto.ServiceTag: parts[0]}, nil
	}
	cleanedTags := strings.ReplaceAll(parts[1], "}", "")
	tags, err := envoy_tags.TagsFromString(cleanedTags)
	if err != nil {
		return nil, err
	}
	tags[mesh_proto.ServiceTag] = parts[0]
	return tags, nil
}

const (
	sniFormatVersion = "a"
	dnsLabelLimit    = 63
	sniFormatPrefix  = "sni"
	dnsHostnameLimit = 253
)

func SNIForResource(resName string, meshName string, resType model.ResourceType, port int32, additionalData map[string]string) string {
	var mapStrings []string
	for _, key := range maps.SortedKeys(additionalData) {
		mapStrings = append(mapStrings, fmt.Sprintf("%s=%s", key, additionalData[key]))
	}

	hash := fnv.New64a()
	_, _ = fmt.Fprintf(hash, "%s;%s;%v", resName, meshName, strings.Join(mapStrings, ",")) // fnv64a does not return error
	hashBytes := hash.Sum(nil)

	if len(resName) > dnsLabelLimit-1 {
		resName = resName[:dnsLabelLimit-1] + "x"
	}
	if len(meshName) > dnsLabelLimit-1 {
		meshName = meshName[:dnsLabelLimit-1] + "x"
	}

	resTypeAbbrv := ""
	switch resType {
	case meshservice_api.MeshServiceType:
		resTypeAbbrv = "ms"
	case meshexternalservice_api.MeshExternalServiceType:
		resTypeAbbrv = "mes"
	case meshmzservice_api.MeshMultiZoneServiceType:
		resTypeAbbrv = "mzms"
	default:
		panic("resource type not supported for SNI")
	}

	return fmt.Sprintf("%s%x.%s.%d.%s.%s", sniFormatVersion, hashBytes, resName, port, meshName, resTypeAbbrv)
}

// SNIFromKRI builds an SNI in the KRI-derived format described in MADR 101.
//
// The format is:
//
//	sni.<short>.<mesh>.<name>.<sectionName>                          (5 segments) — global-originated
//	sni.<short>.<mesh>.<zone>.<name>.<sectionName>                   (6 segments) — zone-originated resource on universal
//	sni.<short>.<mesh>.<zone>.<namespace>.<name>.<sectionName>       (7 segments) — zone-originated resource on k8s
func SNIFromKRI(id kri.Identifier) string {
	return strings.Join(buildSNISegments(id), ".")
}

// ValidateSNIForKRI returns nil if the SNI that SNIFromKRI would produce
// satisfies the MADR-101 naming rules:
//
//   - Mesh, Name and SectionName are non-empty
//   - if Namespace is non-empty, Zone must also be non-empty
//   - no segment contains "."
//   - each segment length ≤ 63 (DNS label limit)
//   - total length ≤ 253 (DNS hostname limit)
func ValidateSNIForKRI(id kri.Identifier) error {
	if id.Mesh == "" || id.Name == "" || id.SectionName == "" {
		return errors.Errorf("SNI: mesh, name and sectionName must be non-empty: %+v", id)
	}
	if id.Namespace != "" && id.Zone == "" {
		return errors.Errorf("SNI: namespace %q set without zone: %+v", id.Namespace, id)
	}
	segments := buildSNISegments(id)
	total := len(segments) - 1 // dots between segments
	for _, s := range segments {
		if strings.ContainsRune(s, '.') {
			return errors.Errorf("SNI: segment %q contains '.': %+v", s, id)
		}
		if len(s) > dnsLabelLimit {
			return errors.Errorf("SNI: segment %q exceeds DNS label limit (%d > %d): %+v", s, len(s), dnsLabelLimit, id)
		}
		total += len(s)
	}
	if total > dnsHostnameLimit {
		return errors.Errorf("SNI: total length %d exceeds DNS hostname limit %d: %+v", total, dnsHostnameLimit, id)
	}
	return nil
}

func buildSNISegments(id kri.Identifier) []string {
	desc, err := registry.Global().DescriptorFor(id.ResourceType)
	if err != nil {
		panic(errors.Wrapf(err, "SNIFromKRI: unknown resource type %q", id.ResourceType))
	}
	switch id.ResourceType {
	case meshservice_api.MeshServiceType,
		meshexternalservice_api.MeshExternalServiceType,
		meshmzservice_api.MeshMultiZoneServiceType:
	default:
		panic(fmt.Sprintf("SNIFromKRI: resource type %q is not supported for SNI", id.ResourceType))
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
