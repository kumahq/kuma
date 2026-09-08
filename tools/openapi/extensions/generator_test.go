package extensions

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	hostnamegenerator_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	core_extensions "github.com/kumahq/kuma/v3/pkg/core/resources/extensions"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
)

type FakeConfig struct {
	Field string `json:"field,omitempty"`
}

var fakePoint = core_extensions.Point{
	ResourceType:   "FakeResource",
	SchemaPath:     []string{"spec", "extension"},
	Discriminator:  "type",
	ConfigProperty: "config",
}

// registerFake stubs the registry for one spec. Populating the real one would
// leave every spec depending on the order the suite happened to run in.
func registerFake(values ...string) {
	registered := make([]core_extensions.Extension, 0, len(values))
	for _, v := range values {
		registered = append(registered, core_extensions.Extension{Point: fakePoint, Value: v, Config: &FakeConfig{}})
	}
	previous := registeredExtensions
	registeredExtensions = func() []core_extensions.Extension { return registered }
	DeferCleanup(func() { registeredExtensions = previous })
}

// specYAML is the shape the generator expects to find in the document: an item
// schema with an extension object carrying a discriminator.
const specYAML = `
components:
  schemas:
    FakeResourceItem:
      properties:
        spec:
          properties:
            extension:
              type: object
              properties:
                type:
                  type: string
                config:
                  x-kubernetes-preserve-unknown-fields: true
`

func specWithExtensionPoint() map[string]any {
	var spec map[string]any
	Expect(yaml.Unmarshal([]byte(specYAML), &spec)).To(Succeed())
	return spec
}

func fakeWrappers(values ...string) []wrapper {
	registered := make([]core_extensions.Extension, 0, len(values))
	for _, v := range values {
		registered = append(registered, core_extensions.Extension{Point: fakePoint, Value: v, Config: &FakeConfig{}})
	}
	wrappers, err := newWrappers(registered)
	Expect(err).ToNot(HaveOccurred())
	return wrappers
}

func golden(name ...string) string {
	return filepath.Join(append([]string{"testdata"}, name...)...)
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
			{Point: fakePoint, Value: "vault", Config: &FakeConfig{}},
			{Point: fakePoint, Value: "Vault", Config: &FakeConfig{}},
		}

		_, err := newWrappers(registered)

		Expect(err).To(MatchError(ContainSubstring(`both generate the schema name "FakeResourceVault"`)))
	})

	It("should reject a registration with no character usable in a name", func() {
		unusable := core_extensions.Point{ResourceType: "123", SchemaPath: []string{"spec"}, Discriminator: "type"}
		registered := []core_extensions.Extension{{Point: unusable, Value: "---", Config: &FakeConfig{}}}

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

	It("should reject a config property the extension object does not have", func() {
		moved := fakePoint
		moved.ConfigProperty = "settings"

		err := checkExtensionPoint(specWithExtensionPoint(), moved)

		Expect(err).To(MatchError(ContainSubstring(`has no "settings" property to hold the configuration`)))
	})

	// The points are declared next to the resources but referenced only by a
	// downstream generate run, so without this nothing here notices them drifting
	// from the spec policy-gen produces.
	DescribeTable("should match the generated rest.yaml of the resource that declares it",
		func(point core_extensions.Point, restYAML string) {
			raw, err := os.ReadFile(restYAML)
			Expect(err).ToNot(HaveOccurred())
			var spec map[string]any
			Expect(yaml.Unmarshal(raw, &spec)).To(Succeed())

			Expect(checkExtensionPoint(spec, point)).To(Succeed())
		},
		Entry("HostnameGenerator", hostnamegenerator_api.ExtensionPoint,
			"../../../pkg/core/resources/apis/hostnamegenerator/api/v1alpha1/rest.yaml"),
		Entry("MeshIdentity", meshidentity_api.ExtensionPoint,
			"../../../pkg/core/resources/apis/meshidentity/api/v1alpha1/rest.yaml"),
	)
})

var _ = Describe("propertyPath", func() {
	It("should interleave properties so a schema path addresses a real node", func() {
		Expect(propertyPath([]string{"spec", "provider", "extension"})).To(Equal(
			[]string{"properties", "spec", "properties", "provider", "properties", "extension"}))
	})
})

