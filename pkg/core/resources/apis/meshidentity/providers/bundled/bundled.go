package bundled

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"reflect"
	"time"

	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/kri"
	meshidentity_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/metadata"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/providers"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	bldrs_auth "github.com/kumahq/kuma/pkg/envoy/builders/auth"
	bldrs_common "github.com/kumahq/kuma/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/pkg/envoy/builders/core"
	bldrs_tls "github.com/kumahq/kuma/pkg/envoy/builders/tls"
	"github.com/kumahq/kuma/pkg/metrics"
	util_tls "github.com/kumahq/kuma/pkg/tls"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/cache/once"
)

const (
	DefaultAllowedClockSkew = 10 * time.Second

	cacheExpirationTime = 5 * time.Second
	caCacheEntryKey     = "ca_pair"
)

var DefaultWorkloadCertValidityPeriod = k8s.Duration{Duration: 24 * time.Hour}

var _ providers.IdentityProvider = &bundledIdentityProvider{}

type bundledIdentityProvider struct {
	logger          logr.Logger
	roSecretManager manager.ReadOnlyResourceManager
	secretManager   manager.ResourceManager
	cache           *once.Cache
	zone            string
}

func NewBundledIdentityProvider(roSecretManager manager.ReadOnlyResourceManager, secretManager manager.ResourceManager, metrics metrics.Metrics, zone string) (providers.IdentityProvider, error) {
	c, err := once.New(cacheExpirationTime, "ca_cache", metrics)
	if err != nil {
		return nil, err
	}
	logger := core.Log.WithName("identity-provider").WithName("bundled")
	return &bundledIdentityProvider{
		logger:          logger,
		roSecretManager: roSecretManager,
		secretManager:   secretManager,
		cache:           c,
		zone:            zone,
	}, nil
}

func (b *bundledIdentityProvider) Validate(ctx context.Context, identity *meshidentity_api.MeshIdentityResource) error {
	if !pointer.DerefOr(identity.Spec.Provider.Bundled.InsecureAllowSelfSigned, false) {
		if pointer.DerefOr(identity.Spec.Provider.Bundled.Autogenerate.Enabled, false) {
			return errors.Errorf("self-signed certificates are not allowed")
		}
		ca, err := b.GetRootCA(ctx, identity)
		if err != nil {
			return err
		}
		selfSigned, err := isSelfSigned(ca)
		if err != nil {
			return err
		}
		if selfSigned {
			return errors.Errorf("self-signed certificates are not allowed")
		}
	}
	b.logger.V(1).Info("identity is valid", "identity", model.MetaToResourceKey(identity.GetMeta()))
	return nil
}

func (b *bundledIdentityProvider) Initialize(ctx context.Context, identity *meshidentity_api.MeshIdentityResource) error {
	if pointer.DerefOr(identity.Spec.Provider.Bundled.Autogenerate.Enabled, false) {
		if _, err := b.getCAKeyPair(ctx, identity, identity.Meta.GetMesh()); err != nil && core_store.IsNotFound(err) {
			trustDomain, err := identity.Spec.GetTrustDomain(identity.GetMeta(), b.zone)
			if err != nil {
				return err
			}
			b.logger.V(1).Info("initializing provider", "identity", model.MetaToResourceKey(identity.GetMeta()), "trustDomain", trustDomain)
			keyPair, err := GenerateRootCA(trustDomain)
			if err != nil {
				return err
			}
			certSecret := &core_system.SecretResource{
				Spec: &system_proto.Secret{
					Data: util_proto.Bytes(keyPair.CertPEM),
				},
			}
			mesh := identity.Meta.GetMesh()
			if err := b.secretManager.Create(ctx, certSecret, core_store.CreateWithOwner(identity), core_store.CreateByKey(RootCAName(model.GetDisplayName(identity.GetMeta())), mesh)); err != nil {
				if !core_store.IsAlreadyExists(err) {
					return err
				}
			}
			keySecret := &core_system.SecretResource{
				Spec: &system_proto.Secret{
					Data: util_proto.Bytes(keyPair.KeyPEM),
				},
			}
			if err := b.secretManager.Create(ctx, keySecret, core_store.CreateWithOwner(identity), core_store.CreateByKey(PrivateKeyName(model.GetDisplayName(identity.GetMeta())), mesh)); err != nil {
				if !core_store.IsAlreadyExists(err) {
					return err
				}
			}
			b.logger.V(1).Info("initialized", "identity", model.MetaToResourceKey(identity.GetMeta()), "trustDomain", trustDomain)
		} else if err != nil {
			return err
		} else {
			b.logger.V(1).Info("provider already initialized", "identity", model.MetaToResourceKey(identity.GetMeta()))
		}
	}
	return nil
}

