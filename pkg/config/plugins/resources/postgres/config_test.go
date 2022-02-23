package postgres_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
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
	type stringTestCase struct {
		given    postgres.PostgresStoreConfig
		expected string
	}
	DescribeTable("converts to Postgres connection string",
		func(testCase stringTestCase) {
			// when
			str, err := testCase.given.ConnectionString()
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(str).To(Equal(testCase.expected))
		},
		Entry("basic config", stringTestCase{
			given: postgres.PostgresStoreConfig{
				Host:     "localhost",
				User:     "postgres",
				Password: `postgres`,
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
			expected: `host='localhost' port=0 user='postgres' password='postgres' dbname='kuma' connect_timeout=0 sslmode=verify-full sslcert='/path' sslkey='/path' sslrootcert='/path'`,
		}),
		Entry("password needing escape", stringTestCase{
			given: postgres.PostgresStoreConfig{
				Host:     "localhost",
				User:     "postgres",
				Password: `'\`,
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
			expected: `host='localhost' port=0 user='postgres' password='\'\\' dbname='kuma' connect_timeout=0 sslmode=verify-full sslcert='/path' sslkey='/path' sslrootcert='/path'`,
		}),
	)
	type validateTestCase struct {
		config postgres.PostgresStoreConfig
		error  string
	}
	DescribeTable("should validate invalid config",
		func(given validateTestCase) {
			// when
			err := given.config.Validate()

			// then
			Expect(err).To(MatchError(given.error))
		},
		Entry("MinReconnectInterval is equal to MaxReconnectInterval", validateTestCase{
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
		Entry("MinReconnectInterval is greater than MaxReconnectInterval", validateTestCase{
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
