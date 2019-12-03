package postgres_test

import (
	"github.com/Kong/kuma/pkg/config/plugins/resources/postgres"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSLPostgresStoreConfig", func() {

	type testCase struct {
		config postgres.SSLPostgresStoreConfig
		error  string
	}
	DescribeTable("should validate invalid config",
		func(given testCase) {
			// when
			err := given.config.Validate()

			// then
			Expect(err).To(MatchError(given.error))
		},
		Entry("Require without certPath", testCase{
			config: postgres.SSLPostgresStoreConfig{
				Mode:    postgres.Require,
				KeyPath: "/path",
			},
			error: "CertPath cannot be empty",
		}),
		Entry("Require without keyPath", testCase{
			config: postgres.SSLPostgresStoreConfig{
				Mode:     postgres.Require,
				CertPath: "/path",
			},
			error: "KeyPath cannot be empty",
		}),
		Entry("VerifyCA without RootCertPath", testCase{
			config: postgres.SSLPostgresStoreConfig{
				Mode:     postgres.VerifyCa,
				KeyPath:  "/path",
				CertPath: "/path",
			},
			error: "RootCertPath cannot be empty",
		}),
		Entry("VerifyCA without CertPath", testCase{
			config: postgres.SSLPostgresStoreConfig{
				Mode:         postgres.VerifyCa,
				RootCertPath: "/path",
				KeyPath:      "/path",
			},
			error: "CertPath cannot be empty",
		}),
		Entry("VerifyCA without KeyPath", testCase{
			config: postgres.SSLPostgresStoreConfig{
				Mode:         postgres.VerifyCa,
				RootCertPath: "/path",
				CertPath:     "/path",
			},
			error: "KeyPath cannot be empty",
		}),
		Entry("VerifyFull without RootCertPath", testCase{
			config: postgres.SSLPostgresStoreConfig{
				Mode:     postgres.VerifyFull,
				KeyPath:  "/path",
				CertPath: "/path",
			},
			error: "RootCertPath cannot be empty",
		}),
		Entry("VerifyFull without CertPath", testCase{
			config: postgres.SSLPostgresStoreConfig{
				Mode:         postgres.VerifyFull,
				RootCertPath: "/path",
				KeyPath:      "/path",
			},
			error: "CertPath cannot be empty",
		}),
		Entry("VerifyFull without KeyPath", testCase{
			config: postgres.SSLPostgresStoreConfig{
				Mode:         postgres.VerifyFull,
				RootCertPath: "/path",
				CertPath:     "/path",
			},
			error: "KeyPath cannot be empty",
		}),
	)

	DescribeTable("should validate valid config",
		func(cfg postgres.SSLPostgresStoreConfig) {
			Expect(cfg.Validate()).To(Succeed())
		},
		Entry("mode Disable", postgres.SSLPostgresStoreConfig{
			Mode: postgres.Disable,
		}),
		Entry("mode Require", postgres.SSLPostgresStoreConfig{
			Mode:     postgres.Require,
			KeyPath:  "/path",
			CertPath: "/path",
		}),
		Entry("mode VerifyCA", postgres.SSLPostgresStoreConfig{
			Mode:         postgres.VerifyCa,
			RootCertPath: "/path",
			KeyPath:      "/path",
			CertPath:     "/path",
		}),
		Entry("mode VerifyFull", postgres.SSLPostgresStoreConfig{
			Mode:         postgres.VerifyFull,
			RootCertPath: "/path",
			KeyPath:      "/path",
			CertPath:     "/path",
		}),
	)
})
