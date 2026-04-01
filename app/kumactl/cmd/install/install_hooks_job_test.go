package install_test

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/v2/deployments"
	"github.com/kumahq/kuma/v2/pkg/util/data"
)

// hookStripLabels strips Helm-managed labels from rendered content, matching the
// behavior of the main renderHelmFiles pipeline.
var hookStripLabels = []*regexp.Regexp{
	regexp.MustCompile(`(?m)[\r\n]+.*app\.kubernetes\.io/managed-by.*`),
	regexp.MustCompile(`(?m)[\r\n]+.*helm\.sh/chart.*`),
	regexp.MustCompile(`(?m)[\r\n]+.*app\.kubernetes\.io/version.*`),
}

// renderHookTemplates renders only the pre-* and post-* Helm hook templates,
// which are excluded from the normal kumactl install control-plane output.
func renderHookTemplates(valuesFile string) ([]byte, error) {
	templateFiles, err := data.ReadFiles(deployments.KumaChartFS())
	if err != nil {
		return nil, err
	}

	overrideValues := map[string]any{}
	if valuesFile != "" {
		valBytes, err := os.ReadFile(valuesFile)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(valBytes, &overrideValues); err != nil {
			return nil, err
		}
	}

	// Load chart including pre-* / post-* templates (no filtering).
	var files []*loader.BufferedFile
	for _, f := range templateFiles {
		files = append(files, &loader.BufferedFile{
			Name: f.FullPath,
			Data: f.Data,
		})
	}

	kumaChart, err := loader.LoadFiles(files)
	if err != nil {
		return nil, err
	}

	if err := chartutil.ProcessDependencies(kumaChart, overrideValues); err != nil {
		return nil, err
	}

	options := chartutil.ReleaseOptions{
		Name:      kumaChart.Metadata.Name,
		Namespace: "kuma-system",
		Revision:  1,
		IsInstall: true,
	}

	valuesToRender, err := chartutil.ToRenderValues(kumaChart, overrideValues, options, chartutil.DefaultCapabilities)
	if err != nil {
		return nil, err
	}

	rendered, err := engine.Render(kumaChart, valuesToRender)
	if err != nil {
		return nil, err
	}

	// Collect only hook template files, sorted for deterministic output.
	var hookKeys []string
	for k := range rendered {
		if strings.Contains(k, "/templates/pre-") || strings.Contains(k, "/templates/post-") {
			hookKeys = append(hookKeys, k)
		}
	}
	sort.Strings(hookKeys)

	var parts []string
	for _, k := range hookKeys {
		content := rendered[k]
		for _, re := range hookStripLabels {
			content = re.ReplaceAllString(content, "")
		}
		if strings.TrimSpace(content) != "" {
			parts = append(parts, strings.TrimSpace(content))
		}
	}

	out := []byte(strings.Join(parts, "\n---\n") + "\n")
	return out, nil
}

var _ = Context("kumactl install control-plane hook jobs", func() {
	DescribeTable("should render hook job templates", func(valuesFile string) {
		rendered, err := renderHookTemplates(valuesFile)
		Expect(err).ToNot(HaveOccurred())

		goldenFile := strings.Replace(valuesFile, ".values.yaml", ".golden.yaml", 1)
		ExpectMatchesGoldenFiles(rendered, goldenFile)
	}, func() []TableEntry {
		var res []TableEntry
		testDir := filepath.Join("testdata", "install-cp-helm-hooks")
		files, err := os.ReadDir(testDir)
		Expect(err).ToNot(HaveOccurred())
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".values.yaml") {
				res = append(res, Entry(f.Name(), path.Join(testDir, f.Name())))
			}
		}
		return res
	}())
})
