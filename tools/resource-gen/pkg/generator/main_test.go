package generator_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/tools/resource-gen/pkg/generator"
)

var _ = Describe("Module Version", func() {
	Describe("getModuleVersion", func() {
		It("should extract version from go.mod", func() {
			// When
			moduleVersion, err := generator.GetModuleVersion()

			// Then
			Expect(err).ToNot(HaveOccurred())

			// Find and read go.mod to verify the result
			goModPath := findGoMod()
			content, err := os.ReadFile(goModPath)
			Expect(err).ToNot(HaveOccurred())

			// Extract expected version from go.mod
			lines := strings.Split(string(content), "\n")
			Expect(lines).ToNot(BeEmpty())

			moduleLine := strings.TrimSpace(lines[0])
			Expect(moduleLine).To(HavePrefix("module "))

			modulePath := strings.TrimPrefix(moduleLine, "module ")
			modulePath = strings.TrimSpace(modulePath)

			// Extract expected version suffix
			parts := strings.Split(modulePath, "/")
			var expectedVersion string
			if len(parts) > 0 {
				lastPart := parts[len(parts)-1]
				if strings.HasPrefix(lastPart, "v") {
					expectedVersion = "/" + lastPart
				}
			}

			// Verify the function returned the correct version
			Expect(moduleVersion).To(Equal(expectedVersion), "Module path: %s", modulePath)

			// Additional validation: ensure it's a valid version format
			if moduleVersion != "" {
				Expect(moduleVersion).To(HavePrefix("/v"))
			}
		})

		It("should return valid version format", func() {
			// When
			version, err := generator.GetModuleVersion()

			// Then
			Expect(err).ToNot(HaveOccurred())

			// Should be either empty or start with "/v"
			if version != "" {
				Expect(version).To(HavePrefix("/v"))
			}
		})

		It("should cache the result on subsequent calls", func() {
			// When
			version1, err1 := generator.GetModuleVersion()
			version2, err2 := generator.GetModuleVersion()

			// Then
			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
			Expect(version1).To(Equal(version2))
		})
	})
})

func findGoMod() string {
	// Start from current directory and walk up to find go.mod
	dir, err := os.Getwd()
	Expect(err).ToNot(HaveOccurred())

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return goModPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod
			Fail("Could not find go.mod in any parent directory")
		}
		dir = parent
	}
}
