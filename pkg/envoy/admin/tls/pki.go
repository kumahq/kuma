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

const (
	ClientCertSAN = "kuma-cp"
)

var globalSecretKey = model.ResourceKey{
	Name: system.EnvoyAdminCA,
}

// GenerateCA generates CA for Envoy Admin communication (CP sending requests to Envoy Admin).
// While we could reuse CA from enable mTLS backend on a Mesh object there are two problems
//  1. mTLS on Mesh can be disabled and Envoy Admin communication needs security in place.
//     Otherwise, malicious actor could execute /quitquitquit endpoint and perform DDoS
//  2. ZoneIngress and ZoneEgress are not scoped to a Mesh.
//
// To solve this we need at least self-signed client certificate for the control plane.
// But we can just as well have a CA and generate client and server certs from it.
//
// Rotation: users can change the CA. To do this, they can swap the secret and restart all instances of the CP.
// Multizone: CA is generated for every zone. There is no need for it to be stable.
func GenerateCA() (*util_tls.KeyPair, error) {
	subject := pkix.Name{
		Organization:       []string{"Kuma"},
		OrganizationalUnit: []string{"Mesh"},
		CommonName:         "Envoy Admin CA",
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

func GenerateClientCert(ca tls.Certificate) (util_tls.KeyPair, error) {
	rootCert, err := x509.ParseCertificate(ca.Certificate[0])
	if err != nil {
		return util_tls.KeyPair{}, err
	}
	return util_tls.NewCert(*rootCert, ca.PrivateKey.(*rsa.PrivateKey), util_tls.ClientCertType, util_tls.DefaultKeyType, ClientCertSAN)
}

func GenerateServerCert(ca tls.Certificate, hosts ...string) (util_tls.KeyPair, error) {
	rootCert, err := x509.ParseCertificate(ca.Certificate[0])
	if err != nil {
		return util_tls.KeyPair{}, err
	}
	return util_tls.NewCert(*rootCert, ca.PrivateKey.(*rsa.PrivateKey), util_tls.ServerCertType, util_tls.DefaultKeyType, hosts...)
}
