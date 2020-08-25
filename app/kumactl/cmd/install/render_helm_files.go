package install

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
)

var kumaSystemNamespace = func(namespace string) string {
	return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    kuma.io/system-namespace: "true"
`, namespace)
}

func renderHelmFiles(templates []data.File, args interface{}) ([]data.File, error) {
	loadedChart, err := loadCharts(templates)
	if err != nil {
		return nil, errors.Errorf("Failed to load charts: %s", err)
	}

	overrideValues := generateOverrideValues(args)
	if err := chartutil.ProcessDependencies(loadedChart, overrideValues); err != nil {
		return nil, errors.Errorf("Failed to process dependencies: %s", err)
	}

	namespace := overrideValues["Namespace"].(string)
	options := generateReleaseOptions(loadedChart.Metadata.Name, namespace)

	valuesToRender, err := chartutil.ToRenderValues(loadedChart, overrideValues, options, nil)
	if err != nil {
		return nil, errors.Errorf("Failed to render values: %s", err)
	}

	files, err := engine.Render(loadedChart, valuesToRender)
	if err != nil {
		return nil, errors.Errorf("Failed to render templates: %s", err)
	}
	files["namespace.yaml"] = kumaSystemNamespace(namespace)

	return postRender(loadedChart, files), nil
}

func loadCharts(templates []data.File) (*chart.Chart, error) {
	files := []*loader.BufferedFile{}
	for _, f := range templates {
		files = append(files, &loader.BufferedFile{
			Name: f.FullPath[1:],
			Data: f.Data,
		})
	}

	loadedChart, err := loader.LoadFiles(files)
	if err != nil {
		return nil, err
	}

	// Filter out the pre- templates
	loadedTemplates := loadedChart.Templates
	loadedChart.Templates = []*chart.File{}

	for _, t := range loadedTemplates {
		if !strings.HasPrefix(t.Name, "templates/pre-") {
			loadedChart.Templates = append(loadedChart.Templates, &chart.File{
				Name: t.Name,
				Data: t.Data,
			})
		}
	}

	return loadedChart, nil
}

func generateOverrideValues(args interface{}) map[string]interface{} {
	overrideValues := map[string]interface{}{}

	v := reflect.ValueOf(args)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		value := v.FieldByName(name).Interface()

		splitName := strings.Split(name, "_")
		len := len(splitName)

		root := overrideValues

		for i := 0; i < len-1; i++ {
			n := splitName[i]

			if _, ok := root[n]; !ok {
				root[n] = map[string]interface{}{}
			}
			root = root[n].(map[string]interface{})
		}
		root[splitName[len-1]] = value
	}

	return overrideValues
}

func generateReleaseOptions(name, namespace string) chartutil.ReleaseOptions {
	return chartutil.ReleaseOptions{
		Name:      name,
		Namespace: namespace,
		Revision:  1,
		IsInstall: true,
		IsUpgrade: false,
	}
}

func postRender(loadedChart *chart.Chart, files map[string]string) []data.File {
	result := []data.File{}

	for _, crd := range loadedChart.CRDObjects() {
		result = append(result, data.File{
			Data: crd.File.Data,
			Name: crd.Name,
		})
	}

	for n, d := range files {
		if strings.HasSuffix(n, "yaml") {
			result = append(result, data.File{
				Data: []byte(d),
				Name: n,
			})
		}
	}

	return result
}
