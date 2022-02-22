package auth

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func AuthUniversal() {
	var cluster Cluster

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)

		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should generate user for group admin and log in", func() {
		// given
		token, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("generate", "user-token",
			"--name", "new-admin",
			"--group", "mesh-system:admin",
			"--valid-for", "24h",
		)
		Expect(err).ToNot(HaveOccurred())

		// when kumactl is configured with new token
		err = cluster.GetKumactlOptions().KumactlConfigControlPlanesAdd(
			"test-admin",
			cluster.GetKuma().GetAPIServerAddress(),
			token,
		)
		Expect(err).ToNot(HaveOccurred())

		// then the new admin can access secrets
		kumactl, err := NewKumactlOptions(cluster.GetTesting(), cluster.GetKuma().GetName(), false)
		Expect(err).ToNot(HaveOccurred())
		Expect(kumactl.RunKumactl("get", "secrets")).To(Succeed())
	})

	It("should generate user for group member and log in", func() {
		// given
		token, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("generate", "user-token",
			"--name", "team-a-member",
			"--group", "team-a",
			"--valid-for", "24h",
		)
		Expect(err).ToNot(HaveOccurred())

		// when kumactl is configured with new token
		err = cluster.GetKumactlOptions().KumactlConfigControlPlanesAdd(
			"test-user",
			cluster.GetKuma().GetAPIServerAddress(),
			token,
		)
		Expect(err).ToNot(HaveOccurred())

		// then the new member can access dataplanes but not secrets because they are not admin
		kumactl, err := NewKumactlOptions(cluster.GetTesting(), cluster.GetKuma().GetName(), false)
		Expect(err).ToNot(HaveOccurred())
		Expect(kumactl.RunKumactl("get", "dataplanes")).To(Succeed())
		Expect(kumactl.RunKumactl("get", "secrets")).ToNot(Succeed())
	})
}
