package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
			blob = ApplyResourceRawResponse(multizone.Global, apiPath, string(body))
		case "zone-uni":
			blob = ApplyResourceRawResponse(multizone.UniZone1, apiPath, string(body))
		case "zone-k8s":
			rendered := utils.FromTemplate(Default, string(body), Config)
			tmpfile, terr := k8s.StoreConfigToTempFileE(multizone.KubeZone1.GetTesting(), rendered)
			Expect(terr).ToNot(HaveOccurred())
			defer os.Remove(tmpfile)
			out, kerr := k8s.RunKubectlAndGetOutputE(multizone.KubeZone1.GetTesting(), multizone.KubeZone1.GetKubectlOptions(), "apply", "-f", tmpfile)
			if kerr != nil {
				out = kerr.Error()
			}
			blob = strings.TrimRight(out, "\n") + "\n"
		default:
			Fail(fmt.Sprintf("unknown target slug %q in %s", targetSlug, inputFile))
		}

		goldenPath := strings.Replace(inputFile, ".input.yaml", ".golden.yaml", 1)
		Expect(blob).To(matchers.MatchGoldenEqual(goldenPath))
	}, test.EntriesForFolder("", "validation"))
}
