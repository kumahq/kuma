package api_server_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
)

// pathParam normalizes path parameter names so that a route like
// "/meshes/{mesh}/dataplanes/{name}/_config" matches the spec entry
// regardless of how the parameter is named on either side.
var pathParam = regexp.MustCompile(`\{[^}]+\}`)

// normalizePath blanks out path parameter names and, except for the last
// segment, resource collection names (WsPath). The spec sometimes documents a
// family of routes with a generic parameter (e.g.
// "/meshes/{mesh}/{policyType}/{policyName}/_resources/dataplanes") while the
// server registers one concrete route per resource type. The last segment is
// kept literal so that cross-mesh list routes ("/meshmetrics") stay
// distinguishable from each other.
func normalizePath(path string, wsPaths []string) string {
	p := pathParam.ReplaceAllString(path, "{}")
	segments := strings.Split(p, "/")
	for i, segment := range segments {
		if i == len(segments)-1 {
			continue
		}
		for _, wsPath := range wsPaths {
			if segment == wsPath {
				segments[i] = "{}"
			}
		}
	}
	return strings.Join(segments, "/")
}

// undocumentedRoutes are routes served by the API server that have no OpenAPI
// spec entry even after normalization. Each one is known drift between the
// specs and the registered routes; fixing an entry means documenting it in
// the spec and removing it from this list.
var undocumentedRoutes = map[string]string{
	"GET /dataplane-insights":       "no rest.yaml fragment is generated for this resource",
	"GET /global-secrets":           "no rest.yaml fragment is generated for this resource",
	"GET /globalsecrets":            "no rest.yaml fragment is generated for this resource",
	"GET /zone-insights":            "no rest.yaml fragment is generated for this resource",
	"GET /zones":                    "no rest.yaml fragment is generated for this resource",
	"GET /{}/{}/dataplane-insights": "in-mesh list missing from rest.yaml fragment",
	"GET /{}/{}/{}/{}/{}":           "generic dataplane envoy admin route, spec documents the concrete xds/clusters/stats subpaths",
	"GET /global-insights":          "spec path is /global-insight, server serves /global-insights",
	"GET /policies":                 "not documented in the OpenAPI spec",
	"GET /who-am-i":                 "not documented in the OpenAPI spec",
	"POST /tokens/dataplane":        "not documented in the OpenAPI spec",
	"POST /tokens/zone":             "not documented in the OpenAPI spec",
}

var _ = Describe("OpenAPI conformance", func() {
	wsPaths := func() []string {
		var paths []string
		for _, desc := range registry.Global().ObjectDescriptors(model.HasWsEnabled()) {
			paths = append(paths, desc.WsPath)
			if desc.AlternativeWsPath != "" {
				paths = append(paths, desc.AlternativeWsPath)
			}
		}
		return paths
	}()

	specPaths := func() map[string]bool {
		files := []string{filepath.Join("..", "..", "api", "openapi", "specs", "api.yaml")}
		for _, pattern := range []string{
			filepath.Join("..", "..", "api", "mesh", "v1alpha1", "*", "rest.yaml"),
			filepath.Join("..", "..", "api", "system", "v1alpha1", "*", "rest.yaml"),
			filepath.Join("..", "..", "api", "openapi", "specs", "kri", "kri.yaml"),
			filepath.Join("..", "..", "pkg", "plugins", "policies", "*", "api", "v1alpha1", "rest.yaml"),
			filepath.Join("..", "..", "pkg", "core", "resources", "apis", "*", "api", "v1alpha1", "rest.yaml"),
		} {
			matches, err := filepath.Glob(pattern)
			Expect(err).ToNot(HaveOccurred())
			files = append(files, matches...)
		}

		paths := map[string]bool{}
		for _, file := range files {
			content, err := os.ReadFile(file)
			Expect(err).ToNot(HaveOccurred())
			var spec struct {
				Paths map[string]map[string]json.RawMessage `json:"paths"`
			}
			Expect(yaml.Unmarshal(content, &spec)).To(Succeed(), file)
			for path, methods := range spec.Paths {
				for method := range methods {
					paths[strings.ToUpper(method)+" "+normalizePath(path, wsPaths)] = true
				}
			}
		}
		return paths
	}

	It("registered routes should be documented in the OpenAPI spec", func() {
		apiServer, _, stop := StartApiServer(NewTestApiServerConfigurer().WithStore(memory.NewStore()))
		defer stop()

		spec := specPaths()
		usedAllowlist := map[string]bool{}
		var missingInSpec []string
		for _, route := range apiServer.Routes() {
			method, path, _ := strings.Cut(route, " ")
			normalized := method + " " + normalizePath(path, wsPaths)
			if spec[normalized] {
				continue
			}
			if _, allowed := undocumentedRoutes[normalized]; allowed {
				usedAllowlist[normalized] = true
				continue
			}
			missingInSpec = append(missingInSpec, normalized)
		}
		sort.Strings(missingInSpec)
		Expect(missingInSpec).To(BeEmpty(), "routes served by the API server but missing from the OpenAPI spec")

		var staleAllowlist []string
		for route := range undocumentedRoutes {
			if !usedAllowlist[route] {
				staleAllowlist = append(staleAllowlist, route)
			}
		}
		sort.Strings(staleAllowlist)
		Expect(staleAllowlist).To(BeEmpty(), "allowlist entries no longer served or now documented; remove them from undocumentedRoutes")
	})
})
