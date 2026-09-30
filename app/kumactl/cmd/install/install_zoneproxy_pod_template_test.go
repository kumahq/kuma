package install_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	chartcommon "helm.sh/helm/v4/pkg/chart/common"
	chartcommonutil "helm.sh/helm/v4/pkg/chart/common/util"
	"helm.sh/helm/v4/pkg/chart/loader/archive"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
	"helm.sh/helm/v4/pkg/engine"
	kube_apps "k8s.io/api/apps/v1"
	kube_core "k8s.io/api/core/v1"
	k8s_yaml "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/kumahq/kuma/v3/deployments"
	"github.com/kumahq/kuma/v3/pkg/util/data"
)

const zoneProxyDeploymentTemplate = "kuma/templates/mesh-zoneproxy-deployment.yaml"

var zoneProxyValues = map[string]any{
	"meshes": []any{
		map[string]any{
			"name":    "default",
			"ingress": map[string]any{"enabled": true},
			"egress":  map[string]any{"enabled": true},
		},
	},
}

// renderChartAtVersion renders the chart the way helm does, keeping the
// version labels kumactl strips, so a version leaking into a pod template shows.
func renderChartAtVersion(version string) map[string]string {
	templateFiles, err := data.ReadFiles(deployments.KumaChartFS())
	Expect(err).ToNot(HaveOccurred())

	var files []*archive.BufferedFile
	for _, f := range templateFiles {
		files = append(files, &archive.BufferedFile{Name: f.FullPath, Data: f.Data})
	}
	kumaChart, err := loader.LoadFiles(files)
	Expect(err).ToNot(HaveOccurred())
	kumaChart.Metadata.Version = version
	kumaChart.Metadata.AppVersion = version

	options := chartcommon.ReleaseOptions{
		Name:      kumaChart.Metadata.Name,
		Namespace: "kuma-system",
		Revision:  1,
		IsUpgrade: true,
	}
	valuesToRender, err := chartcommonutil.ToRenderValues(kumaChart, zoneProxyValues, options, chartcommon.DefaultCapabilities)
	Expect(err).ToNot(HaveOccurred())

	rendered, err := engine.Render(kumaChart, valuesToRender)
	Expect(err).ToNot(HaveOccurred())
	return rendered
}

func decodeDocuments[T any](content string) []T {
	var docs []T
	decoder := k8s_yaml.NewYAMLOrJSONDecoder(strings.NewReader(content), 4096)
	for {
		var doc T
		err := decoder.Decode(&doc)
		if errors.Is(err, io.EOF) {
			return docs
		}
		Expect(err).ToNot(HaveOccurred())
		docs = append(docs, doc)
	}
}

func zoneProxyDeployments(rendered map[string]string) []kube_apps.Deployment {
	var result []kube_apps.Deployment
	for _, d := range decodeDocuments[kube_apps.Deployment](rendered[zoneProxyDeploymentTemplate]) {
		if d.Kind == "Deployment" {
			result = append(result, d)
		}
	}
	Expect(result).To(HaveLen(2))
	return result
}

// referencedConfigs lists the ConfigMaps and Secrets a pod reads at start,
// keyed as kind/name.
func referencedConfigs(spec kube_core.PodSpec) []string {
	var refs []string
	for _, v := range spec.Volumes {
		if v.ConfigMap != nil {
			refs = append(refs, "ConfigMap/"+v.ConfigMap.Name)
		}
		if v.Secret != nil {
			refs = append(refs, "Secret/"+v.Secret.SecretName)
		}
		if v.Projected != nil {
			for _, s := range v.Projected.Sources {
				if s.ConfigMap != nil {
					refs = append(refs, "ConfigMap/"+s.ConfigMap.Name)
				}
				if s.Secret != nil {
					refs = append(refs, "Secret/"+s.Secret.Name)
				}
			}
		}
	}
	for _, c := range append(spec.InitContainers, spec.Containers...) {
		for _, e := range c.EnvFrom {
			if e.ConfigMapRef != nil {
				refs = append(refs, "ConfigMap/"+e.ConfigMapRef.Name)
			}
			if e.SecretRef != nil {
				refs = append(refs, "Secret/"+e.SecretRef.Name)
			}
		}
		for _, e := range c.Env {
			if e.ValueFrom == nil {
				continue
			}
			if e.ValueFrom.ConfigMapKeyRef != nil {
				refs = append(refs, "ConfigMap/"+e.ValueFrom.ConfigMapKeyRef.Name)
			}
			if e.ValueFrom.SecretKeyRef != nil {
				refs = append(refs, "Secret/"+e.ValueFrom.SecretKeyRef.Name)
			}
		}
	}
	return refs
}

var _ = Context("mesh-scoped zone proxy pod template", func() {
	It("should not change when the chart version changes", func() {
		before := zoneProxyDeployments(renderChartAtVersion("1.0.0"))
		after := zoneProxyDeployments(renderChartAtVersion("2.0.0"))

		for i := range before {
			Expect(after[i].Spec.Template).To(Equal(before[i].Spec.Template),
				"pod template of %s depends on the chart version, so every upgrade restarts it", before[i].Name)
		}
	})

	It("should carry a checksum of every chart-rendered ConfigMap and Secret it reads", func() {
		rendered := renderChartAtVersion("1.0.0")

		type object struct {
			Kind     string `json:"kind"`
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		}
		renderedIn := map[string]string{}
		for file, content := range rendered {
			if !strings.HasSuffix(file, ".yaml") {
				continue
			}
			for _, obj := range decodeDocuments[object](content) {
				if obj.Kind == "ConfigMap" || obj.Kind == "Secret" {
					renderedIn[obj.Kind+"/"+obj.Metadata.Name] = file
				}
			}
		}

		for _, deployment := range zoneProxyDeployments(rendered) {
			checksums := map[string]struct{}{}
			for key, value := range deployment.Spec.Template.Annotations {
				if strings.HasPrefix(key, "checksum/") {
					checksums[value] = struct{}{}
				}
			}
			for _, ref := range referencedConfigs(deployment.Spec.Template.Spec) {
				file, ok := renderedIn[ref]
				if !ok {
					continue
				}
				sum := sha256.Sum256([]byte(rendered[file]))
				Expect(checksums).To(HaveKey(hex.EncodeToString(sum[:])),
					"%s reads %s but has no checksum/ annotation of %s, so changing it does not restart the pods", deployment.Name, ref, file)
			}
		}
	})
})
