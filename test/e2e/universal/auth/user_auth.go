package auth

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func UserAuth() {
	It("should generate user for group admin and log in", func() {
		// given
		token, err := env.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("generate", "user-token",
			"--name", "new-admin",
			"--group", "mesh-system:admin",
			"--valid-for", "24h",
		)
		Expect(err).ToNot(HaveOccurred())

		// when kumactl is configured with new token
		err = env.Cluster.GetKumactlOptions().KumactlConfigControlPlanesAdd(
			"test-admin",
			env.Cluster.GetKuma().GetAPIServerAddress(),
			token,
		)
		Expect(err).ToNot(HaveOccurred())

		// then the new admin can access secrets
		kumactl, err := NewKumactlOptions(env.Cluster.GetTesting(), env.Cluster.GetKuma().GetName(), false)
		Expect(err).ToNot(HaveOccurred())
		Expect(kumactl.RunKumactl("get", "secrets")).To(Succeed())
	})

	It("should generate user for group member and log in", func() {
		// given
		token, err := env.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("generate", "user-token",
			"--name", "team-a-member",
			"--group", "team-a",
			"--valid-for", "24h",
		)
		Expect(err).ToNot(HaveOccurred())

		// when kumactl is configured with new token
		err = env.Cluster.GetKumactlOptions().KumactlConfigControlPlanesAdd(
			"test-user",
			env.Cluster.GetKuma().GetAPIServerAddress(),
			token,
		)
		Expect(err).ToNot(HaveOccurred())

		// then the new member can access dataplanes but not secrets because they are not admin
		kumactl, err := NewKumactlOptions(env.Cluster.GetTesting(), env.Cluster.GetKuma().GetName(), false)
		Expect(err).ToNot(HaveOccurred())
		Expect(kumactl.RunKumactl("get", "dataplanes")).To(Succeed())
		Expect(kumactl.RunKumactl("get", "secrets")).ToNot(Succeed())
	})
}
