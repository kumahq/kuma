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
// kept literal so that routes of different resource types stay distinguishable.
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
	"GET /{}/{}/{}/{}/{}": "generic dataplane envoy admin route, spec documents the concrete xds/clusters/stats subpaths",
}

// specFiles returns every OpenAPI document of Kuma's own REST API: the base
// specs plus the fragments the generators write next to their resources.
func specFiles() []string {
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
	return files
}

// resourceMetaTimestamps are the meta fields every resource response carries:
// the REST layer always serializes them (see rest/v1alpha1.ResourceMeta), so
// the schema of a resource response has to document them.
var resourceMetaTimestamps = []string{"creationTime", "modificationTime"}

// isResourceResponseSchema reports whether a schema name describes a resource
// response. The generators name those <Type>Item for the CRUD endpoints, and
// the overview specs name them <Type>WithMeta.
func isResourceResponseSchema(name string) bool {
	return strings.HasSuffix(name, "Item") || strings.HasSuffix(name, "WithMeta")
}

// schemaProperties collects the property names a schema declares, following
// local $ref and allOf so a composed schema is treated as a single object.
// Cross-file references cannot be resolved here, and a resource response takes
// its meta from the document it is defined in anyway.
func schemaProperties(schemas map[string]any, schema map[string]any, seen map[string]bool, out map[string]bool) {
	if ref, ok := schema["$ref"].(string); ok {
		name := ref[strings.LastIndex(ref, "/")+1:]
		if !seen[name] {
			seen[name] = true
			if target, ok := schemas[name].(map[string]any); ok {
				schemaProperties(schemas, target, seen, out)
			}
		}
		return
	}
	if properties, ok := schema["properties"].(map[string]any); ok {
		for name := range properties {
			out[name] = true
		}
	}
	if allOf, ok := schema["allOf"].([]any); ok {
		for _, part := range allOf {
			if partSchema, ok := part.(map[string]any); ok {
				schemaProperties(schemas, partSchema, seen, out)
			}
		}
	}
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
		files := specFiles()

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
		readOnlyPosts := map[string]bool{}
		for _, desc := range registry.Global().ObjectDescriptors(model.HasWsEnabled()) {
			if !desc.ReadOnly {
				continue
			}
			for _, path := range []string{desc.WsPath, desc.AlternativeWsPath} {
				if path == "" {
					continue
				}
				path = "/" + path
				if desc.Scope == model.ScopeMesh {
					path = "/meshes/{mesh}" + path
				}
				readOnlyPosts["POST "+normalizePath(path, wsPaths)] = true
			}
		}

		usedAllowlist := map[string]bool{}
		var missingInSpec []string
		for _, route := range apiServer.Routes() {
			method, path, _ := strings.Cut(route, " ")
			normalized := method + " " + normalizePath(path, wsPaths)
			if spec[normalized] || readOnlyPosts[normalized] {
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

	It("resource responses should document the resource meta timestamps", func() {
		var missing []string
		for _, file := range specFiles() {
			content, err := os.ReadFile(file)
			Expect(err).ToNot(HaveOccurred())

			var spec struct {
				Components struct {
					Schemas map[string]any `json:"schemas"`
				} `json:"components"`
			}
			Expect(yaml.Unmarshal(content, &spec)).To(Succeed(), file)

			for name, raw := range spec.Components.Schemas {
				if !isResourceResponseSchema(name) {
					continue
				}
				schema, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				properties := map[string]bool{}
				schemaProperties(spec.Components.Schemas, schema, map[string]bool{name: true}, properties)
				for _, field := range resourceMetaTimestamps {
					if !properties[field] {
						missing = append(missing, filepath.Base(filepath.Dir(file))+"/"+name+"."+field)
					}
				}
			}
		}
		sort.Strings(missing)
		Expect(missing).To(BeEmpty(), "resource response schemas that do not document every resource meta timestamp")
	})
})
