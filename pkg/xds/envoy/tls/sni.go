package tls

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/util/maps"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
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
)

func SNIForResource(resName string, meshName string, resType model.ResourceType, port uint32, additionalData map[string]string) string {
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