var _ = Describe("branches", func() {
	// The catch-all matters: without it the spec would reject extension types the
	// control plane accepts, including any a third party adds.
	It("should pair each value with the schema it selects and end with a catch-all", func() {
		result, err := yaml.Marshal(branches(fakePoint, fakeWrappers("acmpca", "vault")))

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(matchers.MatchGoldenYAML(golden("branches.golden.yaml")))
	})

	It("should ignore extensions belonging to another resource", func() {
		other := fakePoint
		other.ResourceType = "OtherResource"
		mixed := append(fakeWrappers("vault"), wrapper{
			ext:        core_extensions.Extension{Point: other, Value: "elsewhere", Config: &FakeConfig{}},
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
		schemas := map[string]any{"FakeResourceRoute53": map[string]any{"type": "object"}}

		expr, err := patchExpression(fakeWrappers("Route53"), schemas)

		Expect(err).ToNot(HaveOccurred())
		Expect(expr).To(matchers.MatchGoldenEqual(golden("patch-expression.golden.yq")))
	})

	// The config is a resource spec like any other, so its own discriminated
	// unions have to be described too.
	It("should describe a union inside the config", func() {
		schemas := map[string]any{"FakeResourceVault": map[string]any{
			"properties": map[string]any{
				"type":   map[string]any{"enum": []any{"Server", "Agent"}},
				"server": map[string]any{"type": "object"},
				"agent":  map[string]any{"type": "object"},
			},
		}}

		expr, err := patchExpression(fakeWrappers("vault"), schemas)

		Expect(err).ToNot(HaveOccurred())
		Expect(expr).To(matchers.MatchGoldenEqual(golden("patch-expression-union.golden.yq")))
	})
})

var _ = Describe("writePackage", func() {
	It("should wrap each config in a CRD root type controller-gen will walk", func() {
		dir := GinkgoT().TempDir()

		Expect(writePackage(dir, fakeWrappers("Route53", "vault"))).To(Succeed())

		Expect(os.ReadFile(filepath.Join(dir, "doc.go"))).To(
			matchers.MatchGoldenEqual(golden("wrapper-package", "doc.go.golden")))
		Expect(os.ReadFile(filepath.Join(dir, "types.go"))).To(
			matchers.MatchGoldenEqual(golden("wrapper-package", "types.go.golden")))
	})
})

var _ = Describe("Generate", func() {
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

		Expect(os.WriteFile(specPath, []byte(specYAML), 0o600)).To(Succeed())

		opts = Options{
			Spec:             specPath,
			WorkDir:          filepath.Join(dir, "work"),
			ControllerGenBin: stubControllerGen(dir),
			YqBin:            stubYq(dir, exprPath),
			Stderr:           GinkgoWriter,
		}
	})

	// This is the property that keeps a downstream product's extensions out of
	// Kuma's spec, so it is asserted rather than left to the empty registry.
	It("should leave the document alone when nothing is registered", func() {
		registerFake()

		Expect(Generate(context.Background(), opts)).To(Succeed())

		Expect(os.ReadFile(specPath)).To(Equal([]byte(specYAML)))
		Expect(exprPath).ToNot(BeAnExistingFile())
	})

	It("should fail when the document it was told to patch is not there", func() {
		registerFake("fake")
		opts.Spec = filepath.Join(dir, "--controller-gen-bin")

		Expect(Generate(context.Background(), opts)).To(MatchError(ContainSubstring("no such file or directory")))
	})

	Context("with a registered extension", func() {
		BeforeEach(func() { registerFake("fake") })

		It("should patch the document with the schema controller-gen produced", func() {
			Expect(Generate(context.Background(), opts)).To(Succeed())

			Expect(os.ReadFile(exprPath)).To(matchers.MatchGoldenEqual(golden("generate.golden.yq")))
		})

		// Extension points live in one document per resource until they are merged,
		// so a document that is about something else is not an error.
		It("should skip a resource this document does not describe", func() {
			elsewhere := fakePoint
			elsewhere.ResourceType = "OtherResource"
			registeredExtensions = func() []core_extensions.Extension {
				return []core_extensions.Extension{{Point: elsewhere, Value: "fake", Config: &FakeConfig{}}}
			}

			Expect(Generate(context.Background(), opts)).To(Succeed())

			Expect(exprPath).ToNot(BeAnExistingFile())
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

		// A document that describes the resource but not the node underneath it is
		// drift, not a document about something else, so it has to be loud.
		It("should fail when the resource is here but the extension point is not", func() {
			drifted := "components:\n  schemas:\n    FakeResourceItem:\n      properties:\n        spec:\n          properties: {}\n"
			Expect(os.WriteFile(specPath, []byte(drifted), 0o600)).To(Succeed())

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
