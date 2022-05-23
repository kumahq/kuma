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
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/metrics"
)

var log = core.Log.WithName("xds").WithName("secrets")

type MeshCa struct {
	Mesh     string
	CaSecret *core_xds.CaSecret
}

type Secrets interface {
	GetForDataPlane(dataplane *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource, otherMeshes []*core_mesh.MeshResource) (*core_xds.IdentitySecret, []MeshCa, error)
	GetForGatewayListener(mesh *core_mesh.MeshResource, dataplane *core_mesh.DataplaneResource, otherMeshes []*core_mesh.MeshResource) (*core_xds.IdentitySecret, MeshCa, error)
	GetForZoneEgress(zoneEgress *core_mesh.ZoneEgressResource, mesh *core_mesh.MeshResource) (*core_xds.IdentitySecret, []MeshCa, error)
	Info(dpKey model.ResourceKey) *Info
	Cleanup(dpKey model.ResourceKey)
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

	MeshInfos []MeshInfo
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

func (s *secrets) Info(dpKey model.ResourceKey) *Info {
	certs := s.certs(dpKey)
	if certs == nil {
		return nil
	}
	return certs.info
}

type certs struct {
	identity   *core_xds.IdentitySecret
	cas        []MeshCa
	allInOneCa MeshCa
	info       *Info
}

func (c *certs) Info() *Info {
	if c == nil {
		return nil
	}
	return c.info
}

func (s *secrets) certs(dpKey model.ResourceKey) *certs {
	s.RLock()
	defer s.RUnlock()

	return s.cachedCerts[dpKey]
}

func (s *secrets) GetForDataPlane(
	dataplane *core_mesh.DataplaneResource,
	mesh *core_mesh.MeshResource,
	otherMeshes []*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, []MeshCa, error) {
	identity, cas, _, err := s.get(dataplane, dataplane.Spec.TagSet(), mesh, otherMeshes)
	return identity, cas, err
}

func (s *secrets) GetForGatewayListener(
	mesh *core_mesh.MeshResource,
	dataplane *core_mesh.DataplaneResource,
	otherMeshes []*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, MeshCa, error) {
	identity, _, allInOne, err := s.get(dataplane, dataplane.Spec.TagSet(), mesh, otherMeshes)
	return identity, allInOne, err
}

func (s *secrets) GetForZoneEgress(
	zoneEgress *core_mesh.ZoneEgressResource,
	mesh *core_mesh.MeshResource,
) (*core_xds.IdentitySecret, []MeshCa, error) {
	tags := mesh_proto.MultiValueTagSetFrom(map[string][]string{
		mesh_proto.ServiceTag: {
			mesh_proto.ZoneEgressServiceName,
		},
	})

	identity, cas, _, err := s.get(zoneEgress, tags, mesh, nil)
	return identity, cas, err
}

func (s *secrets) get(
	resource model.Resource,
	tags mesh_proto.MultiValueTagSet,
	mesh *core_mesh.MeshResource,
	otherMeshes []*core_mesh.MeshResource,
) (*core_xds.IdentitySecret, []MeshCa, MeshCa, error) {
	if !mesh.MTLSEnabled() {
		return nil, nil, MeshCa{}, nil
	}

	meshName := mesh.GetMeta().GetName()

	resourceKey := model.MetaToResourceKey(resource.GetMeta())
	resourceKey.Mesh = meshName
	certs := s.certs(resourceKey)

	if shouldGenerate, reason := s.shouldGenerateCerts(
		certs.Info(),
		append([]*core_mesh.MeshResource{mesh}, otherMeshes...),
		tags,
	); shouldGenerate {
		log.Info(
			"generating certificate",
			string(resource.Descriptor().Name), resourceKey, "reason", reason,
		)

		certs, err := s.generateCerts(meshName, tags, mesh, otherMeshes)
		if err != nil {
			return nil, nil, MeshCa{}, errors.Wrap(err, "could not generate certificates")
		}

		s.Lock()
		s.cachedCerts[resourceKey] = certs
		s.Unlock()

		return certs.identity, certs.cas, certs.allInOneCa, nil
	}

	if certs == nil { // previous "if" should guarantee that the certs are always there
		return nil, nil, MeshCa{}, errors.New("certificates were not generated")
	}

	return certs.identity, certs.cas, certs.allInOneCa, nil
}

func (s *secrets) Cleanup(dpKey model.ResourceKey) {
	s.Lock()
	delete(s.cachedCerts, dpKey)
	s.Unlock()
}

func (s *secrets) shouldGenerateCerts(info *Info, meshInfos []*core_mesh.MeshResource, tags mesh_proto.MultiValueTagSet) (bool, string) {
	if info == nil {
		return true, "mTLS is enabled and DP hasn't received a certificate yet"
	}

	if len(info.MeshInfos) != len(meshInfos) {
		return true, "Mesh mTLS settings have changed"
	}

	for i, mesh := range info.MeshInfos {
		if !proto.Equal(mesh.MTLS, meshInfos[i].Spec.Mtls) {
			return true, "Mesh mTLS settings have changed"
		}
	}

	if tags.String() != info.Tags.String() {
		return true, "DP tags have changed"
	}

	if info.ExpiringSoon() {
		return true, fmt.Sprintf("the certificate expiring soon. Generated at %q, expiring at %q", info.Generation, info.Expiration)
	}

	return false, ""
}

func (s *secrets) generateCerts(
	resourceMesh string,
	tags mesh_proto.MultiValueTagSet,
	mesh *core_mesh.MeshResource,
	otherMeshes []*core_mesh.MeshResource,
) (*certs, error) {
	requester := Identity{
		Services: tags,
		Mesh:     resourceMesh,
	}

	identity, issuedBackend, err := s.identityProvider.Get(context.Background(), requester, mesh)
	if err != nil {
		return nil, errors.Wrap(err, "could not get Dataplane cert pair")
	}

	s.certGenerationsMetric.WithLabelValues(requester.Mesh).Inc()

	ca, supportedBackends, err := s.caProvider.Get(context.Background(), mesh)
	if err != nil {
		return nil, errors.Wrap(err, "could not get mesh CA cert")
	}

	meshInfos := []MeshInfo{{
		MTLS: mesh.Spec.Mtls,
	}}

	cas := []MeshCa{{
		Mesh:     mesh.GetMeta().GetName(),
		CaSecret: ca,
	}}
	var bytes [][]byte
	var names []string

	for _, otherMesh := range otherMeshes {
		otherCa, _, err := s.caProvider.Get(context.Background(), otherMesh)
		if err != nil {
			return nil, errors.Wrap(err, "could not get other mesh CA cert")
		}

		meshInfos = append(meshInfos, MeshInfo{
			MTLS: otherMesh.Spec.Mtls,
		})

		meshName := otherMesh.GetMeta().GetName()

		cas = append(cas, MeshCa{
			Mesh:     meshName,
			CaSecret: otherCa,
		})

		names = append(names, meshName)
		bytes = append(bytes, otherCa.PemCerts...)
	}

	sort.Strings(names)
	allInOneCa := MeshCa{
		Mesh: strings.Join(names, ":"),
		CaSecret: &core_xds.CaSecret{
			PemCerts: bytes,
		},
	}

	info, err := newCertInfo(identity, tags, issuedBackend, supportedBackends, meshInfos)
	if err != nil {
		return nil, errors.Wrap(err, "could not extract info about certificate")
	}

	return &certs{
		identity:   identity,
		cas:        cas,
		allInOneCa: allInOneCa,
		info:       info,
	}, nil
}

func newCertInfo(identityCert *core_xds.IdentitySecret, tags mesh_proto.MultiValueTagSet, issuedBackend string, supportedBackends []string, meshInfos []MeshInfo) (*Info, error) {
	block, _ := pem.Decode(identityCert.PemCerts[0])
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	certInfo := &Info{
		Tags:              tags,
		Expiration:        cert.NotAfter,
		Generation:        core.Now(),
		IssuedBackend:     issuedBackend,
		SupportedBackends: supportedBackends,
		MeshInfos:         meshInfos,
	}
	return certInfo, nil
}
