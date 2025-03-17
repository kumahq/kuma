package tls

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"

	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	util_tls "github.com/kumahq/kuma/pkg/tls"
)

var globalSecretKey = model.ResourceKey{
	Name: system.InterCpCA,
}

func GenerateCA() (*util_tls.KeyPair, error) {
	subject := pkix.Name{
		Organization:       []string{"Kuma"},
		OrganizationalUnit: []string{"Mesh"},
		CommonName:         "Control Plane Intercommunication CA",
	}
	return util_tls.GenerateCA(util_tls.DefaultKeyType, subject)
}

func LoadCA(ctx context.Context, resManager manager.ReadOnlyResourceManager) (tls.Certificate, error) {
	globalSecret := system.NewGlobalSecretResource()
	if err := resManager.Get(ctx, globalSecret, store.GetBy(globalSecretKey)); err != nil {
		return tls.Certificate{}, err
	}
	bytes := globalSecret.Spec.GetData().GetValue()
	certBlock, rest := pem.Decode(bytes)
	keyBlock, _ := pem.Decode(rest)
	return tls.X509KeyPair(pem.EncodeToMemory(certBlock), pem.EncodeToMemory(keyBlock))
}

func CreateCA(ctx context.Context, keyPair util_tls.KeyPair, resManager manager.ResourceManager) error {
	bytes := append(keyPair.CertPEM, keyPair.KeyPEM...)
	globalSecret := system.NewGlobalSecretResource()
	globalSecret.Spec.Data = &wrapperspb.BytesValue{
		Value: bytes,
	}
	return resManager.Create(ctx, globalSecret, store.CreateBy(globalSecretKey))
}

func GenerateClientCert(ca tls.Certificate, ip string) (tls.Certificate, error) {
	rootCert, err := x509.ParseCertificate(ca.Certificate[0])
	if err != nil {
		return tls.Certificate{}, err
	}
	pair, err := util_tls.NewCert(*rootCert, ca.PrivateKey.(*rsa.PrivateKey), util_tls.ClientCertType, util_tls.DefaultKeyType, ip)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.X509KeyPair(pair.CertPEM, pair.KeyPEM)
}

func GenerateServerCert(ca tls.Certificate, ip string) (tls.Certificate, error) {
	rootCert, err := x509.ParseCertificate(ca.Certificate[0])
	if err != nil {
		return tls.Certificate{}, err
	}
	pair, err := util_tls.NewCert(*rootCert, ca.PrivateKey.(*rsa.PrivateKey), util_tls.ServerCertType, util_tls.DefaultKeyType, ip)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.X509KeyPair(pair.CertPEM, pair.KeyPEM)
}
