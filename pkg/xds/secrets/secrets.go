package secrets

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sethvargo/go-retry"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/user"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/metrics"
)

var log = core.Log.WithName("xds").WithName("secrets")

// certBackoffMinBase is the floor applied to a misconfigured (<= 0) base
// backoff. It guards against retry.NewExponential(0) panicking, since
// GeneralConfig.Validate() is not currently wired into Config.Validate().
const certBackoffMinBase = time.Second

// CertBackoff returns a factory for the backoff used to throttle certificate
// generation retries after a failure so that a misconfigured or failing CA
// backend can't be hammered on every DP sync tick (which defaults to 1s). The
// backoff is exponential from base, capped at max, with jitter so proxies don't
// retry in lockstep. base and max are configured via KUMA_GENERAL_CERT_GENERATION_*.
func CertBackoff(base, maxBackoff time.Duration) func() retry.Backoff {
	if base <= 0 {
		base = certBackoffMinBase
	}
	if maxBackoff < base {
		maxBackoff = base
	}
	return func() retry.Backoff {
		return retry.WithJitter(base, retry.WithCappedDuration(maxBackoff, retry.NewExponential(base)))
	}
}

type MeshCa struct {
	Mesh     string
	CaSecret *core_xds.CaSecret
}

type Secrets interface {
	GetForDataPlane(ctx context.Context, dataplane *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource, otherMeshes []*core_mesh.MeshResource) (*core_xds.IdentitySecret, map[string]*core_xds.CaSecret, error)
	GetForZoneEgress(ctx context.Context, zoneEgress *core_mesh.ZoneEgressResource, mesh *core_mesh.MeshResource) (*core_xds.IdentitySecret, *core_xds.CaSecret, error)
	GetAllInOne(ctx context.Context, mesh *core_mesh.MeshResource, dataplane *core_mesh.DataplaneResource, otherMeshes []*core_mesh.MeshResource) (*core_xds.IdentitySecret, *core_xds.CaSecret, error)
	Info(mesh_proto.ProxyType, model.ResourceKey) *Info
	Cleanup(mesh_proto.ProxyType, model.ResourceKey)
}

type MeshInfo struct {
	MTLS *mesh_proto.Mesh_Mtls
}

type Info struct {
	Expiration time.Time
	Generation time.Time

	Tags mesh_proto.MultiValueTagSet

	IssuedBackend     string
	SupportedBackends []string

	OwnMesh        MeshInfo
	OtherMeshInfos map[string]MeshInfo
	// this marks our info as having failed last time to get the mesh CAs that
	// we wanted and so we should retry next time we want certs.
	failedOtherMeshes bool

	// adds information if the secrets is delivered and managed by a different SDS server
	ManagedExternally bool
}

func (c *Info) CertLifetime() time.Duration {
	return c.Expiration.Sub(c.Generation)
}

func (c *Info) ExpiringSoon() bool {
	return core.Now().After(c.Generation.Add(c.CertLifetime() / 5 * 4))
}

func NewSecrets(caProvider CaProvider, identityProvider IdentityProvider, metrics metrics.Metrics, newBackoff func() retry.Backoff) (Secrets, error) {
	certGenerationsMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Help: "Number of generated certificates",
		Name: "cert_generation",
	}, []string{"mesh"})
	if err := metrics.Register(certGenerationsMetric); err != nil {
		return nil, err
	}

	certGenerationFailuresMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Help: "Number of failed certificate generations",
		Name: "cert_generation_failure",
	}, []string{"mesh"})
	if err := metrics.Register(certGenerationFailuresMetric); err != nil {
		return nil, err
	}

	return &secrets{
		caProvider:                   caProvider,
		identityProvider:             identityProvider,
		newBackoff:                   newBackoff,
		cachedCerts:                  map[certCacheKey]*certs{},
		backoffs:                     map[certCacheKey]*certBackoff{},
		certGenerationsMetric:        certGenerationsMetric,
		certGenerationFailuresMetric: certGenerationFailuresMetric,
	}, nil
}

type certCacheKey struct {
	resource  model.ResourceKey
	proxyType mesh_proto.ProxyType
}

// certBackoff tracks when the next certificate generation attempt is allowed
// for a given proxy after a failure.
type certBackoff struct {
	backoff retry.Backoff
	nextTry time.Time
}

