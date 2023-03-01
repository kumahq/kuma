package auth

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func OfflineAuth() {
	meshName := "offline-auth"

	var universal Cluster

	cpCfg := `
apiServer:
  authn:
    type: tokens
    tokens:
      enableIssuer: true
      validator:
        useSecrets: true
        publicKeys:
        - kid: static-1
          key: |
            -----BEGIN RSA PUBLIC KEY-----
            MIIBCgKCAQEAqwbFZ7LSuRGEkFPsZOLYuimsjDeie4sdtqIVW9bLDrTSql+o2sBL
            wt22MJ897/oq7+jZhVlENE1ddAKdFSWv3nhOI/XK9VJt7qNudcoC9252XrycIi5h
            i700CDgdRgRt+2paZiRCgc5afNMHJmVIp2d2lQTUKn/pQGlqY4ufuA3U1z+8t++k
            oGnj0sKIcXzqa5ZZxZ/81khp0e0Ze7llTmfEU3gQXu/Coa2y7LEUHdrNalM3si0v
            FvX0KmBtADEJ4n9Jo4ja3hDmp83Q4KjJq0xKbhh9Fp3AjwjDb0fVFwbt+8SdVgyV
            5PE+7HdigwlJ/cOVb9IY/UKVgCzlW5inCQIDAQAB
            -----END RSA PUBLIC KEY-----
dpServer:
  authn:
    dpProxy:
      type: dpToken
      dpToken:
        enableIssuer: false
        validator:
          useSecrets: false
          publicKeys:
          - kid: static-1
            mesh: offline-auth
            key: |
              -----BEGIN RSA PUBLIC KEY-----
              MIIBCgKCAQEAqwbFZ7LSuRGEkFPsZOLYuimsjDeie4sdtqIVW9bLDrTSql+o2sBL
              wt22MJ897/oq7+jZhVlENE1ddAKdFSWv3nhOI/XK9VJt7qNudcoC9252XrycIi5h
              i700CDgdRgRt+2paZiRCgc5afNMHJmVIp2d2lQTUKn/pQGlqY4ufuA3U1z+8t++k
              oGnj0sKIcXzqa5ZZxZ/81khp0e0Ze7llTmfEU3gQXu/Coa2y7LEUHdrNalM3si0v
              FvX0KmBtADEJ4n9Jo4ja3hDmp83Q4KjJq0xKbhh9Fp3AjwjDb0fVFwbt+8SdVgyV
              5PE+7HdigwlJ/cOVb9IY/UKVgCzlW5inCQIDAQAB
              -----END RSA PUBLIC KEY-----
`

	BeforeAll(func() {
		universal = NewUniversalCluster(NewTestingT(), "kuma-offline-auth", Silent)
		Expect(NewClusterSetup().
			Install(Kuma(core.Standalone,
				WithYamlConfig(cpCfg),
			)).
			Install(MeshUniversal(meshName)).
			Setup(universal)).To(Succeed())
	})

	AfterAll(func() {
		Expect(universal.DismissCluster()).To(Succeed())
	})

	It("should use user-token generated offline", func() {
		// given
		token, err := universal.GetKumactlOptions().RunKumactlAndGetOutput("generate", "user-token",
			"--name", "new-admin",
			"--group", "mesh-system:admin",
			"--valid-for", "24h",
			"--kid", "static-1",
			"--signing-key-path", filepath.Join("..", "..", "keys", "samplekey.pem"),
		)
		Expect(err).ToNot(HaveOccurred())

		// when kumactl is configured with new token
		kumactl := NewKumactlOptions(universal.GetTesting(), universal.GetKuma().GetName()+"test-admin", false)
		err = kumactl.KumactlConfigControlPlanesAdd(
			"test-admin",
			universal.GetKuma().GetAPIServerAddress(),
			token,
		)

		// then the new admin can access secrets
		Expect(err).ToNot(HaveOccurred())
		Expect(kumactl.RunKumactl("get", "secrets")).To(Succeed())
	})

	It("should use dp-token generated offline", func() {
		// given
		token, err := universal.GetKumactlOptions().RunKumactlAndGetOutput("generate", "dataplane-token",
			"--mesh", meshName,
			"--kid", "static-1",
			"--valid-for", "24h",
			"--signing-key-path", filepath.Join("..", "..", "keys", "samplekey.pem"),
		)
		Expect(err).ToNot(HaveOccurred())

		// when
		Expect(universal.Install(DemoClientUniversal("test-server", meshName, WithToken(token)))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			online, _, err := IsDataplaneOnline(universal, meshName, "test-server")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(online).To(BeTrue())
		}, "30s", "1s").Should(Succeed())
	})
}
