package auth

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func UserAuth() {
	It("should generate user for group admin and log in", func() {
		// given
		token, err := universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("generate", "user-token",
			"--name", "new-admin",
			"--group", "mesh-system:admin",
			"--valid-for", "24h",
		)
		Expect(err).ToNot(HaveOccurred())

		// when kumactl is configured with new token
		kumactl := NewKumactlOptionsE2E(universal.Cluster.GetTesting(), universal.Cluster.GetKuma().GetName()+"test-admin", false)
		err = kumactl.KumactlConfigControlPlanesAdd(
			"test-admin",
			universal.Cluster.GetKuma().GetAPIServerAddress(),
			token,
			nil,
		)

		// then the new admin can access secrets
		Expect(err).ToNot(HaveOccurred())
		Expect(kumactl.RunKumactl("get", "secrets")).To(Succeed())
	})

	It("should generate user for group member and log in", func() {
		// given
		token, err := universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("generate", "user-token",
			"--name", "team-a-member",
			"--group", "team-a",
			"--valid-for", "24h",
		)
		Expect(err).ToNot(HaveOccurred())

		// when kumactl is configured with new token
		kumactl := NewKumactlOptionsE2E(universal.Cluster.GetTesting(), universal.Cluster.GetKuma().GetName()+"test-user", false)
		err = kumactl.KumactlConfigControlPlanesAdd(
			"test-user",
			universal.Cluster.GetKuma().GetAPIServerAddress(),
			token,
			nil,
		)

		// then the new member can access dataplanes but not secrets because they are not admin
		Expect(err).ToNot(HaveOccurred())
		Expect(kumactl.RunKumactl("get", "dataplanes")).To(Succeed())
		Expect(kumactl.RunKumactl("get", "secrets")).ToNot(Succeed())
	})

	DescribeTable("should ignore auth data on unauthorized endpoints",
		func(endpoint string) {
			// given
			req, err := http.NewRequest(http.MethodGet, universal.Cluster.GetKuma().GetAPIServerAddress()+endpoint, http.NoBody)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("authorization", "Bearer invliddata")

			// when
			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(200))
		},
		Entry("index", "/"),
		Entry("gui", "/gui/"),
	)
}
