package secrets

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/metrics"
)

var log = core.Log.WithName("xds").WithName("secrets")

type Secrets interface {
	Get(dataplane *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) (*core_xds.IdentitySecret, *core_xds.CaSecret, error)
	Info(dpKey model.ResourceKey) *Info
	Cleanup(dpKey model.ResourceKey)
}

type Info struct {
	Expiration time.Time
	Generation time.Time

	Tags mesh_proto.MultiValueTagSet
	MTLS *mesh_proto.Mesh_Mtls

	IssuedBackend     string
	SupportedBackends []string
}

func (c *Info) CertLifetime() time.Duration {
	return c.Expiration.Sub(c.Generation)
}

func (c *Info) ExpiringSoon() bool {
	return core.Now().After(c.Generation.Add(c.CertLifetime() / 5 * 4))
}

func NewSecrets(caProvider CaProvider, identityProvider IdentityProvider, metrics metrics.Metrics) (Secrets, error) {
	certGenerationsMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Help: "Number of generated certificates",
		Name: "cert_generation",
	}, []string{"mesh"})
	if err := metrics.Register(certGenerationsMetric); err != nil {
		return nil, err
	}

	return &secrets{
		caProvider:            caProvider,
		identityProvider:      identityProvider,
		cachedCerts:           map[model.ResourceKey]*certs{},
		certGenerationsMetric: certGenerationsMetric,
	}, nil
}

type secrets struct {
	caProvider       CaProvider
	identityProvider IdentityProvider

	sync.RWMutex
	cachedCerts           map[model.ResourceKey]*certs
	certGenerationsMetric *prometheus.CounterVec
}

var _ Secrets = &secrets{}

func (c *secrets) Info(dpKey model.ResourceKey) *Info {
	certs := c.certs(dpKey)
	if certs == nil {
		return nil
	}
	return certs.info
}

type certs struct {
	identity *core_xds.IdentitySecret
	ca       *core_xds.CaSecret
	info     *Info
}

func (c *certs) Info() *Info {
	if c == nil {
		return nil
	}
	return c.info
}

func (c *secrets) certs(dpKey model.ResourceKey) *certs {
	c.RLock()
	defer c.RUnlock()
	return c.cachedCerts[dpKey]
}

func (c *secrets) Get(dataplane *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
	if !mesh.MTLSEnabled() {
		return nil, nil, nil
	}

	dpKey := model.MetaToResourceKey(dataplane.GetMeta())
	certs := c.certs(dpKey)
	tags := dataplane.Spec.TagSet()
	if shouldGenerate, reason := c.shouldGenerateCerts(certs.Info(), mesh.Spec.Mtls, tags); shouldGenerate {
		log.Info("generating certificate", "dp", dpKey, "reason", reason)
		certs, err := c.generateCerts(dataplane, mesh)
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not generate certificates")
		}
		c.Lock()
		c.cachedCerts[dpKey] = certs
		c.Unlock()
		return certs.identity, certs.ca, nil
	}

	certs = c.certs(dpKey)
	if certs == nil { // previous "if" should guarantee that the certs are always there
		return nil, nil, errors.New("certificates were not generated")
	}
	return certs.identity, certs.ca, nil
}

func (c *secrets) Cleanup(dpKey model.ResourceKey) {
	c.Lock()
	delete(c.cachedCerts, dpKey)
	c.Unlock()
}

func (c *secrets) shouldGenerateCerts(info *Info, mtls *mesh_proto.Mesh_Mtls, tags mesh_proto.MultiValueTagSet) (bool, string) {
	if info == nil {
		return true, "mTLS is enabled and DP hasn't received a certificate yet"
	}

	if !proto.Equal(info.MTLS, mtls) {
		return true, "Mesh mTLS settings have changed"
	}

	if tags.String() != info.Tags.String() {
		return true, "DP tags have changed"
	}

	if info.ExpiringSoon() {
		return true, fmt.Sprintf("the certificate expiring soon. Generated at %q, expiring at %q", info.Generation, info.Expiration)
	}
	return false, ""
}

func (c *secrets) generateCerts(dataplane *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) (*certs, error) {
	tags := dataplane.Spec.TagSet()
	requestor := Identity{
		Services: tags,
		Mesh:     dataplane.GetMeta().GetMesh(),
	}
	identity, issuedBackend, err := c.identityProvider.Get(context.Background(), requestor, mesh)
	if err != nil {
		return nil, errors.Wrap(err, "could not get Dataplane cert pair")
	}
	c.certGenerationsMetric.WithLabelValues(requestor.Mesh).Inc()

	ca, supportedBackends, err := c.caProvider.Get(context.Background(), mesh)
	if err != nil {
		return nil, errors.Wrap(err, "could not get mesh CA cert")
	}

	info, err := newCertInfo(identity, mesh.Spec.Mtls, tags, issuedBackend, supportedBackends)
	if err != nil {
		return nil, errors.Wrap(err, "could not extract info about certificate")
	}

	return &certs{
		identity: identity,
		ca:       ca,
		info:     info,
	}, nil
}

func newCertInfo(identityCert *core_xds.IdentitySecret, mtls *mesh_proto.Mesh_Mtls, tags mesh_proto.MultiValueTagSet, issuedBackend string, supportedBackends []string) (*Info, error) {
	block, _ := pem.Decode(identityCert.PemCerts[0])
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	certInfo := &Info{
		Tags:              tags,
		MTLS:              mtls,
		Expiration:        cert.NotAfter,
		Generation:        core.Now(),
		IssuedBackend:     issuedBackend,
		SupportedBackends: supportedBackends,
	}
	return certInfo, nil
}
