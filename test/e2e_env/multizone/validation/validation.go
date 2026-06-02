package validation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/kds/hash"
	meshtimeout_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/api/v1alpha1"
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
	DescribeTable("validates labels on apply", func(inputFile string) {
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
			out, kerr := RunKubectlWithStdinAndGetOutputE(context.Background(), multizone.KubeZone1.GetTesting(), multizone.KubeZone1.GetKubectlOptions(), rendered, "apply", "-f", "-")
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

	Describe("validates origin on delete", func() {
		const mtName = "mt-delete-origin"

		mtUniversal := func(origin string) string {
			return fmt.Sprintf(`
type: MeshTimeout
name: %s
mesh: %s
labels:
  kuma.io/origin: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
      default:
        connectionTimeout: 5s
`, mtName, mesh, origin)
		}

		mtK8s := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
      default:
        connectionTimeout: 5s
`, mtName, Config.KumaNamespace, mesh)

		descriptor := meshtimeout_api.MeshTimeoutResourceTypeDescriptor

		// Names after KDS sync:
		//   global -> zones: hash(mesh, name) on Universal; same + ".<ns>" in the
		//     kuma-API on K8s; the raw K8s object name is just the hash.
		//   zones -> global: hash(mesh, name, zoneName).
		nameFromGlobal := hash.HashedName(mesh, mtName)
		nameOnGlobalFromUni := hash.HashedName(mesh, mtName, Kuma4)
		nameOnGlobalFromK8s := hash.HashedName(mesh, mtName, Kuma1, Config.KumaNamespace)
		nameOnK8sZone := fmt.Sprintf("%s.%s", nameFromGlobal, Config.KumaNamespace)

		deleteFromK8sZone := func(k8sName string) (string, error) {
			return k8s.RunKubectlAndGetOutputContextE(
				multizone.KubeZone1.GetTesting(),
				context.Background(),
				multizone.KubeZone1.GetKubectlOptions(Config.KumaNamespace),
				"delete", "meshtimeout", k8sName,
			)
		}

		It("rejects deletion of global-originated resource from zones", func() {
			Expect(ApplyResourceRawResponse(multizone.Global, "meshtimeouts", mtUniversal("global"))).
				To(ContainSubstring("201 Created"))
			Expect(WaitForResource(descriptor, core_model.ResourceKey{Mesh: mesh, Name: nameFromGlobal}, multizone.UniZone1)).To(Succeed())
			Expect(WaitForResource(descriptor, core_model.ResourceKey{Mesh: mesh, Name: nameOnK8sZone}, multizone.KubeZone1)).To(Succeed())

			Expect(DeleteResourceRawResponse(multizone.UniZone1, "meshtimeouts", mesh, nameFromGlobal)).
				To(ContainSubstring("kuma.io/origin should be 'zone', got 'global'"))

			out, kerr := deleteFromK8sZone(nameFromGlobal)
			Expect(kerr).To(HaveOccurred(), "kubectl delete should fail: %s", out)
			Expect(out + kerr.Error()).To(ContainSubstring("Operation not allowed"))

			Expect(DeleteResourceRawResponse(multizone.Global, "meshtimeouts", mesh, mtName)).
				To(ContainSubstring("200 OK"))
		})

		It("rejects deletion of zone-k8s-originated resource from global", func() {
			rendered := utils.FromTemplate(Default, mtK8s, Config)
			_, err := RunKubectlWithStdinAndGetOutputE(context.Background(),
				multizone.KubeZone1.GetTesting(), multizone.KubeZone1.GetKubectlOptions(),
				rendered, "apply", "-f", "-")
			Expect(err).ToNot(HaveOccurred())
			Expect(WaitForResource(descriptor, core_model.ResourceKey{Mesh: mesh, Name: nameOnGlobalFromK8s}, multizone.Global)).To(Succeed())

			Expect(DeleteResourceRawResponse(multizone.Global, "meshtimeouts", mesh, nameOnGlobalFromK8s)).
				To(ContainSubstring("kuma.io/origin should be 'global', got 'zone'"))

			out, kerr := deleteFromK8sZone(mtName)
			Expect(kerr).ToNot(HaveOccurred(), "kubectl delete on zone of origin should succeed: %s", out)
		})

		It("rejects deletion of zone-uni-originated resource from global", func() {
			Expect(ApplyResourceRawResponse(multizone.UniZone1, "meshtimeouts", mtUniversal("zone"))).
				To(ContainSubstring("201 Created"))
			Expect(WaitForResource(descriptor, core_model.ResourceKey{Mesh: mesh, Name: nameOnGlobalFromUni}, multizone.Global)).To(Succeed())

			Expect(DeleteResourceRawResponse(multizone.Global, "meshtimeouts", mesh, nameOnGlobalFromUni)).
				To(ContainSubstring("kuma.io/origin should be 'global', got 'zone'"))

			Expect(DeleteResourceRawResponse(multizone.UniZone1, "meshtimeouts", mesh, mtName)).
				To(ContainSubstring("200 OK"))
		})
	})
}
