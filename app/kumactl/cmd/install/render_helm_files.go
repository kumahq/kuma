package install

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/client-go/rest"

	"github.com/kumahq/kuma/pkg/util/data"
)

func labelRegex(label string) *regexp.Regexp {
	return regexp.MustCompile("(?m)[\r\n]+^.*" + label + ".*$")
}

var stripLabelsRegexps = []*regexp.Regexp{
	labelRegex("app\\.kubernetes\\.io/managed-by"),
	labelRegex("helm\\.sh/chart"),
	labelRegex("app\\.kubernetes\\.io/version"),
}

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

type onlyWriteWarnings struct {
	writer io.Writer
}

func (w onlyWriteWarnings) Write(p []byte) (int, error) {
	if bytes.HasPrefix(p, []byte("Warning:")) {
		return w.writer.Write(p)
	}
	return io.Discard.Write(p)
}

func renderHelmFiles(
	templates []data.File,
	namespace string,
	overrideValues chartutil.Values,
	kubeClientConfig *rest.Config,
	capabilities chartutil.Capabilities,
) ([]data.File, error) {
	kumaChart, err := loadCharts(templates)
	if err != nil {
		return nil, errors.Errorf("Failed to load charts: %s", err)
	}

	// This is necessary because ProcessDependencies can output warnings as well
	// as trash
	writer := log.Writer()
	log.SetOutput(onlyWriteWarnings{writer: writer})
	if err := chartutil.ProcessDependencies(kumaChart, overrideValues); err != nil {
		return nil, errors.Errorf("Failed to process dependencies: %s", err)
	}
	log.SetOutput(writer)

	options := generateReleaseOptions(kumaChart.Metadata.Name, namespace)

	valuesToRender, err := chartutil.ToRenderValues(kumaChart, overrideValues, options, &capabilities)
	if err != nil {
		return nil, errors.Errorf("Failed to render values: %s", err)
	}

	var files map[string]string
	if kubeClientConfig == nil {
		files, err = engine.Render(kumaChart, valuesToRender)
	} else {
		files, err = engine.RenderWithClient(kumaChart, valuesToRender, kubeClientConfig)
	}
	if err != nil {
		return nil, errors.Errorf("Failed to render templates: %s", err)
	}
	files["namespace.yaml"] = kumaSystemNamespace(namespace)

	return postRender(kumaChart, files), nil
}

func loadCharts(templates []data.File) (*chart.Chart, error) {
	var files []*loader.BufferedFile

	for _, template := range templates {
		files = append(files, &loader.BufferedFile{
			Name: template.FullPath,
			Data: template.Data,
		})
	}

	var fileteredFiles []*loader.BufferedFile
	for _, f := range files {
		if strings.Contains(f.Name, "templates/pre-") || strings.Contains(f.Name, "templates/post-") {
			continue
		}
		fileteredFiles = append(fileteredFiles, f)
	}

	return loader.LoadFiles(fileteredFiles)
}

func generateOverrideValues(args interface{}, helmValuesPrefix string) map[string]interface{} {
	overrideValues := map[string]interface{}{}

	v := reflect.ValueOf(args)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		value := v.FieldByName(name)
		tag := t.Field(i).Tag.Get("helm")

		splitTag := strings.Split(tag, ",")
		if len(splitTag) == 0 {
			continue
		}

		var omitEmpty bool
		for _, tagSplit := range splitTag[0:] {
			omitEmpty = omitEmpty || tagSplit == "omitempty"
		}

		if omitEmpty && value.IsZero() {
			continue
		}

		valuePath := strings.Split(splitTag[0], ".")
		tagCount := len(valuePath)

		root := overrideValues

		for i := 0; i < tagCount-1; i++ {
			n := valuePath[i]

			if _, ok := root[n]; !ok {
				root[n] = map[string]interface{}{}
			}
			root = root[n].(map[string]interface{})
		}
		root[valuePath[tagCount-1]] = adjustType(value.Interface())
	}

	if helmValuesPrefix != "" {
		return map[string]interface{}{
			helmValuesPrefix: overrideValues,
		}
	}

	return overrideValues
}

// If the parameter value is map it has to be of a type map[string]interface{} therefore we need to convert it
func adjustType(value interface{}) interface{} {
	if m, ok := value.(map[string]string); ok {
		result := map[string]interface{}{}
		for k, v := range m {
			result[k] = v
		}
		return result
	}
	return value
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

	// sorted map of files to ensure consistency of the output
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if strings.HasSuffix(k, "yaml") {
			content := files[k]

			// strip Helm Chart specific labels
			for _, stripRegEx := range stripLabelsRegexps {
				content = stripRegEx.ReplaceAllString(content, "")
			}

			result = append(result, data.File{
				Data: []byte(content),
				Name: k,
			})
		}
	}

	return result
}