type secrets struct {
	caProvider       CaProvider
	identityProvider IdentityProvider
	newBackoff       func() retry.Backoff

	sync.RWMutex
	cachedCerts                  map[certCacheKey]*certs
	backoffs                     map[certCacheKey]*certBackoff
	certGenerationsMetric        *prometheus.CounterVec
	certGenerationFailuresMetric *prometheus.CounterVec
}

var _ Secrets = &secrets{}

func (s *secrets) Info(proxyType mesh_proto.ProxyType, dpKey model.ResourceKey) *Info {
	certs := s.certs(proxyType, dpKey)
	if certs == nil {
		return nil
	}
	return certs.info
}

type certs struct {
	identity   *core_xds.IdentitySecret
	ownCa      MeshCa
	otherCas   []MeshCa
	allInOneCa MeshCa
	info       *Info
}

func (c *certs) Info() *Info {
	if c == nil {
		return nil
	}
	return c.info
}

func (s *secrets) certs(proxyType mesh_proto.ProxyType, dpKey model.ResourceKey) *certs {
	s.RLock()
	defer s.RUnlock()

	return s.cachedCerts[certCacheKey{proxyType: proxyType, resource: dpKey}]
}

func (s *secrets) GetForDataPlane(
	ctx context.Context,
	dataplane *core_mesh.DataplaneResource,
	mesh *core_mesh.MeshResource,
	otherMeshes []*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, map[string]*core_xds.CaSecret, error) {
	identity, cas, _, err := s.get(ctx, mesh_proto.DataplaneProxyType, dataplane, dataplane.Spec.TagSet(), mesh, otherMeshes)
	return identity, cas, err
}

func (s *secrets) GetAllInOne(
	ctx context.Context,
	mesh *core_mesh.MeshResource,
	dataplane *core_mesh.DataplaneResource,
	otherMeshes []*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
	identity, _, allInOne, err := s.get(ctx, mesh_proto.DataplaneProxyType, dataplane, dataplane.Spec.TagSet(), mesh, otherMeshes)
	return identity, allInOne.CaSecret, err
}

func (s *secrets) GetForZoneEgress(
	ctx context.Context,
	zoneEgress *core_mesh.ZoneEgressResource,
	mesh *core_mesh.MeshResource,
) (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
	tags := mesh_proto.MultiValueTagSetFrom(map[string][]string{
		mesh_proto.ServiceTag: {
			mesh_proto.ZoneEgressServiceName,
		},
	})

	identity, cas, _, err := s.get(ctx, mesh_proto.EgressProxyType, zoneEgress, tags, mesh, nil)
	return identity, cas[mesh.GetMeta().GetName()], err
}

func (s *secrets) get(
	ctx context.Context,
	proxyType mesh_proto.ProxyType,
	resource model.Resource,
	tags mesh_proto.MultiValueTagSet,
	mesh *core_mesh.MeshResource,
	otherMeshes []*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, map[string]*core_xds.CaSecret, MeshCa, error) {
	if !mesh.MTLSEnabled() {
		return nil, nil, MeshCa{}, nil
	}

	meshName := mesh.GetMeta().GetName()

	resourceKey := model.MetaToResourceKey(resource.GetMeta())
	resourceKey.Mesh = meshName
	key := certCacheKey{
		resource:  resourceKey,
		proxyType: proxyType,
	}
	certs := s.certs(proxyType, resourceKey)

	if updateKinds, debugReason := s.shouldGenerateCerts(
		certs.Info(),
		tags,
		mesh,
		otherMeshes,
	); len(updateKinds) > 0 {
		// A previous generation failed and we're still within the backoff
		// window - return an error without calling the CA. We must NOT serve
		// stale certs here: a nil error lets the watchdog advance its hash
		// (dataplane_watchdog.go) and skip subsequent ticks, so the backoff
		// would never elapse and the DP would keep the old certs until the next
		// mesh change. Returning the error keeps the hash stale so ticks keep
		// re-checking the backoff (cheaply, without hitting the CA).
		if remaining := s.backoffRemaining(key); remaining > 0 {
			return nil, nil, MeshCa{}, errors.Errorf("backing off certificate generation for %s after a previous failure", remaining)
		}

		log.Info(
			"generating certificate",
			string(resource.Descriptor().Name), resourceKey, "reason", debugReason,
		)

		newCerts, err := s.generateCerts(ctx, tags, mesh, otherMeshes, certs, updateKinds)
		if err != nil {
			backoff := s.recordFailure(key)
			s.certGenerationFailuresMetric.WithLabelValues(meshName).Inc()
			log.Info(
				"could not generate certificates, backing off before retrying",
				string(resource.Descriptor().Name), resourceKey, "backoff", backoff, "error", err,
			)
			return nil, nil, MeshCa{}, errors.Wrap(err, "could not generate certificates")
		}
		s.clearBackoff(key)

		s.Lock()
		s.cachedCerts[key] = newCerts
		s.Unlock()

		return newCerts.identity, caMap(newCerts, meshName), newCerts.allInOneCa, nil
	}

	if certs == nil { // previous "if" should guarantee that the certs are always there
		return nil, nil, MeshCa{}, errors.New("certificates were not generated")
	}

	return certs.identity, caMap(certs, meshName), certs.allInOneCa, nil
}