func (b *bundledIdentityProvider) GetRootCA(ctx context.Context, identity *meshidentity_api.MeshIdentityResource) ([]byte, error) {
	bundled := pointer.Deref(identity.Spec.Provider.Bundled)
	var err error
	var cert []byte
	if bundled.Autogenerate == nil || !pointer.Deref(bundled.Autogenerate.Enabled) && bundled.CA != nil {
		cert, err = bundled.CA.Certificate.ReadByControlPlane(ctx, b.secretManager, identity.Meta.GetMesh())
		if err != nil {
			return nil, err
		}
	} else {
		ca := core_system.NewSecretResource()
		if err := b.roSecretManager.Get(ctx, ca, core_store.GetByKey(RootCAName(model.GetDisplayName(identity.GetMeta())), identity.Meta.GetMesh())); err != nil {
			return nil, err
		}
		cert = ca.Spec.Data.GetValue()
	}
	return cert, err
}

// Instead of loading the CA pair on each dataplane workload identity generation,
// we can cache it and refresh the cache periodically (e.g., every few seconds).
// This reduces the load on the underlying store (e.g., OS, DB), as the CA pair doesn't change frequently.
func (b *bundledIdentityProvider) getCAKeyPair(ctx context.Context, identity *meshidentity_api.MeshIdentityResource, mesh string) (*util_tls.KeyPair, error) {
	ca, err := b.cache.GetOrRetrieve(ctx, caCacheEntryKey, once.RetrieverFunc(func(ctx context.Context, cacheKey string) (interface{}, error) {
		bundled := pointer.Deref(identity.Spec.Provider.Bundled)
		var err error
		var cert, key []byte
		if (bundled.Autogenerate == nil || !pointer.Deref(bundled.Autogenerate.Enabled)) && bundled.CA != nil {
			cert, err = bundled.CA.Certificate.ReadByControlPlane(ctx, b.secretManager, mesh)
			if err != nil {
				return nil, err
			}
			key, err = bundled.CA.PrivateKey.ReadByControlPlane(ctx, b.secretManager, mesh)
			if err != nil {
				return nil, err
			}
		} else {
			// ca
			ca := core_system.NewSecretResource()
			if err := b.roSecretManager.Get(ctx, ca, core_store.GetByKey(RootCAName(model.GetDisplayName(identity.GetMeta())), mesh)); err != nil {
				return nil, err
			}
			cert = ca.Spec.Data.GetValue()
			// privateKey
			caKey := core_system.NewSecretResource()
			if err := b.roSecretManager.Get(ctx, caKey, core_store.GetByKey(PrivateKeyName(model.GetDisplayName(identity.GetMeta())), mesh)); err != nil {
				return nil, err
			}
			key = caKey.Spec.Data.GetValue()
		}
		return &util_tls.KeyPair{
			CertPEM: cert,
			KeyPEM:  key,
		}, nil
	}))
	if err != nil {
		return nil, err
	}
	return ca.(*util_tls.KeyPair), nil
}

func (b *bundledIdentityProvider) CreateIdentity(ctx context.Context, identity *meshidentity_api.MeshIdentityResource, proxy *xds.Proxy) (*xds.WorkloadIdentity, error) {
	pair, err := b.getCAKeyPair(ctx, identity, proxy.Dataplane.Meta.GetMesh())
	if err != nil {
		return nil, err
	}
	caCert, caPrivateKey, err := loadKeyPair(pointer.Deref(pair))
	if err != nil {
		return nil, errors.Wrap(err, "failed to load CA key pair")
	}
	publicKey, privateKey, err := generateKey(caPrivateKey)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	serialNumber, err := newSerialNumber()
	if err != nil {
		return nil, err
	}
	trustDomain, err := identity.Spec.GetTrustDomain(identity.GetMeta(), b.zone)
	if err != nil {
		return nil, err
	}
	spiffeID, err := identity.Spec.GetSpiffeID(trustDomain, proxy.Dataplane.GetMeta())
	if err != nil {
		return nil, err
	}
	id, err := spiffeid.FromString(spiffeID)
	if err != nil {
		return nil, err
	}
	certValidity := pointer.DerefOr(identity.Spec.Provider.Bundled.CertificateParameters.Expiry, DefaultWorkloadCertValidityPeriod)
	b.logger.V(1).Info("creating an identity", "dpp", model.MetaToResourceKey(proxy.Dataplane.GetMeta()), "spiffeId", spiffeID, "identity", model.MetaToResourceKey(identity.GetMeta()))

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		URIs:         []*url.URL{id.URL()},
		NotBefore:    now.Add(-DefaultAllowedClockSkew),
		NotAfter:     now.Add(certValidity.Duration),
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageKeyAgreement |
			x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
		PublicKey:             publicKey,
	}
	workloadCert, err := x509.CreateCertificate(rand.Reader, template, caCert, publicKey, caPrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate X509 certificate")
	}
	identityPair, err := util_tls.ToKeyPair(privateKey, workloadCert)
	if err != nil {
		return nil, err
	}
	identifier := kri.From(identity)
	resources, err := additionalResources(identifier.String(), identityPair)
	if err != nil {
		return nil, err
	}
	b.logger.V(1).Info("identity created", "dpp", model.MetaToResourceKey(proxy.Dataplane.GetMeta()), "spiffeId", spiffeID, "identity", model.MetaToResourceKey(identity.GetMeta()))

	return &xds.WorkloadIdentity{
		KRI:                      identifier,
		ManagementMode:           xds.KumaManagementMode,
		ExpirationTime:           pointer.To(template.NotAfter),
		GenerationTime:           pointer.To(now),
		IdentitySourceConfigurer: sourceConfigurer(identifier.String()),
		AdditionalResources:      resources,
	}, nil
}

