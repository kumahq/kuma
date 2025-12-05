package v1alpha1_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
)

var _ = Describe("MeshTrust", func() {
	DescribeTable("should validate all folders", func(inputFile string) {
		// setup
		meshTrust := v1alpha1.NewMeshTrustResource()

		// when
		contents, err := os.ReadFile(inputFile)
		Expect(err).ToNot(HaveOccurred())
		err = core_model.FromYAML(contents, &meshTrust.Spec)
		Expect(err).ToNot(HaveOccurred())

		meshTrust.SetMeta(&test_model.ResourceMeta{
			Name: "test",
			Mesh: core_model.DefaultMesh,
		})

		// and
		verr := meshTrust.Validate()
		actual, err := yaml.Marshal(verr)
		if string(actual) == "null\n" {
			actual = []byte{}
		}
		Expect(err).ToNot(HaveOccurred())

		// then
		goldenFile := strings.ReplaceAll(inputFile, ".input.yaml", ".golden.yaml")
		Expect(actual).To(matchers.MatchGoldenYAML(goldenFile))
	}, test.EntriesForFolder("spec"))

	Describe("Origin Field", func() {
		It("should allow status.origin to be set", func() {
			// given
			meshTrust := v1alpha1.NewMeshTrustResource()
			meshTrust.SetMeta(&test_model.ResourceMeta{
				Name: "test",
				Mesh: core_model.DefaultMesh,
			})
			meshTrust.Spec = &v1alpha1.MeshTrust{
				TrustDomain: "test.local",
				CABundles: []v1alpha1.CABundle{
					{
						Type: v1alpha1.PemCABundleType,
						PEM: &v1alpha1.PEM{
							Value: "-----BEGIN CERTIFICATE-----\nMIIBkTCB+wIJAKHHCgVZU7POMA0GCSqGSIb3DQEBCwUAMBExDzANBgNVBAMMBm15\nLWNhMB4XDTIzMTEwNzA3MzAwMFoXDTI0MTEwNjA3MzAwMFowETEPMA0GA1UEAwwG\nbXktY2EwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBANHkxZmVR1n0P7nPmqTf\nqFqBqKf3gLqHWD0VqLCMqQGNqE3xZqnWQvVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqH\nXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKG\nqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqAgMBAAEwDQYJKoZIhvcNAQELBQADgYEA\nYQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKG\nqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZq\nKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVO\nZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhq\nVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQs=\n-----END CERTIFICATE-----",
						},
					},
				},
			}
			kri := "kri_mid_default__default_my-identity_"
			meshTrust.Status = &v1alpha1.MeshTrustStatus{
				Origin: &v1alpha1.Origin{KRI: &kri},
			}

			// when
			err := meshTrust.Validate()

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should warn about deprecated spec.origin", func() {
			// given
			meshTrust := v1alpha1.NewMeshTrustResource()
			meshTrust.SetMeta(&test_model.ResourceMeta{
				Name: "test",
				Mesh: core_model.DefaultMesh,
			})
			kri := "kri_mid_default__default_my-identity_"
			meshTrust.Spec = &v1alpha1.MeshTrust{
				TrustDomain: "test.local",
				CABundles: []v1alpha1.CABundle{
					{
						Type: v1alpha1.PemCABundleType,
						PEM: &v1alpha1.PEM{
							Value: "-----BEGIN CERTIFICATE-----\nMIIBkTCB+wIJAKHHCgVZU7POMA0GCSqGSIb3DQEBCwUAMBExDzANBgNVBAMMBm15\nLWNhMB4XDTIzMTEwNzA3MzAwMFoXDTI0MTEwNjA3MzAwMFowETEPMA0GA1UEAwwG\nbXktY2EwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBANHkxZmVR1n0P7nPmqTf\nqFqBqKf3gLqHWD0VqLCMqQGNqE3xZqnWQvVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqH\nXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKG\nqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqAgMBAAEwDQYJKoZIhvcNAQELBQADgYEA\nYQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKG\nqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZq\nKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVO\nZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhq\nVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQs=\n-----END CERTIFICATE-----",
						},
					},
				},
				Origin: &v1alpha1.Origin{KRI: &kri},
			}

			// when
			err := meshTrust.Validate()

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("spec.origin is deprecated"))
			Expect(err.Error()).To(ContainSubstring("use status.origin instead"))
		})

		It("should validate resources without origin", func() {
			// given
			meshTrust := v1alpha1.NewMeshTrustResource()
			meshTrust.SetMeta(&test_model.ResourceMeta{
				Name: "test",
				Mesh: core_model.DefaultMesh,
			})
			meshTrust.Spec = &v1alpha1.MeshTrust{
				TrustDomain: "test.local",
				CABundles: []v1alpha1.CABundle{
					{
						Type: v1alpha1.PemCABundleType,
						PEM: &v1alpha1.PEM{
							Value: "-----BEGIN CERTIFICATE-----\nMIIBkTCB+wIJAKHHCgVZU7POMA0GCSqGSIb3DQEBCwUAMBExDzANBgNVBAMMBm15\nLWNhMB4XDTIzMTEwNzA3MzAwMFoXDTI0MTEwNjA3MzAwMFowETEPMA0GA1UEAwwG\nbXktY2EwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBANHkxZmVR1n0P7nPmqTf\nqFqBqKf3gLqHWD0VqLCMqQGNqE3xZqnWQvVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqH\nXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKG\nqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqAgMBAAEwDQYJKoZIhvcNAQELBQADgYEA\nYQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKG\nqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZq\nKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVO\nZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhq\nVOZqKGqHXQsVZZQv3Lq3kqTcTZ+5KqHhqVOZqKGqHXQs=\n-----END CERTIFICATE-----",
						},
					},
				},
			}

			// when
			err := meshTrust.Validate()

			// then
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