func caMap(certs *certs, meshName string) map[string]*core_xds.CaSecret {
	result := map[string]*core_xds.CaSecret{
		meshName: certs.ownCa.CaSecret,
	}
	for _, otherCa := range certs.otherCas {
		result[otherCa.Mesh] = otherCa.CaSecret
	}
	return result
}

// backoffRemaining returns how long certificate generation for the given proxy
// is still being backed off after a previous failure, or 0 if it may proceed.
func (s *secrets) backoffRemaining(key certCacheKey) time.Duration {
	s.RLock()
	defer s.RUnlock()
	b, ok := s.backoffs[key]
	if !ok {
		return 0
	}
	if remaining := b.nextTry.Sub(core.Now()); remaining > 0 {
		return remaining
	}
	return 0
}

// recordFailure registers a generation failure and returns the delay before the
// next attempt is allowed. The delay grows exponentially up to a cap (see
// CertBackoff).
func (s *secrets) recordFailure(key certCacheKey) time.Duration {
	s.Lock()
	defer s.Unlock()
	b, ok := s.backoffs[key]
	if !ok {
		b = &certBackoff{backoff: s.newBackoff()}
		s.backoffs[key] = b
	}

	delay, _ := b.backoff.Next()
	b.nextTry = core.Now().Add(delay)
	return delay
}

func (s *secrets) clearBackoff(key certCacheKey) {
	s.Lock()
	defer s.Unlock()
	delete(s.backoffs, key)
}

func (s *secrets) Cleanup(proxyType mesh_proto.ProxyType, dpKey model.ResourceKey) {
	key := certCacheKey{
		resource:  dpKey,
		proxyType: proxyType,
	}
	s.Lock()
	delete(s.cachedCerts, key)
	delete(s.backoffs, key)
	s.Unlock()
}

func (s *secrets) shouldGenerateCerts(info *Info, tags mesh_proto.MultiValueTagSet, ownMesh *core_mesh.MeshResource, otherMeshInfos []*core_mesh.MeshResource) (UpdateKinds, string) {
	if info == nil {
		return UpdateEverything(), "mTLS is enabled and DP hasn't received a certificate yet"
	}

	var reason string
	updates := UpdateKinds{}

	if !proto.Equal(info.OwnMesh.MTLS, ownMesh.Spec.Mtls) {
		updates.AddKind(OwnMeshChange)
		reason = "Mesh mTLS settings have changed"
	}

	if len(info.OtherMeshInfos) != len(otherMeshInfos) || info.failedOtherMeshes {
		updates.AddKind(OtherMeshChange)
		reason = "Another mesh has been added or removed or we must retry"
	} else {
		for _, otherMesh := range otherMeshInfos {
			if previousInfo, found := info.OtherMeshInfos[otherMesh.GetMeta().GetName()]; found {
				if !proto.Equal(previousInfo.MTLS, otherMesh.Spec.Mtls) {
					updates.AddKind(OtherMeshChange)
					reason = "Another Mesh's mTLS settings have changed"
					break
				}
			} else {
				updates.AddKind(OtherMeshChange)
				reason = "Another Mesh's mTLS settings have changed"
				break
			}
		}
	}

	if tags.String() != info.Tags.String() {
		updates.AddKind(IdentityChange)
		reason = "DP tags have changed"
	}

	if info.ExpiringSoon() {
		updates.AddKind(IdentityChange)
		reason = fmt.Sprintf("the certificate expiring soon. Generated at %q, expiring at %q", info.Generation, info.Expiration)
	}

	return updates, reason
}

