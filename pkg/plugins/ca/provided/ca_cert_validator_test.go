package provided_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"math/big"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	ca_issuer "github.com/Kong/kuma/pkg/core/ca/issuer"
	. "github.com/Kong/kuma/pkg/plugins/ca/provided"

	util_tls "github.com/Kong/kuma/pkg/tls"
)

var _ = Describe("ValidateCaCert()", func() {

	It("should accept proper CA certificates", func() {
		// when
		signingPair, err := ca_issuer.NewRootCA("demo")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = ValidateCaCert(*signingPair)
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	NewSelfSignedCert := func(newTemplate func() *x509.Certificate) (*util_tls.KeyPair, error) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate a private key")
		}
		template := newTemplate()
		template.PublicKey = key.Public()
		cert, err := x509.CreateCertificate(rand.Reader, template, template, key.Public(), key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to sign X509 certificate")
		}
		return util_tls.ToKeyPair(key, cert)
	}

	type testCase struct {
		input       util_tls.KeyPair
		expectedErr string
	}

	DescribeTable("should reject invalid input",
		func(givenFunc func() testCase) {
			given := givenFunc()

			// when
			err := ValidateCaCert(given.input)
			// then
			Expect(err).ToNot(BeNil())

			// when
			actual, err := yaml.Marshal(err)
			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(actual).To(MatchYAML(given.expectedErr))
		},
		Entry("empty key pair", func() testCase {
			return testCase{
				expectedErr: `
                violations:
                - field: cert
                  message: 'not a valid TLS key pair: tls: failed to find any PEM data in certificate input'
`,
			}
		}),
		Entry("bad plain text values", func() testCase {
			return testCase{
				expectedErr: `
                violations:
                - field: cert
                  message: 'not a valid TLS key pair: tls: failed to find any PEM data in certificate input'
`,
				input: util_tls.KeyPair{
					CertPEM: []byte("CERT"),
					KeyPEM:  []byte("KEY"),
				},
			}
		}),
		Entry("cert and key don't match", func() testCase {
			return testCase{
				expectedErr: `
                violations:
                - field: cert
                  message: 'not a valid TLS key pair: tls: private key does not match public key'
`,
				input: util_tls.KeyPair{
					CertPEM: []byte(`
-----BEGIN CERTIFICATE-----
MIIDGzCCAgOgAwIBAgIBADANBgkqhkiG9w0BAQsFADAwMQ0wCwYDVQQKEwRLdW1h
MQ0wCwYDVQQLEwRNZXNoMRAwDgYDVQQDEwdkZWZhdWx0MB4XDTIwMDEyOTE2MDgw
NFoXDTMwMDEyNjE2MDgxNFowMDENMAsGA1UEChMES3VtYTENMAsGA1UECxMETWVz
aDEQMA4GA1UEAxMHZGVmYXVsdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBANhewNnHZI0f+55vsm+iGej9NAYFtCb2FNzFHZlGcu0F07YSAyorPJuM+6V3
BFcY2IkWHL8WOooNmJ0X/yzBd4RSrb3TacGKtaDayRjTo8JOW7Nlh+WvwR18KHjC
QXDjlqkmfExdYIUZjOqJOhu9nO59fqz0SJFvo2WKkkP7CaTLQXt1p3+Hm1Xo5WCX
ZfD7W57YhNBZZLljip/N8pDL7b2Vkhe+txbv/PqVrDRMGoyRBnPNAfS7SPocRbcE
S9th2CesNu+Iwltu4gJBOQbpydBIjJvr1zrx/zxbxM+EbqbGr6gwquGvKTyXHq20
u5CE4tWy3GKSh5LEVItPS066d5UCAwEAAaNAMD4wDgYDVR0PAQH/BAQDAgEGMA8G
A1UdEwEB/wQFMAMBAf8wGwYDVR0RBBQwEoYQc3BpZmZlOi8vZGVmYXVsdDANBgkq
hkiG9w0BAQsFAAOCAQEAMvMqCzbjEveuMlTch9q+/6KcFVUkwQcTcxxs0MPnw5Lw
hY6xo7FvIHNLJDRlShoAyjI6OJZobJ7PFaFnIXWlNcN1F+gA9OZSSWYgJNJl4zee
eS2pHgbxZ6OJqCbDGYWekF3d1vEcI3gaRfxFDVa8eJFBq+B0v2Pho8A2bY5srO0S
LG//nB619RAlzta6BxweiCmFxPyB+oqJl3X9NfWC7988CzfZAqHA+CqO6tJS5i53
WMciH6+W7o9wdsBrfFVx2tGZqc4bZgoZptHFieqz7YBnT0Ozg+NwBU6apAtAc5Ym
DMoTRP2Vo+BEm4uS4GcIFZYqrOsPuuyMuBd0NDE33g==
-----END CERTIFICATE-----
`),
					KeyPEM: []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAvh4DhQDuE/oXjEJzt1TmgmM6s5DtsIYIEPssxdDiDEjmJzhG
APd0BmAfrz3/fXZ6CqqqfAvTKkTDGjSN0hGtt1thF72weNfmQX9c1gUukEA0gERS
OsSWlzp8Azc5fjXtl5jiQkvaapNoQ8PljLc8SEsgndvdxD1tDVWN6h90bMgtuqyK
M7Fn/1+nwhRExKRTOSlR0iS9eddWPcN3HNGrl3vjCF0oUF/79+nJJqeaXawt0+2P
CMkxY6T6yzFaS/h0Oy3mXLOQ1H5n3cZ2MSEMKQSs9ff8gzeHpWyXpgB1XEKNcUvi
rVKfaJ/WE5v0rieG89B/YnQTduQnrwmPVuwm4wIDAQABAoIBAQCxKKrC7+DqwKvc
ybem6Ph8HBeBaNX1HpC5sjVAiKt8IxpFBc1F7VEy97POywkfUp3a/rorKaG2y6i6
7KoTTOIB8KcDRoIBub4Y3qQV03JWfV3vALtXhAWIGrmhDX8Hux0RnSeJ+8EmewI3
034+qCkGfOuB7nYy/cJ3IHhD6NfG3Q3FrBrGfsI2TGEeGmPJ2Xi8ZyfbluRb/1Bt
NesS6pDbRpZ5/IoauLUtITY3bazpzghm2tJNdrJIP7ohaoMF0WYciPyD5xpNlykt
V8Q2jzNmPYVXuUpi4oPekq4Mg1vd/LPS/JE558Am1LEiXrycelGNrDvJW7hTDLVx
DLRFuMMxAoGBAMkjupL3mxAfNytXM++WxJbdWPuw/vvAeN60ifFu6RUrMs/aXocn
4xSunNF58O2aRfSq/B9LJ+pXtmdITV+Bu0Y1XefKtNUNoqIapAbA8gAWUcFSkDRd
999rh0vWPbx4d3k69iS6xIjVaRcxeuaBbKRWqUcrxDuAydhwTLIRMD1vAoGBAPH4
quLGkr1MdTeZ3qPAWc9mGelp0LhHukjnLB+nMdI73OH7IlX5or11yr6La/+sTmmQ
fI+oITLuCyey7VnWBDhrPmWFGA1BmZIVDqjkJJNwyWQO7N27rQEQoNKm5n6Q+boy
StNKa/ljduYXCjsBndOmF1wSrAwL+u9rQ3x4k9vNAoGAGY5vm1LYofDFar1WvP90
FRMkxj4T99rZwLpBuKp19RmbCCvfzN51jOAuzrLmuNncP50mEbfT54OjinX2Vsc+
C0qmltf7qAJmgqBN7QnA9d/gHWcnKXAzGXEpLKqZB4Rq8b1bHwmYBSbQhoDj87vI
GQ1lzsQx17mia9zA8fMbJQMCgYB0D+2PpuW9rM3QpJp4+wtZAsVNAzddHPKKg2/T
ovOvvoz9S+M1T+8yZyyfZuqfkTtvQSGuGlwKPMnW+ekFHTWbBj3Ani1iNmP+AOGu
OvgcTI4c01fkJ2AdUaeCQxHuBYXzPKpNXLYbwgzG4qhCk0zrtxAfVsl1Yc20R0Pw
kTmCxQKBgQCzd/OOLm7vDEUqYNUNgKlf8I9xM84IwUy+pJP8RaeFDtFj7tVDpY2P
GXHBXcIBDRPnmBxC7cGHCB3KBWJp11smw2qA0ZgmBIShNm1RDHf/1h0yOxSz2+fB
bgeEDefxTxoTMgJ1urwl0KX6R9dbv9YWZWJXk2DQj6UwkMEyXpc+kw==
-----END RSA PRIVATE KEY-----
`),
				},
			}
		}),
		Entry("chain of CAs", func() testCase {
			return testCase{
				expectedErr: `
                violations:
                - field: cert
                  message: "certificate must be a root CA (certificate chains are not allowed)"
`,
				input: util_tls.KeyPair{
					CertPEM: []byte(`
-----BEGIN CERTIFICATE-----
MIIDKzCCAhOgAwIBAgIBADANBgkqhkiG9w0BAQsFADAwMQ0wCwYDVQQKEwRLdW1h
MQ0wCwYDVQQLEwRNZXNoMRAwDgYDVQQDEwdkZWZhdWx0MB4XDTIwMDEyOTE2MDgw
NFoXDTMwMDEyNjE2MDgxNFowQDENMAsGA1UEChMES3VtYTEdMAsGA1UECxMETWVz
aDAOBgNVBAsTB2xldmVsLTExEDAOBgNVBAMTB2RlZmF1bHQwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQC+HgOFAO4T+heMQnO3VOaCYzqzkO2whggQ+yzF
0OIMSOYnOEYA93QGYB+vPf99dnoKqqp8C9MqRMMaNI3SEa23W2EXvbB41+ZBf1zW
BS6QQDSARFI6xJaXOnwDNzl+Ne2XmOJCS9pqk2hDw+WMtzxISyCd293EPW0NVY3q
H3RsyC26rIozsWf/X6fCFETEpFM5KVHSJL1511Y9w3cc0auXe+MIXShQX/v36ckm
p5pdrC3T7Y8IyTFjpPrLMVpL+HQ7LeZcs5DUfmfdxnYxIQwpBKz19/yDN4elbJem
AHVcQo1xS+KtUp9on9YTm/SuJ4bz0H9idBN25CevCY9W7CbjAgMBAAGjQDA+MA4G
A1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MBsGA1UdEQQUMBKGEHNwaWZm
ZTovL2RlZmF1bHQwDQYJKoZIhvcNAQELBQADggEBACVXnYWCCrji551pbJsOCGYJ
GEqlvcwNnnYdykas4GrfsbW2rglmaXv0uG8iH2sAH+4/MjGjnlQ6Y6Fj7mDFnidj
ugU964sEDnLuU0CtaIpHl7VZ13I0EzmfY+GsCrcIXIxbAxwWTJhz77XqbHe3baLx
Sh9wHgz/aZuy99rq9OoAvUALEaIfxrvUsVs25jLuv0Xzy57B2Dpqo0odshDA4WSS
MynQnSX7aFg1jqZQL4YjPHryEQQRj8mgjqiWp8M4/PHq5s09zDMB0DCag0QtdC/k
ydtqRoojiRS2fXY8DhFRqqRVBqLvA+7eTEKpzfjUTyEovMqxIM2n4U5MSGKQlbM=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDGzCCAgOgAwIBAgIBADANBgkqhkiG9w0BAQsFADAwMQ0wCwYDVQQKEwRLdW1h
MQ0wCwYDVQQLEwRNZXNoMRAwDgYDVQQDEwdkZWZhdWx0MB4XDTIwMDEyOTE2MDgw
NFoXDTMwMDEyNjE2MDgxNFowMDENMAsGA1UEChMES3VtYTENMAsGA1UECxMETWVz
aDEQMA4GA1UEAxMHZGVmYXVsdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBANhewNnHZI0f+55vsm+iGej9NAYFtCb2FNzFHZlGcu0F07YSAyorPJuM+6V3
BFcY2IkWHL8WOooNmJ0X/yzBd4RSrb3TacGKtaDayRjTo8JOW7Nlh+WvwR18KHjC
QXDjlqkmfExdYIUZjOqJOhu9nO59fqz0SJFvo2WKkkP7CaTLQXt1p3+Hm1Xo5WCX
ZfD7W57YhNBZZLljip/N8pDL7b2Vkhe+txbv/PqVrDRMGoyRBnPNAfS7SPocRbcE
S9th2CesNu+Iwltu4gJBOQbpydBIjJvr1zrx/zxbxM+EbqbGr6gwquGvKTyXHq20
u5CE4tWy3GKSh5LEVItPS066d5UCAwEAAaNAMD4wDgYDVR0PAQH/BAQDAgEGMA8G
A1UdEwEB/wQFMAMBAf8wGwYDVR0RBBQwEoYQc3BpZmZlOi8vZGVmYXVsdDANBgkq
hkiG9w0BAQsFAAOCAQEAMvMqCzbjEveuMlTch9q+/6KcFVUkwQcTcxxs0MPnw5Lw
hY6xo7FvIHNLJDRlShoAyjI6OJZobJ7PFaFnIXWlNcN1F+gA9OZSSWYgJNJl4zee
eS2pHgbxZ6OJqCbDGYWekF3d1vEcI3gaRfxFDVa8eJFBq+B0v2Pho8A2bY5srO0S
LG//nB619RAlzta6BxweiCmFxPyB+oqJl3X9NfWC7988CzfZAqHA+CqO6tJS5i53
WMciH6+W7o9wdsBrfFVx2tGZqc4bZgoZptHFieqz7YBnT0Ozg+NwBU6apAtAc5Ym
DMoTRP2Vo+BEm4uS4GcIFZYqrOsPuuyMuBd0NDE33g==
-----END CERTIFICATE-----
`),
					KeyPEM: []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAvh4DhQDuE/oXjEJzt1TmgmM6s5DtsIYIEPssxdDiDEjmJzhG
APd0BmAfrz3/fXZ6CqqqfAvTKkTDGjSN0hGtt1thF72weNfmQX9c1gUukEA0gERS
OsSWlzp8Azc5fjXtl5jiQkvaapNoQ8PljLc8SEsgndvdxD1tDVWN6h90bMgtuqyK
M7Fn/1+nwhRExKRTOSlR0iS9eddWPcN3HNGrl3vjCF0oUF/79+nJJqeaXawt0+2P
CMkxY6T6yzFaS/h0Oy3mXLOQ1H5n3cZ2MSEMKQSs9ff8gzeHpWyXpgB1XEKNcUvi
rVKfaJ/WE5v0rieG89B/YnQTduQnrwmPVuwm4wIDAQABAoIBAQCxKKrC7+DqwKvc
ybem6Ph8HBeBaNX1HpC5sjVAiKt8IxpFBc1F7VEy97POywkfUp3a/rorKaG2y6i6
7KoTTOIB8KcDRoIBub4Y3qQV03JWfV3vALtXhAWIGrmhDX8Hux0RnSeJ+8EmewI3
034+qCkGfOuB7nYy/cJ3IHhD6NfG3Q3FrBrGfsI2TGEeGmPJ2Xi8ZyfbluRb/1Bt
NesS6pDbRpZ5/IoauLUtITY3bazpzghm2tJNdrJIP7ohaoMF0WYciPyD5xpNlykt
V8Q2jzNmPYVXuUpi4oPekq4Mg1vd/LPS/JE558Am1LEiXrycelGNrDvJW7hTDLVx
DLRFuMMxAoGBAMkjupL3mxAfNytXM++WxJbdWPuw/vvAeN60ifFu6RUrMs/aXocn
4xSunNF58O2aRfSq/B9LJ+pXtmdITV+Bu0Y1XefKtNUNoqIapAbA8gAWUcFSkDRd
999rh0vWPbx4d3k69iS6xIjVaRcxeuaBbKRWqUcrxDuAydhwTLIRMD1vAoGBAPH4
quLGkr1MdTeZ3qPAWc9mGelp0LhHukjnLB+nMdI73OH7IlX5or11yr6La/+sTmmQ
fI+oITLuCyey7VnWBDhrPmWFGA1BmZIVDqjkJJNwyWQO7N27rQEQoNKm5n6Q+boy
StNKa/ljduYXCjsBndOmF1wSrAwL+u9rQ3x4k9vNAoGAGY5vm1LYofDFar1WvP90
FRMkxj4T99rZwLpBuKp19RmbCCvfzN51jOAuzrLmuNncP50mEbfT54OjinX2Vsc+
C0qmltf7qAJmgqBN7QnA9d/gHWcnKXAzGXEpLKqZB4Rq8b1bHwmYBSbQhoDj87vI
GQ1lzsQx17mia9zA8fMbJQMCgYB0D+2PpuW9rM3QpJp4+wtZAsVNAzddHPKKg2/T
ovOvvoz9S+M1T+8yZyyfZuqfkTtvQSGuGlwKPMnW+ekFHTWbBj3Ani1iNmP+AOGu
OvgcTI4c01fkJ2AdUaeCQxHuBYXzPKpNXLYbwgzG4qhCk0zrtxAfVsl1Yc20R0Pw
kTmCxQKBgQCzd/OOLm7vDEUqYNUNgKlf8I9xM84IwUy+pJP8RaeFDtFj7tVDpY2P
GXHBXcIBDRPnmBxC7cGHCB3KBWJp11smw2qA0ZgmBIShNm1RDHf/1h0yOxSz2+fB
bgeEDefxTxoTMgJ1urwl0KX6R9dbv9YWZWJXk2DQj6UwkMEyXpc+kw==
-----END RSA PRIVATE KEY-----
`),
				},
			}
		}),
		Entry("not a self-signed certificate", func() testCase {
			return testCase{
				expectedErr: `
                violations:
                - field: cert
                  message: "certificate must be self-signed (intermediate CAs are not allowed)"
`,
				input: util_tls.KeyPair{
					CertPEM: []byte(`
-----BEGIN CERTIFICATE-----
MIIDKzCCAhOgAwIBAgIBADANBgkqhkiG9w0BAQsFADAwMQ0wCwYDVQQKEwRLdW1h
MQ0wCwYDVQQLEwRNZXNoMRAwDgYDVQQDEwdkZWZhdWx0MB4XDTIwMDEyOTE2MDgw
NFoXDTMwMDEyNjE2MDgxNFowQDENMAsGA1UEChMES3VtYTEdMAsGA1UECxMETWVz
aDAOBgNVBAsTB2xldmVsLTExEDAOBgNVBAMTB2RlZmF1bHQwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQC+HgOFAO4T+heMQnO3VOaCYzqzkO2whggQ+yzF
0OIMSOYnOEYA93QGYB+vPf99dnoKqqp8C9MqRMMaNI3SEa23W2EXvbB41+ZBf1zW
BS6QQDSARFI6xJaXOnwDNzl+Ne2XmOJCS9pqk2hDw+WMtzxISyCd293EPW0NVY3q
H3RsyC26rIozsWf/X6fCFETEpFM5KVHSJL1511Y9w3cc0auXe+MIXShQX/v36ckm
p5pdrC3T7Y8IyTFjpPrLMVpL+HQ7LeZcs5DUfmfdxnYxIQwpBKz19/yDN4elbJem
AHVcQo1xS+KtUp9on9YTm/SuJ4bz0H9idBN25CevCY9W7CbjAgMBAAGjQDA+MA4G
A1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MBsGA1UdEQQUMBKGEHNwaWZm
ZTovL2RlZmF1bHQwDQYJKoZIhvcNAQELBQADggEBACVXnYWCCrji551pbJsOCGYJ
GEqlvcwNnnYdykas4GrfsbW2rglmaXv0uG8iH2sAH+4/MjGjnlQ6Y6Fj7mDFnidj
ugU964sEDnLuU0CtaIpHl7VZ13I0EzmfY+GsCrcIXIxbAxwWTJhz77XqbHe3baLx
Sh9wHgz/aZuy99rq9OoAvUALEaIfxrvUsVs25jLuv0Xzy57B2Dpqo0odshDA4WSS
MynQnSX7aFg1jqZQL4YjPHryEQQRj8mgjqiWp8M4/PHq5s09zDMB0DCag0QtdC/k
ydtqRoojiRS2fXY8DhFRqqRVBqLvA+7eTEKpzfjUTyEovMqxIM2n4U5MSGKQlbM=
-----END CERTIFICATE-----

`),
					KeyPEM: []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAvh4DhQDuE/oXjEJzt1TmgmM6s5DtsIYIEPssxdDiDEjmJzhG
APd0BmAfrz3/fXZ6CqqqfAvTKkTDGjSN0hGtt1thF72weNfmQX9c1gUukEA0gERS
OsSWlzp8Azc5fjXtl5jiQkvaapNoQ8PljLc8SEsgndvdxD1tDVWN6h90bMgtuqyK
M7Fn/1+nwhRExKRTOSlR0iS9eddWPcN3HNGrl3vjCF0oUF/79+nJJqeaXawt0+2P
CMkxY6T6yzFaS/h0Oy3mXLOQ1H5n3cZ2MSEMKQSs9ff8gzeHpWyXpgB1XEKNcUvi
rVKfaJ/WE5v0rieG89B/YnQTduQnrwmPVuwm4wIDAQABAoIBAQCxKKrC7+DqwKvc
ybem6Ph8HBeBaNX1HpC5sjVAiKt8IxpFBc1F7VEy97POywkfUp3a/rorKaG2y6i6
7KoTTOIB8KcDRoIBub4Y3qQV03JWfV3vALtXhAWIGrmhDX8Hux0RnSeJ+8EmewI3
034+qCkGfOuB7nYy/cJ3IHhD6NfG3Q3FrBrGfsI2TGEeGmPJ2Xi8ZyfbluRb/1Bt
NesS6pDbRpZ5/IoauLUtITY3bazpzghm2tJNdrJIP7ohaoMF0WYciPyD5xpNlykt
V8Q2jzNmPYVXuUpi4oPekq4Mg1vd/LPS/JE558Am1LEiXrycelGNrDvJW7hTDLVx
DLRFuMMxAoGBAMkjupL3mxAfNytXM++WxJbdWPuw/vvAeN60ifFu6RUrMs/aXocn
4xSunNF58O2aRfSq/B9LJ+pXtmdITV+Bu0Y1XefKtNUNoqIapAbA8gAWUcFSkDRd
999rh0vWPbx4d3k69iS6xIjVaRcxeuaBbKRWqUcrxDuAydhwTLIRMD1vAoGBAPH4
quLGkr1MdTeZ3qPAWc9mGelp0LhHukjnLB+nMdI73OH7IlX5or11yr6La/+sTmmQ
fI+oITLuCyey7VnWBDhrPmWFGA1BmZIVDqjkJJNwyWQO7N27rQEQoNKm5n6Q+boy
StNKa/ljduYXCjsBndOmF1wSrAwL+u9rQ3x4k9vNAoGAGY5vm1LYofDFar1WvP90
FRMkxj4T99rZwLpBuKp19RmbCCvfzN51jOAuzrLmuNncP50mEbfT54OjinX2Vsc+
C0qmltf7qAJmgqBN7QnA9d/gHWcnKXAzGXEpLKqZB4Rq8b1bHwmYBSbQhoDj87vI
GQ1lzsQx17mia9zA8fMbJQMCgYB0D+2PpuW9rM3QpJp4+wtZAsVNAzddHPKKg2/T
ovOvvoz9S+M1T+8yZyyfZuqfkTtvQSGuGlwKPMnW+ekFHTWbBj3Ani1iNmP+AOGu
OvgcTI4c01fkJ2AdUaeCQxHuBYXzPKpNXLYbwgzG4qhCk0zrtxAfVsl1Yc20R0Pw
kTmCxQKBgQCzd/OOLm7vDEUqYNUNgKlf8I9xM84IwUy+pJP8RaeFDtFj7tVDpY2P
GXHBXcIBDRPnmBxC7cGHCB3KBWJp11smw2qA0ZgmBIShNm1RDHf/1h0yOxSz2+fB
bgeEDefxTxoTMgJ1urwl0KX6R9dbv9YWZWJXk2DQj6UwkMEyXpc+kw==
-----END RSA PRIVATE KEY-----
`),
				},
			}
		}),
		Entry("certificate without basic constraint `CA`", func() testCase {
			// when
			keyPair, err := NewSelfSignedCert(func() *x509.Certificate {
				return &x509.Certificate{
					SerialNumber: big.NewInt(0),
				}
			})
			// then
			Expect(err).ToNot(HaveOccurred())

			return testCase{
				expectedErr: `
                violations:
                - field: cert
                  message: "basic constraint 'CA' must be set to 'true' (see X509-SVID: 4.1. Basic Constraints)"
                - field: cert
                  message: "key usage extension 'keyCertSign' must be set (see X509-SVID: 4.3. Key Usage)"
`,
				input: *keyPair,
			}
		}),
		Entry("certificate without key usage extension `keyCertSign`", func() testCase {
			// when
			keyPair, err := NewSelfSignedCert(func() *x509.Certificate {
				return &x509.Certificate{
					SerialNumber:          big.NewInt(0),
					BasicConstraintsValid: true,
					IsCA:                  true,
				}
			})
			// then
			Expect(err).ToNot(HaveOccurred())

			return testCase{
				expectedErr: `
                violations:
                - field: cert
                  message: "key usage extension 'keyCertSign' must be set (see X509-SVID: 4.3. Key Usage)"
`,
				input: *keyPair,
			}
		}),
		Entry("certificate with key usage extension `digitalSignature`", func() testCase {
			// when
			keyPair, err := NewSelfSignedCert(func() *x509.Certificate {
				return &x509.Certificate{
					SerialNumber:          big.NewInt(0),
					BasicConstraintsValid: true,
					IsCA:                  true,
					KeyUsage: x509.KeyUsageCertSign |
						x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
				}
			})
			// then
			Expect(err).ToNot(HaveOccurred())

			return testCase{
				expectedErr: `
                violations:
                - field: cert
                  message: "key usage extension 'digitalSignature' must NOT be set (see X509-SVID: Appendix A. X.509 Field Reference)"
`,
				input: *keyPair,
			}
		}),
		Entry("certificate with key usage extension `keyAgreement`", func() testCase {
			// when
			keyPair, err := NewSelfSignedCert(func() *x509.Certificate {
				return &x509.Certificate{
					SerialNumber:          big.NewInt(0),
					BasicConstraintsValid: true,
					IsCA:                  true,
					KeyUsage: x509.KeyUsageCertSign |
						x509.KeyUsageCRLSign | x509.KeyUsageKeyAgreement,
				}
			})
			// then
			Expect(err).ToNot(HaveOccurred())

			return testCase{
				expectedErr: `
                violations:
                - field: cert
                  message: "key usage extension 'keyAgreement' must NOT be set (see X509-SVID: Appendix A. X.509 Field Reference)"
`,
				input: *keyPair,
			}
		}),
		Entry("certificate with key usage extension `keyEncipherment`", func() testCase {
			// when
			keyPair, err := NewSelfSignedCert(func() *x509.Certificate {
				return &x509.Certificate{
					SerialNumber:          big.NewInt(0),
					BasicConstraintsValid: true,
					IsCA:                  true,
					KeyUsage: x509.KeyUsageCertSign |
						x509.KeyUsageCRLSign | x509.KeyUsageKeyEncipherment,
				}
			})
			// then
			Expect(err).ToNot(HaveOccurred())

			return testCase{
				expectedErr: `
                violations:
                - field: cert
                  message: "key usage extension 'keyEncipherment' must NOT be set (see X509-SVID: Appendix A. X.509 Field Reference)"
`,
				input: *keyPair,
			}
		}),
		Entry("certificate with multiple violations", func() testCase {
			// when
			keyPair, err := NewSelfSignedCert(func() *x509.Certificate {
				return &x509.Certificate{
					SerialNumber: big.NewInt(0),
					IsCA:         false,
					KeyUsage: x509.KeyUsageDigitalSignature |
						x509.KeyUsageKeyAgreement | x509.KeyUsageKeyEncipherment,
				}
			})
			// then
			Expect(err).ToNot(HaveOccurred())

			return testCase{
				expectedErr: `
                violations:
                - field: cert
                  message: "basic constraint 'CA' must be set to 'true' (see X509-SVID: 4.1. Basic Constraints)"
                - field: cert
                  message: "key usage extension 'keyCertSign' must be set (see X509-SVID: 4.3. Key Usage)"
                - field: cert
                  message: "key usage extension 'digitalSignature' must NOT be set (see X509-SVID: Appendix A. X.509 Field Reference)"
                - field: cert
                  message: "key usage extension 'keyAgreement' must NOT be set (see X509-SVID: Appendix A. X.509 Field Reference)"
                - field: cert
                  message: "key usage extension 'keyEncipherment' must NOT be set (see X509-SVID: Appendix A. X.509 Field Reference)"
`,
				input: *keyPair,
			}
		}),
	)
})
