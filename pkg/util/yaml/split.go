package yaml

import (
	"regexp"
	"strings"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var (
	sep     = regexp.MustCompile("(?:^|\\s*\n)---\\s*")
	comment = regexp.MustCompile("^#.*$")
)

// SplitYAML takes YAMLs separated by `---` line and splits it into multiple YAMLs. Empty entries are ignored
func SplitYAML(yamls string) []string {
	var result []string
	// Making sure that any extra whitespace in YAML stream doesn't interfere in splitting documents correctly.
	trimYAMLs := strings.TrimSpace(yamls)
	docs := sep.Split(trimYAMLs, -1)
	for _, doc := range docs {
		doc = comment.ReplaceAllString(doc, "")
		if doc == "" {
			continue
		}
		doc = strings.TrimSpace(doc)
		result = append(result, doc)
	}
	return result
}

func GetResourcesToYaml(
	resourceSet *core_xds.ResourceSet,
	typ envoy_resource.Type,
) ([]byte, error) {
	resources, err := resourceSet.ListOf(typ).ToDeltaDiscoveryResponse()
	if err != nil {
		return nil, err
	}
	return util_proto.ToYAML(resources)
}