func (s *secrets) generateCerts(
	ctx context.Context,
	tags mesh_proto.MultiValueTagSet,
	mesh *core_mesh.MeshResource,
	otherMeshes []*core_mesh.MeshResource,
	oldCerts *certs,
	updateKinds UpdateKinds,
) (*certs, error) {
	ctx = user.Ctx(ctx, user.ControlPlane)
	var identity *core_xds.IdentitySecret
	var ownCa MeshCa
	var otherCas []MeshCa
	var allInOneCa MeshCa
	info := &Info{}

	if oldCerts != nil {
		identity = oldCerts.identity
		ownCa = oldCerts.ownCa
		otherCas = oldCerts.otherCas
		allInOneCa = oldCerts.allInOneCa
		info = oldCerts.Info()
	}

	meshName := mesh.GetMeta().GetName()

	if updateKinds.HasType(IdentityChange) || updateKinds.HasType(OwnMeshChange) {
		requester := Identity{
			Services: tags,
			Mesh:     meshName,
		}

		identitySecret, issuedBackend, err := s.identityProvider.Get(ctx, requester, mesh)
		if err != nil {
			return nil, errors.Wrap(err, "could not get Dataplane cert pair")
		}

		s.certGenerationsMetric.WithLabelValues(requester.Mesh).Inc()

		block, _ := pem.Decode(identitySecret.PemCerts[0])
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "could not extract info about certificate")
		}

		info.Tags = tags
		info.IssuedBackend = issuedBackend
		info.Expiration = cert.NotAfter
		info.Generation = core.Now()
		identity = identitySecret
	}

	if updateKinds.HasType(OwnMeshChange) {
		caSecret, supportedBackends, err := s.caProvider.Get(ctx, mesh)
		if err != nil {
			return nil, errors.Wrap(err, "could not get mesh CA cert")
		}

		ownCa = MeshCa{
			Mesh:     meshName,
			CaSecret: caSecret,
		}
		info.SupportedBackends = supportedBackends
		info.OwnMesh = MeshInfo{
			MTLS: mesh.Spec.Mtls,
		}
	}

	if updateKinds.HasType(OtherMeshChange) || updateKinds.HasType(OwnMeshChange) {
		otherMeshInfos := make(map[string]MeshInfo, len(otherMeshes))
		var bytes [][]byte
		var names []string
		otherCas = []MeshCa{}

		failedOtherMeshes := false
		for _, otherMesh := range otherMeshes {
			otherMeshInfos[otherMesh.GetMeta().GetName()] = MeshInfo{
				MTLS: otherMesh.Spec.GetMtls(),
			}

			// We need to track this mesh but we don't do anything with certs
			if !otherMesh.MTLSEnabled() {
				continue
			}

			otherCa, _, err := s.caProvider.Get(ctx, otherMesh)
			if err != nil {
				failedOtherMeshes = true
				// The other CA is misconfigured but this can not affect
				// generation in this mesh.
				log.Error(err, "could not get other mesh CA cert")
				continue
			}
			otherMeshName := otherMesh.GetMeta().GetName()

			otherCas = append(otherCas, MeshCa{
				Mesh:     otherMeshName,
				CaSecret: otherCa,
			})

			names = append(names, otherMeshName)
			bytes = append(bytes, otherCa.PemCerts...)
		}

		names = append(names, meshName)
		bytes = append(bytes, ownCa.CaSecret.PemCerts...)

		sort.Strings(names)
		allInOneCa = MeshCa{
			Mesh: strings.Join(names, ":"),
			CaSecret: &core_xds.CaSecret{
				PemCerts: bytes,
			},
		}

		info.failedOtherMeshes = failedOtherMeshes
		info.OtherMeshInfos = otherMeshInfos
	}

	return &certs{
		identity:   identity,
		ownCa:      ownCa,
		otherCas:   otherCas,
		allInOneCa: allInOneCa,
		info:       info,
	}, nil
}
