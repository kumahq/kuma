package labels_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/v3/pkg/core/resources/labels"
)

var _ = Describe("AllComputedLabels", func() {
	// The OpenAPI spec documents every computed label under Meta.labels. Nothing
	// forces the two to agree, so they drift silently as labels come and go.
	It("should match the labels documented in the OpenAPI spec", func() {
		specPath := filepath.Join("..", "..", "..", "..", "api", "openapi", "specs", "common", "resource.yaml")
		raw, err := os.ReadFile(specPath)
		Expect(err).ToNot(HaveOccurred())

		spec := struct {
			Components struct {
				Schemas struct {
					Meta struct {
						Properties struct {
							Labels struct {
								PatternProperties map[string]any `json:"patternProperties"`
							} `json:"labels"`
						} `json:"properties"`
					} `json:"Meta"`
				} `json:"schemas"`
			} `json:"components"`
		}{}
		Expect(yaml.Unmarshal(raw, &spec)).To(Succeed())

		var documented []string
		for pattern := range spec.Components.Schemas.Meta.Properties.Labels.PatternProperties {
			// patterns are anchored exact matches, e.g. `^kuma\.io/zone$`
			label := strings.NewReplacer("^", "", "$", "", `\.`, ".").Replace(pattern)
			documented = append(documented, label)
		}

		var computed []string
		for label := range labels.AllComputedLabels {
			computed = append(computed, label)
		}

		Expect(documented).To(ConsistOf(computed))
	})
})
