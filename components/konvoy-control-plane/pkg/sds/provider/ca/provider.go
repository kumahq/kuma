package ca

import (
	"context"

	sds_auth "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/sds/auth"
	sds_provider "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/sds/provider"
)

func New() sds_provider.SecretProvider {
	return &meshCaProvider{}
}

type meshCaProvider struct {
}

func (s *meshCaProvider) RequiresIdentity() bool {
	return false
}

func (s *meshCaProvider) Get(ctx context.Context, name string, requestor sds_auth.Identity) (sds_provider.Secret, error) {
	return &MeshCaSecret{
		PemCerts: [][]byte{[]byte(`
-----BEGIN CERTIFICATE-----
MIIC/TCCAeWgAwIBAgIRAKgVvRCrr87LMCVhIoU2k2gwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAeFw0xOTA4MzAxMzQ4NDlaFw0yOTA4MzAwMTQ4
NDlaMBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQDTzb7dLh/7o/5eNatxF4RGRaXKJw85OjwmJKNESJ/4+ykFHOTh2Cg3
BK5E9BXTWlEWMXlDTeMT3sqsJLyKCcQ38G64ue/gxvmqu2fMIXL/kABkta1gOXJF
/QNCf/bjln7gxeTQzwfHHFuMCxq7qq0pdXkvp/gxXDKpJiwCYwOpZ5eT0lWbfUk6
s/pOFniUSlLNg8gi3nNNUSZDVTN9HpivXWVO3IAWkupxYV/8rwwyzb5Z4jheFDVj
G+It6BmItQK7snmoB4f1S8ELUT62nBu/pyyc6Y5AIN2wbzccdhhBBbTUHkZb7RoV
l2fbgnoKFLNNXvYxKAeqq65EhnYXR1fXAgMBAAGjTjBMMA4GA1UdDwEB/wQEAwIC
pDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MBQGA1UdEQQN
MAuCCWxvY2FsaG9zdDANBgkqhkiG9w0BAQsFAAOCAQEAXpxcNfSfEUIwOgfPtDYV
ZzzLxx9m27m9cYyjaXWDFdXVDO/MTPjr/bf/7QVhf3ofoF9l/ojwkl+MwmhkV0Jx
W2kfIyVQpsRh2KvYVv74Zn610zbvMLSTyEHuG2UWO0UeZfOsHmTEJsuJH54GEYJh
DYHKHl31pG9OF1CwAZNyhmOz3nFrzRJr695rR0ZKwYvAfaTA/VUEymmPX3RiHf6W
i18K5lY9hrNfNgI4gMVyNaWbAtJASzNdH1TIlvHv+P5pXqE2okfVFG/Gaw7tBfCJ
EpmbGZ7aOrJIX9zqzgYKMg0ALF0jzKwAcxkyyQpcvGQbN1qMMbNYwrJPdQ6eK6AQ
5w==
-----END CERTIFICATE-----
`),
		},
	}, nil
}
