package tls

import (
	"fmt"
	"hash/fnv"
	"strings"

	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/maps"
)

const (
	sniFormatVersion = "a"
	dnsLabelLimit    = 63
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
