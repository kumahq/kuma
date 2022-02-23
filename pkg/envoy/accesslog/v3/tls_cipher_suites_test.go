package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
)

var _ = Describe("TlsCipherSuite", func() {

	Describe("String()", func() {
		type testCase struct {
			cipherID TlsCipherSuite
			expected string
		}

		DescribeTable("should return an Envoy-compatible TLS Cipher Suite name",
			func(given testCase) {
				Expect(given.cipherID.String()).To(Equal(given.expected))
			},
			Entry("TLS_RSA_WITH_RC4_128_SHA", testCase{
				cipherID: TLS_RSA_WITH_RC4_128_SHA,
				expected: `RSA-RC4-128-SHA`,
			}),
			Entry("TLS_RSA_WITH_3DES_EDE_CBC_SHA", testCase{
				cipherID: TLS_RSA_WITH_3DES_EDE_CBC_SHA,
				expected: `RSA-3DES-EDE-CBC-SHA`,
			}),
			Entry("TLS_RSA_WITH_AES_128_CBC_SHA", testCase{
				cipherID: TLS_RSA_WITH_AES_128_CBC_SHA,
				expected: `RSA-AES-128-CBC-SHA`,
			}),
			Entry("TLS_RSA_WITH_AES_256_CBC_SHA", testCase{
				cipherID: TLS_RSA_WITH_AES_256_CBC_SHA,
				expected: `RSA-AES-256-CBC-SHA`,
			}),
			Entry("TLS_RSA_WITH_AES_128_CBC_SHA256", testCase{
				cipherID: TLS_RSA_WITH_AES_128_CBC_SHA256,
				expected: `RSA-AES-128-CBC-SHA256`,
			}),
			Entry("TLS_RSA_WITH_AES_128_GCM_SHA256", testCase{
				cipherID: TLS_RSA_WITH_AES_128_GCM_SHA256,
				expected: `RSA-AES-128-GCM-SHA256`,
			}),
			Entry("TLS_RSA_WITH_AES_256_GCM_SHA384", testCase{
				cipherID: TLS_RSA_WITH_AES_256_GCM_SHA384,
				expected: `RSA-AES-256-GCM-SHA384`,
			}),
			Entry("TLS_ECDHE_ECDSA_WITH_RC4_128_SHA", testCase{
				cipherID: TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
				expected: `ECDHE-ECDSA-RC4-128-SHA`,
			}),
			Entry("TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA", testCase{
				cipherID: TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				expected: `ECDHE-ECDSA-AES-128-CBC-SHA`,
			}),
			Entry("TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA", testCase{
				cipherID: TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				expected: `ECDHE-ECDSA-AES-256-CBC-SHA`,
			}),
			Entry("TLS_ECDHE_RSA_WITH_RC4_128_SHA", testCase{
				cipherID: TLS_ECDHE_RSA_WITH_RC4_128_SHA,
				expected: `ECDHE-RSA-RC4-128-SHA`,
			}),
			Entry("TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA", testCase{
				cipherID: TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
				expected: `ECDHE-RSA-3DES-EDE-CBC-SHA`,
			}),
			Entry("TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA", testCase{
				cipherID: TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				expected: `ECDHE-RSA-AES-128-CBC-SHA`,
			}),
			Entry("TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA", testCase{
				cipherID: TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				expected: `ECDHE-RSA-AES-256-CBC-SHA`,
			}),
			Entry("TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256", testCase{
				cipherID: TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
				expected: `ECDHE-ECDSA-AES-128-CBC-SHA256`,
			}),
			Entry("TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256", testCase{
				cipherID: TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
				expected: `ECDHE-RSA-AES-128-CBC-SHA256`,
			}),
			Entry("TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", testCase{
				cipherID: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				expected: `ECDHE-RSA-AES-128-GCM-SHA256`,
			}),
			Entry("TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", testCase{
				cipherID: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				expected: `ECDHE-RSA-AES-128-GCM-SHA256`,
			}),
			Entry("TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256", testCase{
				cipherID: TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				expected: `ECDHE-ECDSA-AES-128-GCM-SHA256`,
			}),
			Entry("TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384", testCase{
				cipherID: TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				expected: `ECDHE-RSA-AES-256-GCM-SHA384`,
			}),
			Entry("TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384", testCase{
				cipherID: TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				expected: `ECDHE-ECDSA-AES-256-GCM-SHA384`,
			}),
			Entry("TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305", testCase{
				cipherID: TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				expected: `ECDHE-RSA-CHACHA20-POLY1305`,
			}),
			Entry("TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305", testCase{
				cipherID: TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				expected: `ECDHE-ECDSA-CHACHA20-POLY1305`,
			}),
			Entry("TLS_AES_128_GCM_SHA256", testCase{
				cipherID: TLS_AES_128_GCM_SHA256,
				expected: `AES-128-GCM-SHA256`,
			}),
			Entry("TLS_AES_256_GCM_SHA384", testCase{
				cipherID: TLS_AES_256_GCM_SHA384,
				expected: `AES-256-GCM-SHA384`,
			}),
			Entry("TLS_CHACHA20_POLY1305_SHA256", testCase{
				cipherID: TLS_CHACHA20_POLY1305_SHA256,
				expected: `CHACHA20-POLY1305-SHA256`,
			}),
			Entry("TLS_FALLBACK_SCSV", testCase{
				cipherID: TLS_FALLBACK_SCSV,
				expected: `FALLBACK-SCSV`,
			}),
			Entry("TLS_RSA_WITH_NULL_MD5", testCase{
				cipherID: TlsCipherSuite(0x01),
				expected: `0x1`,
			}),
			Entry("TLS_RSA_WITH_NULL_SHA", testCase{
				cipherID: TlsCipherSuite(0x02),
				expected: `0x2`,
			}),
			Entry("TLS_RSA_EXPORT_WITH_RC4_40_MD5", testCase{
				cipherID: TlsCipherSuite(0x03),
				expected: `0x3`,
			}),
			Entry("TLS_KRB5_WITH_3DES_EDE_CBC_SHA", testCase{
				cipherID: TlsCipherSuite(0x1F),
				expected: `0x1f`,
			}),
		)
	})
})
