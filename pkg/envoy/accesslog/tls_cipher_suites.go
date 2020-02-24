package accesslog

import (
	"fmt"
)

// TlsCipherSuite represents a registered TLS cipher suite.
type TlsCipherSuite uint16

// List of the most commonly used TLS cipher suites.
//
// See https://www.iana.org/assignments/tls-parameters/tls-parameters.xml
const (
	// TLS 1.0 - 1.2 cipher suites.
	TLS_RSA_WITH_RC4_128_SHA                TlsCipherSuite = 0x0005
	TLS_RSA_WITH_3DES_EDE_CBC_SHA           TlsCipherSuite = 0x000a
	TLS_RSA_WITH_AES_128_CBC_SHA            TlsCipherSuite = 0x002f
	TLS_RSA_WITH_AES_256_CBC_SHA            TlsCipherSuite = 0x0035
	TLS_RSA_WITH_AES_128_CBC_SHA256         TlsCipherSuite = 0x003c
	TLS_RSA_WITH_AES_128_GCM_SHA256         TlsCipherSuite = 0x009c
	TLS_RSA_WITH_AES_256_GCM_SHA384         TlsCipherSuite = 0x009d
	TLS_ECDHE_ECDSA_WITH_RC4_128_SHA        TlsCipherSuite = 0xc007
	TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA    TlsCipherSuite = 0xc009
	TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA    TlsCipherSuite = 0xc00a
	TLS_ECDHE_RSA_WITH_RC4_128_SHA          TlsCipherSuite = 0xc011
	TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA     TlsCipherSuite = 0xc012
	TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA      TlsCipherSuite = 0xc013
	TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA      TlsCipherSuite = 0xc014
	TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256 TlsCipherSuite = 0xc023
	TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256   TlsCipherSuite = 0xc027
	TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256   TlsCipherSuite = 0xc02f
	TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 TlsCipherSuite = 0xc02b
	TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384   TlsCipherSuite = 0xc030
	TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 TlsCipherSuite = 0xc02c
	TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305    TlsCipherSuite = 0xcca8
	TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305  TlsCipherSuite = 0xcca9

	// TLS 1.3 cipher suites.
	TLS_AES_128_GCM_SHA256       TlsCipherSuite = 0x1301
	TLS_AES_256_GCM_SHA384       TlsCipherSuite = 0x1302
	TLS_CHACHA20_POLY1305_SHA256 TlsCipherSuite = 0x1303

	// TLS_FALLBACK_SCSV isn't a standard cipher suite but an indicator
	// that the client is doing version fallback. See RFC 7507.
	TLS_FALLBACK_SCSV TlsCipherSuite = 0x5600
)

// String returns an Envoy-compatible name of a TLS cipher suite.
func (s TlsCipherSuite) String() string {
	switch s {
	// TLS 1.0 - 1.2 cipher suites.
	case TLS_RSA_WITH_RC4_128_SHA:
		return "RSA-RC4-128-SHA"
	case TLS_RSA_WITH_3DES_EDE_CBC_SHA:
		return "RSA-3DES-EDE-CBC-SHA"
	case TLS_RSA_WITH_AES_128_CBC_SHA:
		return "RSA-AES-128-CBC-SHA"
	case TLS_RSA_WITH_AES_256_CBC_SHA:
		return "RSA-AES-256-CBC-SHA"
	case TLS_RSA_WITH_AES_128_CBC_SHA256:
		return "RSA-AES-128-CBC-SHA256"
	case TLS_RSA_WITH_AES_128_GCM_SHA256:
		return "RSA-AES-128-GCM-SHA256"
	case TLS_RSA_WITH_AES_256_GCM_SHA384:
		return "RSA-AES-256-GCM-SHA384"
	case TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:
		return "ECDHE-ECDSA-RC4-128-SHA"
	case TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:
		return "ECDHE-ECDSA-AES-128-CBC-SHA"
	case TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:
		return "ECDHE-ECDSA-AES-256-CBC-SHA"
	case TLS_ECDHE_RSA_WITH_RC4_128_SHA:
		return "ECDHE-RSA-RC4-128-SHA"
	case TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:
		return "ECDHE-RSA-3DES-EDE-CBC-SHA"
	case TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:
		return "ECDHE-RSA-AES-128-CBC-SHA"
	case TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:
		return "ECDHE-RSA-AES-256-CBC-SHA"
	case TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256:
		return "ECDHE-ECDSA-AES-128-CBC-SHA256"
	case TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256:
		return "ECDHE-RSA-AES-128-CBC-SHA256"
	case TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return "ECDHE-RSA-AES-128-GCM-SHA256"
	case TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return "ECDHE-ECDSA-AES-128-GCM-SHA256"
	case TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return "ECDHE-RSA-AES-256-GCM-SHA384"
	case TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return "ECDHE-ECDSA-AES-256-GCM-SHA384"
	case TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:
		return "ECDHE-RSA-CHACHA20-POLY1305"
	case TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:
		return "ECDHE-ECDSA-CHACHA20-POLY1305"
	// TLS 1.3 cipher suites.
	case TLS_AES_128_GCM_SHA256:
		return "AES-128-GCM-SHA256"
	case TLS_AES_256_GCM_SHA384:
		return "AES-256-GCM-SHA384"
	case TLS_CHACHA20_POLY1305_SHA256:
		return "CHACHA20-POLY1305-SHA256"
	// RFC 7507
	case TLS_FALLBACK_SCSV:
		return "FALLBACK-SCSV"
	default:
		return fmt.Sprintf("%#x", uint16(s))
	}
}
