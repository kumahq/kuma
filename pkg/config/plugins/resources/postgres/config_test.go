package postgres_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
)

var _ = Describe("TLSPostgresStoreConfig", func() {

	type testCase struct {
		config postgres.TLSPostgresStoreConfig
		error  string
	}
	DescribeTable("should validate invalid config",
		func(given testCase) {
			// when
			err := given.config.Validate()

			// then
			Expect(err).To(MatchError(given.error))
		},
		Entry("VerifyCA without CAPath", testCase{
			config: postgres.TLSPostgresStoreConfig{
				Mode:     postgres.VerifyCa,
				KeyPath:  "/path",
				CertPath: "/path",
			},
			error: "CAPath cannot be empty",
		}),
		Entry("VerifyFull without CAPath", testCase{
			config: postgres.TLSPostgresStoreConfig{
				Mode:     postgres.VerifyFull,
				KeyPath:  "/path",
				CertPath: "/path",
			},
			error: "CAPath cannot be empty",
		}),
		Entry("CertPath without KeyPath", testCase{
			config: postgres.TLSPostgresStoreConfig{
				Mode:     postgres.VerifyNone,
				KeyPath:  "",
				CertPath: "/path",
			},
			error: "KeyPath cannot be empty when CertPath is provided",
		}),
		Entry("KeyPath without CertPath", testCase{
			config: postgres.TLSPostgresStoreConfig{
				Mode:     postgres.VerifyNone,
				KeyPath:  "/path",
				CertPath: "",
			},
			error: "CertPath cannot be empty when KeyPath is provided",
		}),
	)

	DescribeTable("should validate valid config",
		func(cfg postgres.TLSPostgresStoreConfig) {
			Expect(cfg.Validate()).To(Succeed())
		},
		Entry("mode Disable", postgres.TLSPostgresStoreConfig{
			Mode: postgres.Disable,
		}),
		Entry("mode VerifyNone", postgres.TLSPostgresStoreConfig{
			Mode:     postgres.VerifyNone,
			KeyPath:  "/path",
			CertPath: "/path",
		}),
		Entry("mode VerifyCA", postgres.TLSPostgresStoreConfig{
			Mode:     postgres.VerifyCa,
			CAPath:   "/path",
			KeyPath:  "/path",
			CertPath: "/path",
		}),
		Entry("mode VerifyFull", postgres.TLSPostgresStoreConfig{
			Mode:     postgres.VerifyFull,
			CAPath:   "/path",
			KeyPath:  "/path",
			CertPath: "/path",
		}),
	)
})

var _ = Describe("PostgresStoreConfig", func() {
	type testCase struct {
		config postgres.PostgresStoreConfig
		error  string
	}
	DescribeTable("should validate invalid config",
		func(given testCase) {
			// when
			err := given.config.Validate()

			// then
			Expect(err).To(MatchError(given.error))
		},
		Entry("MinReconnectInterval is equal to MaxReconnectInterval", testCase{
			config: postgres.PostgresStoreConfig{
				Host:     "localhost",
				User:     "postgres",
				Password: "postgres",
				DbName:   "kuma",
				TLS: postgres.TLSPostgresStoreConfig{
					Mode:     postgres.VerifyFull,
					CAPath:   "/path",
					KeyPath:  "/path",
					CertPath: "/path",
				},
				MinReconnectInterval: 10 * time.Second,
				MaxReconnectInterval: 10 * time.Second,
			},
			error: "MinReconnectInterval should be less than MaxReconnectInterval",
		}),
		Entry("MinReconnectInterval is greater than MaxReconnectInterval", testCase{
			config: postgres.PostgresStoreConfig{
				Host:     "localhost",
				User:     "postgres",
				Password: "postgres",
				DbName:   "kuma",
				TLS: postgres.TLSPostgresStoreConfig{
					Mode:     postgres.VerifyFull,
					CAPath:   "/path",
					KeyPath:  "/path",
					CertPath: "/path",
				},
				MinReconnectInterval: 10 * time.Second,
				MaxReconnectInterval: 1 * time.Second,
			},
			error: "MinReconnectInterval should be less than MaxReconnectInterval",
		}),
	)
})