func additionalResources(secretName string, keyPair *util_tls.KeyPair) (*xds.ResourceSet, error) {
	resources := xds.NewResourceSet()
	identitySecret, err := bldrs_auth.NewSecret().
		Configure(bldrs_auth.Name(secretName)).
		Configure(bldrs_auth.TlsCertificate(
			bldrs_auth.NewTlsCertificate().
				Configure(bldrs_auth.CertificateChain(
					bldrs_core.NewDataSource().
						Configure(bldrs_core.InlineBytes(bytes.Join([][]byte{keyPair.CertPEM}, []byte("\n")))))).
				Configure(bldrs_auth.PrivateKey(
					bldrs_core.NewDataSource().
						Configure(bldrs_core.InlineBytes(keyPair.KeyPEM)))))).Build()
	if err != nil {
		return nil, err
	}
	resources.Add(&xds.Resource{
		Name:     secretName,
		Origin:   metadata.OriginIdentityBundled,
		Resource: identitySecret,
	})
	return resources, nil
}

func sourceConfigurer(secretName string) func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
	return func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
		return bldrs_tls.SdsSecretConfigSource(
			secretName,
			bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
		)
	}
}

func generateKey(caPrivateKey crypto.PrivateKey) (crypto.PublicKey, crypto.PrivateKey, error) {
	var err error
	var publicKey crypto.PublicKey
	var privateKey crypto.PrivateKey

	switch caKey := caPrivateKey.(type) {
	case ed25519.PrivateKey:
		var pub ed25519.PublicKey
		var priv ed25519.PrivateKey
		pub, priv, err = ed25519.GenerateKey(nil)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate ed25519 key: %w", err)
		}
		publicKey = pub
		privateKey = priv

	case *rsa.PrivateKey:
		priv, err := rsa.GenerateKey(rand.Reader, defaultKeySize)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate rsa key: %w", err)
		}
		publicKey = &priv.PublicKey
		privateKey = priv
	case *ecdsa.PrivateKey:
		priv, err := ecdsa.GenerateKey(caKey.Curve, rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate ecdsa key: %w", err)
		}
		publicKey = &priv.PublicKey
		privateKey = priv
	default:
		return nil, nil, errors.New("unsupported CA key type")
	}
	return publicKey, privateKey, err
}

var maxUint128, one *big.Int

func init() {
	one = big.NewInt(1)
	m := new(big.Int)
	m.Lsh(one, 128)
	maxUint128 = m.Sub(m, one)
}

func isSelfSigned(certPEM []byte) (bool, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return false, errors.New("failed to decode PEM block or block is not of type CERTIFICATE")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, err
	}
	// Check if subject and issuer are the same
	if !reflect.DeepEqual(cert.Subject.ToRDNSequence(), cert.Issuer.ToRDNSequence()) {
		return false, nil
	}
	// Try to verify the certificate using its own public key
	err = cert.CheckSignatureFrom(cert)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func newSerialNumber() (*big.Int, error) {
	res, err := rand.Int(rand.Reader, maxUint128)
	if err != nil {
		return nil, fmt.Errorf("failed generation of serial number: %w", err)
	}
	// Because we generate in the range [0, maxUint128) and 0 is an invalid serial and maxUint128 is valid we add 1
	// to have a number in range [1, maxUint128] See: https://cabforum.org/2016/03/31/ballot-164/
	return res.Add(res, one), nil
}

func loadKeyPair(pair util_tls.KeyPair) (*x509.Certificate, crypto.PrivateKey, error) {
	root, err := tls.X509KeyPair(pair.CertPEM, pair.KeyPEM)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse TLS key pair")
	}
	rootCert, err := x509.ParseCertificate(root.Certificate[0])
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse X509 certificate")
	}
	return rootCert, root.PrivateKey, nil
}
