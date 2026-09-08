package extensions

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_extensions "github.com/kumahq/kuma/v3/pkg/core/resources/extensions"
)

type fakeConfig struct {
	Field string `json:"field,omitempty"`
}

var fakePoint = core_extensions.Point{
	ResourceType:  "FakeResource",
	SchemaPath:    []string{"spec", "extension"},
	Discriminator: "type",
}

func fakeWrappers(values ...string) []wrapper {
	registered := make([]core_extensions.Extension, 0, len(values))
	for _, v := range values {
		registered = append(registered, core_extensions.Extension{Point: fakePoint, Value: v, Config: &fakeConfig{}})
	}
	wrappers, err := newWrappers(registered)
	Expect(err).ToNot(HaveOccurred())
	return wrappers
}

// specWithExtensionPoint is the shape the generator expects to find in the
// document: an item schema with an extension object carrying a discriminator.
func specWithExtensionPoint() map[string]any {
	return map[string]any{
		"components": map[string]any{
			"schemas": map[string]any{
				"FakeResourceItem": map[string]any{
					"properties": map[string]any{
						"spec": map[string]any{
							"properties": map[string]any{
								"extension": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"type":   map[string]any{"type": "string"},
										"config": map[string]any{"x-kubernetes-preserve-unknown-fields": true},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

var _ = Describe("newWrappers", func() {
	It("should name the schema after the resource and the discriminator value", func() {
		wrappers := fakeWrappers("Route53")

		Expect(wrappers[0].kind).To(Equal("FakeResourceRoute53"))
		Expect(wrappers[0].schemaName).To(Equal("FakeResourceRoute53ExtensionConfig"))
	})

	It("should upper-case a lower-case discriminator value, since the kind becomes a Go type", func() {
		Expect(fakeWrappers("vault")[0].kind).To(Equal("FakeResourceVault"))
	})

	It("should drop characters that are illegal in a kind", func() {
		Expect(fakeWrappers("cert-manager.io")[0].kind).To(Equal("FakeResourceCertmanagerio"))
	})

	It("should reject two values that collapse to the same name", func() {
		registered := []core_extensions.Extension{
			{Point: fakePoint, Value: "vault", Config: &fakeConfig{}},
			{Point: fakePoint, Value: "Vault", Config: &fakeConfig{}},
		}

		_, err := newWrappers(registered)

		Expect(err).To(MatchError(ContainSubstring(`both generate the schema name "FakeResourceVault"`)))
	})

	It("should reject a registration with no character usable in a name", func() {
		unusable := core_extensions.Point{ResourceType: "123", SchemaPath: []string{"spec"}, Discriminator: "type"}
		registered := []core_extensions.Extension{{Point: unusable, Value: "---", Config: &fakeConfig{}}}

		_, err := newWrappers(registered)

		Expect(err).To(MatchError(ContainSubstring("no character usable in a schema name")))
	})
})

var _ = Describe("checkExtensionPoint", func() {
	It("should accept a point that resolves to an object with the discriminator", func() {
		Expect(checkExtensionPoint(specWithExtensionPoint(), fakePoint)).To(Succeed())
	})

	It("should reject a point whose resource has no item schema", func() {
		moved := fakePoint
		moved.ResourceType = "Missing"

		err := checkExtensionPoint(specWithExtensionPoint(), moved)

		Expect(err).To(MatchError(ContainSubstring("Missing does not have a MissingItem.spec.extension schema")))
	})

	// yq creates whatever path it is given, so a point that has drifted has to be
	// an error rather than a new branch grown in the document.
	It("should reject a schema path that is not in the document", func() {
		moved := fakePoint
		moved.SchemaPath = []string{"spec", "provider", "extension"}

		err := checkExtensionPoint(specWithExtensionPoint(), moved)

		Expect(err).To(MatchError(ContainSubstring("does not have a FakeResourceItem.spec.provider.extension schema")))
	})

	It("should reject a discriminator the extension object does not have", func() {
		moved := fakePoint
		moved.Discriminator = "name"

		err := checkExtensionPoint(specWithExtensionPoint(), moved)

		Expect(err).To(MatchError(ContainSubstring(`has no "name" property to discriminate on`)))
	})
})

var _ = Describe("propertyPath", func() {
	It("should interleave properties so a schema path addresses a real node", func() {
		Expect(propertyPath([]string{"spec", "provider", "extension"})).To(Equal(
			[]string{"properties", "spec", "properties", "provider", "properties", "extension"}))
	})
})

var _ = Describe("branches", func() {
	It("should pair each value with the schema it selects", func() {
		result := branches(fakePoint, fakeWrappers("Route53"))

		Expect(result[0]).To(Equal(map[string]any{
			"title": "Route53",
			"properties": map[string]any{
				"type":   map[string]any{"const": "Route53"},
				"config": map[string]any{"$ref": "#/components/schemas/FakeResourceRoute53ExtensionConfig"},
			},
		}))
	})

	// Without the catch-all the spec would reject extension types the control
	// plane accepts, including any a third party adds.
	It("should end with a catch-all excluding every known value", func() {
		result := branches(fakePoint, fakeWrappers("acmpca", "vault"))

		Expect(result).To(HaveLen(3))
		catchAll := result[2].(map[string]any)
		Expect(catchAll["title"]).To(Equal("Other"))
		Expect(catchAll["properties"]).To(Equal(map[string]any{
			"type": map[string]any{"not": map[string]any{"enum": []any{"acmpca", "vault"}}},
		}))
	})

	It("should ignore extensions belonging to another resource", func() {
		other := fakePoint
		other.ResourceType = "OtherResource"
		mixed := append(fakeWrappers("vault"), wrapper{
			ext:        core_extensions.Extension{Point: other, Value: "elsewhere", Config: &fakeConfig{}},
			kind:       "OtherResourceElsewhere",
			schemaName: "OtherResourceElsewhereExtensionConfig",
		})

		result := branches(fakePoint, mixed)

		Expect(result).To(HaveLen(2))
		Expect(result[0].(map[string]any)["title"]).To(Equal("vault"))
	})
})

var _ = Describe("patchExpression", func() {
	It("should add the config schema and point the extension node at it", func() {
		wrappers := fakeWrappers("Route53")
		schemas := map[string]any{"FakeResourceRoute53": map[string]any{"type": "object"}}

		expr, err := patchExpression(wrappers, schemas)

		Expect(err).ToNot(HaveOccurred())
		Expect(expr).To(ContainSubstring(`."components"."schemas"."FakeResourceRoute53ExtensionConfig" = {"type":"object"}`))
		Expect(expr).To(ContainSubstring(
			`."components"."schemas"."FakeResourceItem"."properties"."spec"."properties"."extension".oneOf = [`))
	})

	// The config is a resource spec like any other, so its own discriminated
	// unions have to be described too.
	It("should describe a union inside the config", func() {
		wrappers := fakeWrappers("vault")
		schemas := map[string]any{"FakeResourceVault": map[string]any{
			"properties": map[string]any{
				"type":   map[string]any{"enum": []any{"Server", "Agent"}},
				"server": map[string]any{"type": "object"},
				"agent":  map[string]any{"type": "object"},
			},
		}}

		expr, err := patchExpression(wrappers, schemas)

		Expect(err).ToNot(HaveOccurred())
		Expect(expr).To(ContainSubstring(`."components"."schemas"."FakeResourceVaultExtensionConfig".oneOf = `))
	})
})

var _ = Describe("writePackage", func() {
	It("should wrap each config in a CRD root type controller-gen will walk", func() {
		dir := GinkgoT().TempDir()

		Expect(writePackage(dir, fakeWrappers("Route53"))).To(Succeed())

		doc, err := os.ReadFile(filepath.Join(dir, "doc.go"))
		Expect(err).ToNot(HaveOccurred())
		Expect(string(doc)).To(ContainSubstring("+groupName=openapi.kuma.io"))

		types, err := os.ReadFile(filepath.Join(dir, "types.go"))
		Expect(err).ToNot(HaveOccurred())
		Expect(string(types)).To(ContainSubstring(`ext0 "github.com/kumahq/kuma/v3/tools/openapi/extensions"`))
		Expect(string(types)).To(ContainSubstring("+kubebuilder:object:root=true"))
		Expect(string(types)).To(ContainSubstring("type FakeResourceRoute53 struct {"))
		Expect(string(types)).To(ContainSubstring("Spec *ext0.fakeConfig `json:\"spec,omitempty\"`"))
	})
})

var _ = Describe("Generate", func() {
	It("should leave the document alone when nothing is registered", func() {
		// Runs before the end-to-end spec below registers anything, which is the
		// property that keeps a downstream product's extensions out of Kuma's spec.
		Expect(core_extensions.Registered()).To(BeEmpty())

		spec := filepath.Join(GinkgoT().TempDir(), "openapi.yaml")
		Expect(os.WriteFile(spec, []byte("openapi: 3.1.0\n"), 0o600)).To(Succeed())

		Expect(Generate(context.Background(), Options{Spec: spec})).To(Succeed())

		Expect(os.ReadFile(spec)).To(Equal([]byte("openapi: 3.1.0\n")))
	})

	Context("with a registered extension", func() {
		var (
			dir      string
			specPath string
			exprPath string
			opts     Options
		)

		BeforeEach(func() {
			dir = GinkgoT().TempDir()
			specPath = filepath.Join(dir, "openapi.yaml")
			exprPath = filepath.Join(dir, "expression.yq")

			raw, err := yaml.Marshal(specWithExtensionPoint())
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(specPath, raw, 0o600)).To(Succeed())

			opts = Options{
				Spec:             specPath,
				WorkDir:          filepath.Join(dir, "work"),
				ControllerGenBin: stubControllerGen(dir),
				YqBin:            stubYq(dir, exprPath),
				Stderr:           GinkgoWriter,
			}
		})

		It("should patch the document with the schema controller-gen produced", func() {
			core_extensions.Register(core_extensions.Extension{Point: fakePoint, Value: "fake", Config: &fakeConfig{}})

			Expect(Generate(context.Background(), opts)).To(Succeed())

			expr, err := os.ReadFile(exprPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(expr)).To(ContainSubstring(
				`."components"."schemas"."FakeResourceFakeExtensionConfig" = {"description":"Fake config"`))
			Expect(string(expr)).To(ContainSubstring(
				`."components"."schemas"."FakeResourceItem"."properties"."spec"."properties"."extension".oneOf = [` +
					`{"properties":{"config":{"$ref":"#/components/schemas/FakeResourceFakeExtensionConfig"},` +
					`"type":{"const":"fake"}},"title":"fake"},`))
		})

		It("should remove the scratch directory it created and nothing else", func() {
			keep := filepath.Join(opts.WorkDir, "keep-me")
			Expect(os.MkdirAll(keep, 0o755)).To(Succeed())

			Expect(Generate(context.Background(), opts)).To(Succeed())

			Expect(keep).To(BeADirectory())
			entries, err := os.ReadDir(opts.WorkDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(1), "the generated package should have been cleaned up")
			Expect(entries[0].Name()).To(Equal("keep-me"))
		})

		It("should fail when the document has no such extension point", func() {
			Expect(os.WriteFile(specPath, []byte("openapi: 3.1.0\n"), 0o600)).To(Succeed())

			err := Generate(context.Background(), opts)

			Expect(err).To(MatchError(ContainSubstring("FakeResource does not have a FakeResourceItem.spec.extension schema")))
		})

		It("should reject an empty work dir rather than guess one", func() {
			opts.WorkDir = ""

			Expect(Generate(context.Background(), opts)).To(MatchError(ContainSubstring("work dir must not be empty")))
		})
	})
})

// stubControllerGen stands in for controller-gen, writing the CRD the generator
// expects to read back. It keeps the end-to-end spec independent of a binary the
// test environment may not have.
func stubControllerGen(dir string) string {
	return writeScript(dir, "controller-gen", `#!/usr/bin/env bash
set -eu
out=""
for arg in "$@"; do
  case "$arg" in
    output:crd:artifacts:config=*) out="${arg#output:crd:artifacts:config=}" ;;
  esac
done
mkdir -p "$out"
cat > "$out/fake.yaml" <<'YAML'
spec:
  names:
    kind: FakeResourceFake
  versions:
    - schema:
        openAPIV3Schema:
          properties:
            spec:
              description: Fake config
              type: object
              properties:
                field:
                  type: string
YAML
`)
}

// stubYq records the expression it was handed instead of applying it, so the
// assertions are about what the generator asks for rather than about yq.
func stubYq(dir, exprPath string) string {
	return writeScript(dir, "yq", `#!/usr/bin/env bash
set -eu
prev=""
for arg in "$@"; do
  if [ "$prev" = "--from-file" ]; then cp "$arg" `+exprPath+`; fi
  prev="$arg"
done
`)
}

func writeScript(dir, name, body string) string {
	path := filepath.Join(dir, name)
	Expect(os.WriteFile(path, []byte(body), 0o700)).To(Succeed())
	return path
}
