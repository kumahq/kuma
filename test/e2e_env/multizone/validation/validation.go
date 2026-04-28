package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/envs/multizone"
)

func ResourceValidation() {
	const mesh = "multizone-label-validation"

	BeforeAll(func() {
		Expect(multizone.Global.Install(MTLSMeshUniversal(mesh))).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, mesh)
		DebugUniversal(multizone.UniZone1, mesh)
		DebugKube(multizone.KubeZone1, mesh, "default")
	})

	E2EAfterAll(func() {
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())
	})

	// applyUniversal sends HTTP PUT to the kuma-cp REST API and returns the
	// raw response as "<proto> <status>\n\n<body>".
	applyUniversal := func(cluster Cluster, resource, yamlBody string) string {
		jsonBody, err := yaml.YAMLToJSON([]byte(yamlBody))
		if err != nil {
			return fmt.Sprintf("yaml-to-json error: %v", err)
		}

		var meta struct {
			Name string `json:"name"`
			Mesh string `json:"mesh"`
		}
		if err := json.Unmarshal(jsonBody, &meta); err != nil {
			return fmt.Sprintf("parse name/mesh error: %v", err)
		}

		url := fmt.Sprintf("%s/meshes/%s/%s/%s",
			cluster.GetKuma().GetAPIServerAddress(),
			meta.Mesh, resource, meta.Name)
		req, err := http.NewRequestWithContext(context.Background(),
			http.MethodPut, url, bytes.NewReader(jsonBody))
		if err != nil {
			return fmt.Sprintf("request create error: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Sprintf("request error: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)

		return fmt.Sprintf("%s %s\n\n%s", resp.Proto, resp.Status, strings.TrimRight(string(body), "\n"))
	}

	// applyKube applies a K8s-format YAML file via kubectl and returns stdout
	// on success or err.Error() verbatim on failure.
	applyKube := func(cluster Cluster, inputPath string) string {
		out, err := k8s.RunKubectlAndGetOutputE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(),
			"apply", "-f", inputPath,
		)
		if err != nil {
			return err.Error()
		}
		return out
	}

	type target struct {
		slug  string
		apply func(resource, inputPath, yamlBody string) string
	}

	globalT := target{
		slug: "global",
		apply: func(resource, _, yamlBody string) string {
			return applyUniversal(multizone.Global, resource, yamlBody)
		},
	}
	uniZoneT := target{
		slug: "zone-uni",
		apply: func(resource, _, yamlBody string) string {
			return applyUniversal(multizone.UniZone1, resource, yamlBody)
		},
	}
	kubeT := target{
		slug: "zone-k8s",
		apply: func(_, inputPath, _ string) string {
			return applyKube(multizone.KubeZone1, inputPath)
		},
	}

	// resourceSlug matches the testdata filename segment; apiPath is the REST
	// plural path used in HTTP PUT URLs.
	runPair := func(t target, resourceSlug, apiPath string) {
		// CWD when ginkgo runs the suite is the suite package dir
		// (test/e2e_env/multizone), so paths must include the sub-package prefix.
		pattern := filepath.Join("validation", "testdata", t.slug+"."+resourceSlug+".*.input.yaml")
		inputs, err := filepath.Glob(pattern)
		Expect(err).ToNot(HaveOccurred())
		Expect(inputs).ToNot(BeEmpty(), "no input files matched %s", pattern)

		type result struct {
			path string
			blob string
		}
		results := make([]result, len(inputs))

		g := errgroup.Group{}
		for i, path := range inputs {
			g.Go(func() error {
				body, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				results[i] = result{path: path, blob: t.apply(apiPath, path, string(body))}
				return nil
			})
		}
		Expect(g.Wait()).To(Succeed())

		for _, r := range results {
			goldenPath := strings.Replace(r.path, ".input.yaml", ".golden.yaml", 1)
			Expect(r.blob).To(matchers.MatchGoldenEqual(goldenPath),
				"case: %s", filepath.Base(r.path))
		}
	}

	It("Global CP / MeshTimeout", func() { runPair(globalT, "meshtimeout", "meshtimeouts") })
	It("Global CP / MeshExternalService", func() { runPair(globalT, "meshexternalservice", "meshexternalservices") })
	It("UniZone CP / MeshTimeout", func() { runPair(uniZoneT, "meshtimeout", "meshtimeouts") })
	It("UniZone CP / MeshExternalService", func() { runPair(uniZoneT, "meshexternalservice", "meshexternalservices") })
	It("KubeZone CP / MeshTimeout", func() { runPair(kubeT, "meshtimeout", "meshtimeouts") })
	It("KubeZone CP / MeshExternalService", func() { runPair(kubeT, "meshexternalservice", "meshexternalservices") })
}
