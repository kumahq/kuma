package vault_test

import (
	"context"

	vault_api "github.com/hashicorp/vault/api"

	sds_auth "github.com/Kong/kuma/pkg/sds/auth"
	"github.com/Kong/kuma/pkg/sds/provider"
	"github.com/Kong/kuma/pkg/sds/provider/ca"
	"github.com/Kong/kuma/pkg/sds/provider/identity"
	sds_vault "github.com/Kong/kuma/pkg/sds/provider/vault"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Vault Providers", func() {

	const token = "sampleTestToken"
	vaultContainer := VaultContainer{
		Token: token,
	}

	var rootCaCert string

	newClient := func(token string) *vault_api.Client {
		addr, err := vaultContainer.Address()
		Expect(err).ToNot(HaveOccurred())
		cfg := sds_vault.Config{
			Address: addr,
			Token:   token,
			Tls: sds_vault.TLSConfig{
				SkipVerify: true,
			},
		}
		client, err := sds_vault.NewVaultClient(cfg)
		Expect(err).ToNot(HaveOccurred())
		return client
	}

	BeforeSuite(func() {
		Expect(vaultContainer.Start()).To(Succeed())

		client := newClient(token)

		// setup PKI
		payload := map[string]interface{}{
			"type": "pki",
		}
		_, err := client.Logical().Write("/sys/mounts/kuma-pki-default", payload)
		Expect(err).ToNot(HaveOccurred())

		// generate root CA
		payload = map[string]interface{}{
			"type":         "internal",
			"organization": "Kuma",
			"ou":           "Mesh",
			"common_name":  "default",
			"uri_sans":     "spiffe://default",
		}

		resp, err := client.Logical().Write("kuma-pki-default/root/generate/internal", payload)
		Expect(err).ToNot(HaveOccurred())
		rootCaCert = resp.Data["certificate"].(string)

		// create role for web DP
		payload = map[string]interface{}{
			"allowed_uri_sans":                   "spiffe://default/web",
			"key_usage":                          []string{"KeyUsageKeyEncipherment", "KeyUsageKeyAgreement", "KeyUsageDigitalSignature"},
			"ext_key_usage":                      []string{"ExtKeyUsageServerAuth", "ExtKeyUsageClientAuth,"},
			"client_flag":                        true,
			"require_cn":                         false,
			"basic_constraints_valid_for_non_ca": true,
		}
		_, err = client.Logical().Write("kuma-pki-default/roles/dp-web", payload)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterSuite(func() {
		Expect(vaultContainer.Stop()).To(Succeed())
	})

	Describe("MeshCaProvider", func() {
		var secretProvider provider.SecretProvider
		BeforeEach(func() {
			secretProvider = sds_vault.NewMeshCaProvider(newClient(token))
		})

		It("should return cert for default mesh", func() {
			// when
			secret, err := secretProvider.Get(context.Background(), "", sds_auth.Identity{Mesh: "default"})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(secret.(*ca.MeshCaSecret).PemCerts).To(HaveLen(1))
			Expect(string(secret.(*ca.MeshCaSecret).PemCerts[0])).To(Equal(rootCaCert))
		})

		It("should return an error when PKI for mesh not exist", func() {
			// when
			_, err := secretProvider.Get(context.Background(), "", sds_auth.Identity{Mesh: "non-existent"})

			// then
			Expect(err).To(MatchError("there is no PKI enabled for non-existent mesh"))
		})

		It("should return an error when invalid token is used", func() {
			// given
			secretProvider := sds_vault.NewMeshCaProvider(newClient("invalid-token"))

			// when
			_, err := secretProvider.Get(context.Background(), "", sds_auth.Identity{Mesh: "non-existent"})

			// then
			Expect(err).To(MatchError("permission denied - use token that allows to read CA cert of /v1/kuma-pki-non-existent/ca/pem"))
		})
	})

	Describe("IdentityCertProvider", func() {

		var secretProvider provider.SecretProvider

		BeforeEach(func() {
			secretProvider = sds_vault.NewIdentityCertProvider(newClient(token))
		})

		It("should return cer key pair for service web in default mesh", func() {
			// when
			secret, err := secretProvider.Get(context.Background(), "", sds_auth.Identity{Mesh: "default", Service: "web"})

			// then
			Expect(err).ToNot(HaveOccurred())

			identityCertSecret := secret.(*identity.IdentityCertSecret)
			Expect(identityCertSecret.PemCerts).To(HaveLen(1))
			Expect(identityCertSecret.PemKey).ToNot(BeEmpty())
		})

		It("should return an error when there is no role for given dataplane", func() {
			// when
			_, err := secretProvider.Get(context.Background(), "", sds_auth.Identity{Mesh: "default", Service: "unknown"})

			// then
			Expect(err).To(MatchError("there is no dp-unknown role for PKI kuma-pki-default"))
		})

		It("should return an error when invalid token is used", func() {
			// given
			secretProvider := sds_vault.NewIdentityCertProvider(newClient("invalid-token"))

			// when
			_, err := secretProvider.Get(context.Background(), "", sds_auth.Identity{Mesh: "default", Service: "web"})

			// then
			Expect(err).To(MatchError("permission denied - use token that allows to generate cert of kuma-pki-default/issue/dp-web"))
		})
	})
})
