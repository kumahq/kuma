package provided

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/Kong/kuma/pkg/core/validators"

	util_tls "github.com/Kong/kuma/pkg/tls"
)

func ValidateCaCert(signingPair util_tls.KeyPair) error {
	err := validateCaCert(signingPair)
	return err.OrNil()
}

func validateCaCert(signingPair util_tls.KeyPair) (verr validators.ValidationError) {
	tlsKeyPair, err := tls.X509KeyPair(signingPair.CertPEM, signingPair.KeyPEM)
	if err != nil {
		verr.AddViolation("cert", fmt.Sprintf("not a valid TLS key pair: %s", err))
		return
	}
	if len(tlsKeyPair.Certificate) != 1 {
		verr.AddViolation("cert", "certificate must be a root CA (certificate chains are not allowed)") // Envoy constraint
		return
	}
	cert, err := x509.ParseCertificate(tlsKeyPair.Certificate[0])
	if err != nil {
		verr.AddViolation("cert", fmt.Sprintf("not a valid x509 certificate: %s", err))
		return
	}
	if cert.Issuer.String() != cert.Subject.String() {
		verr.AddViolation("cert", "certificate must be self-signed (intermediate CAs are not allowed)") // Envoy constraint
	}
	if !cert.IsCA {
		verr.AddViolation("cert", "basic constraint 'CA' must be set to 'true' (see X509-SVID: 4.1. Basic Constraints)")
	}
	if cert.KeyUsage&x509.KeyUsageCertSign == 0 {
		verr.AddViolation("cert", "key usage extension 'keyCertSign' must be set (see X509-SVID: 4.3. Key Usage)")
	}
	if cert.KeyUsage&x509.KeyUsageDigitalSignature != 0 {
		verr.AddViolation("cert", "key usage extension 'digitalSignature' must NOT be set (see X509-SVID: Appendix A. X.509 Field Reference)")
	}
	if cert.KeyUsage&x509.KeyUsageKeyAgreement != 0 {
		verr.AddViolation("cert", "key usage extension 'keyAgreement' must NOT be set (see X509-SVID: Appendix A. X.509 Field Reference)")
	}
	if cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
		verr.AddViolation("cert", "key usage extension 'keyEncipherment' must NOT be set (see X509-SVID: Appendix A. X.509 Field Reference)")
	}

	return
}
