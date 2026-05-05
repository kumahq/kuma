package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/envs/multizone"
	"github.com/kumahq/kuma/v2/test/framework/utils"
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
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Sprintf("%s %s\n\nbody read error: %v\n", resp.Proto, resp.Status, err)
		}

		return fmt.Sprintf("%s %s\n\n%s\n", resp.Proto, resp.Status, strings.TrimRight(string(body), "\n"))
	}

	// applyKube renders the YAML body with Go templates and pipes it into
	// `kubectl apply -f -` so output is stable across runs (no temp file paths
	// leak into errors). The error format mirrors terratest's ErrWithCmdOutput
	// to keep golden files compatible.
	applyKube := func(cluster Cluster, yamlBody string) string {
		rendered := utils.FromTemplate(Default, yamlBody, Config)
		opts := cluster.GetKubectlOptions()
		args := []string{}
		if opts.ContextName != "" {
			args = append(args, "--context", opts.ContextName)
		}
		if opts.ConfigPath != "" {
			args = append(args, "--kubeconfig", opts.ConfigPath)
		}
		if opts.Namespace != "" {
			args = append(args, "--namespace", opts.Namespace)
		}
		args = append(args, "apply", "-f", "-")
		cmd := exec.CommandContext(context.Background(), "kubectl", args...)
		cmd.Stdin = strings.NewReader(rendered)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		var out string
		if err := cmd.Run(); err != nil {
			out = fmt.Sprintf("error while running command: %v; %s", err, stderr.String())
		} else {
			out = stdout.String()
		}
		return strings.TrimRight(out, "\n") + "\n"
	}

	apiPaths := map[string]string{
		"meshtimeout":         "meshtimeouts",
		"meshexternalservice": "meshexternalservices",
	}

	// Filename convention: "<target>.<resource>.<case>.input.yaml".
	DescribeTable("validates labels", func(inputFile string) {
		parts := strings.Split(filepath.Base(inputFile), ".")
		Expect(len(parts)).To(BeNumerically(">=", 4), "unexpected filename: %s", inputFile)
		targetSlug, resourceSlug := parts[0], parts[1]

		apiPath, ok := apiPaths[resourceSlug]
		Expect(ok).To(BeTrue(), "unknown resource slug %q in %s", resourceSlug, inputFile)

		body, err := os.ReadFile(inputFile)
		Expect(err).ToNot(HaveOccurred())

		var blob string
		switch targetSlug {
		case "global":
			blob = applyUniversal(multizone.Global, apiPath, string(body))
		case "zone-uni":
			blob = applyUniversal(multizone.UniZone1, apiPath, string(body))
		case "zone-k8s":
			blob = applyKube(multizone.KubeZone1, string(body))
		default:
			Fail(fmt.Sprintf("unknown target slug %q in %s", targetSlug, inputFile))
		}

		goldenPath := strings.Replace(inputFile, ".input.yaml", ".golden.yaml", 1)
		Expect(blob).To(matchers.MatchGoldenEqual(goldenPath))
	}, test.EntriesForFolder("", "validation"))
}
