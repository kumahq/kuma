package auth

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func OfflineAuth() {
	meshes := []string{
		"offline-auth-1",
		"offline-auth-2",
	}

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
            mesh: offline-auth-1
            key: |
              -----BEGIN RSA PUBLIC KEY-----
              MIIBCgKCAQEAqwbFZ7LSuRGEkFPsZOLYuimsjDeie4sdtqIVW9bLDrTSql+o2sBL
              wt22MJ897/oq7+jZhVlENE1ddAKdFSWv3nhOI/XK9VJt7qNudcoC9252XrycIi5h
              i700CDgdRgRt+2paZiRCgc5afNMHJmVIp2d2lQTUKn/pQGlqY4ufuA3U1z+8t++k
              oGnj0sKIcXzqa5ZZxZ/81khp0e0Ze7llTmfEU3gQXu/Coa2y7LEUHdrNalM3si0v
              FvX0KmBtADEJ4n9Jo4ja3hDmp83Q4KjJq0xKbhh9Fp3AjwjDb0fVFwbt+8SdVgyV
              5PE+7HdigwlJ/cOVb9IY/UKVgCzlW5inCQIDAQAB
              -----END RSA PUBLIC KEY-----
          - kid: static-spki-1
            mesh: offline-auth-1
            key: |
              -----BEGIN PUBLIC KEY-----
              MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqwbFZ7LSuRGEkFPsZOLY
              uimsjDeie4sdtqIVW9bLDrTSql+o2sBLwt22MJ897/oq7+jZhVlENE1ddAKdFSWv
              3nhOI/XK9VJt7qNudcoC9252XrycIi5hi700CDgdRgRt+2paZiRCgc5afNMHJmVI
              p2d2lQTUKn/pQGlqY4ufuA3U1z+8t++koGnj0sKIcXzqa5ZZxZ/81khp0e0Ze7ll
              TmfEU3gQXu/Coa2y7LEUHdrNalM3si0vFvX0KmBtADEJ4n9Jo4ja3hDmp83Q4KjJ
              q0xKbhh9Fp3AjwjDb0fVFwbt+8SdVgyV5PE+7HdigwlJ/cOVb9IY/UKVgCzlW5in
              CQIDAQAB
              -----END PUBLIC KEY-----
          - kid: offline-auth-nomesh-1
            key: |
              -----BEGIN RSA PUBLIC KEY-----
              MIIBCgKCAQEAsGQSfwmBU/DMDLnKCbg7cKUrBEAxDinCPaQ5foF87H8aul4EAzym
              KswoSpwXyyhAqVf2pHJYqkIX0HwL5xkgGy3lvNekgJPLeQaGMg0qVol+tU0/go6i
              50LUzSvPo6kBHCBOiFTNxZ+HRiCdTJd655ALBn1a4LbVPGDqPnHikSWsZg69gkV7
              T+jdPz4rBqfhNahREinVRe1DsLVJ0trjc91+2dRYj1e+tKVQDwCNj5cP2GzYUkAb
              XaMpe1ZGQSC9/gTlJIEU7Lyz7fyOJcCZbGASy8nBixM6E5l8QPrFVIDVkeNJNVQj
              35gOQBJWtsCEiBx3spsKLeoim62wun05HwIDAQAB
              -----END RSA PUBLIC KEY-----
`

	BeforeAll(func() {
		universal = NewUniversalCluster(NewTestingT(), "kuma-offline-auth", Silent)
		Expect(NewClusterSetup().
			Install(Kuma(core.Zone,
				WithYamlConfig(cpCfg),
			)).
			Install(MeshUniversal(meshes[0])).
			Install(MeshUniversal(meshes[1])).
			Setup(universal)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal, meshes[0])
		DebugUniversal(universal, meshes[1])
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
		kumactl := NewKumactlOptionsE2E(universal.GetTesting(), universal.GetKuma().GetName()+"test-admin", false)
		err = kumactl.KumactlConfigControlPlanesAdd(
			"test-admin",
			universal.GetKuma().GetAPIServerAddress(),
			token,
			[]string{},
		)
		Expect(err).ToNot(HaveOccurred())

		// then the new admin can access secrets
		Expect(kumactl.RunKumactl("get", "secrets")).To(Succeed())
	})

	It("should use dp-token generated offline", func() {
		// given
		token, err := universal.GetKumactlOptions().RunKumactlAndGetOutput("generate", "dataplane-token",
			"--mesh", meshes[0],
			"--kid", "static-1",
			"--valid-for", "24h",
			"--signing-key-path", filepath.Join("..", "..", "keys", "samplekey.pem"),
		)
		Expect(err).ToNot(HaveOccurred())

		// when
		Expect(universal.Install(DemoClientUniversal("test-server-1", meshes[0], WithToken(token)))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			online, _, err := IsDataplaneOnline(universal, meshes[0], "test-server-1")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(online).To(BeTrue())
		}, "30s", "1s").Should(Succeed())
	})

	It("should use a dp-token generated offline, validated with a non-mesh scoped key", func() {
		// given
		token, err := universal.GetKumactlOptions().RunKumactlAndGetOutput("generate", "dataplane-token",
			"--mesh", meshes[1],
			"--kid", "offline-auth-nomesh-1",
			"--valid-for", "24h",
			"--signing-key-path", filepath.Join("..", "..", "keys", "samplekey-2.pem"),
		)
		Expect(err).ToNot(HaveOccurred())

		// when
		Expect(universal.Install(DemoClientUniversal("test-server-2", meshes[1], WithToken(token)))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			online, _, err := IsDataplaneOnline(universal, meshes[1], "test-server-2")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(online).To(BeTrue())
		}, "30s", "1s").Should(Succeed())
	})
}
