package multizone

import (
	"context"
	"encoding/json"

	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/deployments/zoneproxy"
)

// MeshEnv is the setup a multizone suite repeats: a Mesh, the MeshIdentity its
// proxies authenticate with, the mesh scoped zone proxies of every zone, and
// the apps the suite asserts on. Zone proxies are mesh scoped, so a suite can
// no longer lean on the shared zone proxies the environment installs - MeshEnv
// keeps that out of the suites.
type MeshEnv struct {
	Mesh      string
	Namespace string

	meshBuilder *builders.MeshBuilder
	identity    string
	global      []InstallFunc
	zones       []*meshEnvZone
	proxyOpts   []zoneproxy.DeploymentOptsFn
	noProxies   bool
}

type meshEnvZone struct {
	cluster  Cluster
	setup    []InstallFunc
	installs []InstallFunc
}

// NewMeshEnv returns an env for a mesh, using the mesh name as the Kubernetes
// namespace of its apps and zone proxies.
func NewMeshEnv(mesh string) *MeshEnv {
	return &MeshEnv{
		Mesh:        mesh,
		Namespace:   mesh,
		meshBuilder: builders.Mesh().WithName(mesh),
		identity:    mesh + "-identity",
		proxyOpts:   []zoneproxy.DeploymentOptsFn{zoneproxy.WithIngress()},
	}
}

// WithNamespace overrides the namespace, which otherwise matches the mesh name.
func (e *MeshEnv) WithNamespace(namespace string) *MeshEnv {
	e.Namespace = namespace
	return e
}

// WithMesh replaces the Mesh the env installs, for suites that need more than a
// plain mesh.
func (e *MeshEnv) WithMesh(mesh *builders.MeshBuilder) *MeshEnv {
	e.meshBuilder = mesh
	return e
}

// WithGlobal installs extra resources on Global, after the Mesh and before the
// zones are set up.
func (e *MeshEnv) WithGlobal(installs ...InstallFunc) *MeshEnv {
	e.global = append(e.global, installs...)
	return e
}

// WithZone adds a zone to the mesh together with the apps it runs. Zones can be
// added while the suite is building its Ginkgo tree or inside BeforeAll, so a
// cluster the suite creates itself is added the same way as a shared one.
func (e *MeshEnv) WithZone(cluster Cluster, installs ...InstallFunc) *MeshEnv {
	zone := e.zone(cluster)
	zone.installs = append(zone.installs, installs...)
	return e
}

// WithZoneSetup installs resources in a zone one by one, before its apps and
// zone proxies go up in parallel. It is where the namespaces the apps live in
// belong, and where a cluster the suite created gets its control plane.
func (e *MeshEnv) WithZoneSetup(cluster Cluster, installs ...InstallFunc) *MeshEnv {
	zone := e.zone(cluster)
	zone.setup = append(zone.setup, installs...)
	return e
}

// WithZoneEgress deploys an egress next to the ingress of every zone.
func (e *MeshEnv) WithZoneEgress() *MeshEnv {
	e.proxyOpts = append(e.proxyOpts, zoneproxy.WithEgress())
	return e
}

// WithZoneProxyOpts passes options to the zone proxies of every zone.
func (e *MeshEnv) WithZoneProxyOpts(opts ...zoneproxy.DeploymentOptsFn) *MeshEnv {
	e.proxyOpts = append(e.proxyOpts, opts...)
	return e
}

// WithoutZoneProxies leaves the zone proxies to the suite, for suites that
// assert on how a mesh behaves without them or deploy them themselves.
func (e *MeshEnv) WithoutZoneProxies() *MeshEnv {
	e.noProxies = true
	return e
}

// Setup installs the mesh on Global, the apps and zone proxies in every zone,
// and returns once cross-zone traffic can be routed.
func (e *MeshEnv) Setup() error {
	global := NewClusterSetup().
		Install(Yaml(e.meshBuilder)).
		Install(MeshIdentityBundled(e.Mesh, e.identity)).
		Install(MeshTrafficPermissionAllowAllUniversalWorkloadIdentity(e.Mesh, e.trustDomains()...))
	for _, install := range e.global {
		global = global.Install(install)
	}
	if err := global.Setup(Global); err != nil {
		return err
	}

	group := errgroup.Group{}
	for _, zone := range e.zones {
		setup := NewClusterSetup()
		if _, ok := zone.cluster.(*K8sCluster); ok {
			setup = setup.Install(NamespaceWithSidecarInjection(e.Namespace))
		}
		for _, install := range zone.setup {
			setup = setup.Install(install)
		}
		// Waiting per zone instead of up front lets a zone bring its control
		// plane up in its own setup, and keeps the apps from starting before
		// KDS delivered the mesh they belong to.
		setup.
			Install(waitForMeshInZone(e.Mesh)).
			Install(Parallel(e.zoneInstalls(zone)...)).
			SetupInGroup(zone.cluster, &group)
	}
	if err := group.Wait(); err != nil {
		return err
	}

	// Every zone mints its own CA from the MeshIdentity, so each one has to be
	// told about the others before cross-zone mTLS can be established.
	if err := DistributeMeshTrusts(Global, e.Mesh, e.identity, e.clusters()...); err != nil {
		return err
	}
	if e.noProxies {
		return nil
	}
	return e.waitForZoneProxies()
}

// RegisterHooks declares the Ginkgo nodes a suite would otherwise repeat: debug
// output on failure and teardown of everything Setup created. Call it while the
// tree is built, the env can still gain zones in BeforeAll.
func (e *MeshEnv) RegisterHooks() {
	AfterEachFailure(func() {
		e.Debug()
	})
	E2EAfterAll(func() {
		Expect(e.Cleanup()).To(Succeed())
	})
}

// Debug dumps the state of Global and of every zone of the mesh.
func (e *MeshEnv) Debug() {
	DebugUniversal(Global, e.Mesh)
	for _, zone := range e.zones {
		if _, ok := zone.cluster.(*K8sCluster); ok {
			DebugKube(zone.cluster, e.Mesh, e.Namespace)
			continue
		}
		DebugUniversal(zone.cluster, e.Mesh)
	}
}

// Cleanup deletes the apps and zone proxies of every zone and the Mesh itself.
// Clusters the suite created stay up, dismissing them is the suite's job.
func (e *MeshEnv) Cleanup() error {
	for _, zone := range e.zones {
		switch cluster := zone.cluster.(type) {
		case *K8sCluster:
			if err := cluster.TriggerDeleteNamespace(e.Namespace); err != nil {
				return err
			}
		case *UniversalCluster:
			if err := cluster.DeleteMeshApps(e.Mesh); err != nil {
				return err
			}
		}
	}
	return Global.DeleteMesh(e.Mesh)
}

func (e *MeshEnv) zone(cluster Cluster) *meshEnvZone {
	for _, zone := range e.zones {
		if zone.cluster == cluster {
			return zone
		}
	}
	zone := &meshEnvZone{cluster: cluster}
	e.zones = append(e.zones, zone)
	return zone
}

// waitForMeshInZone waits for the mesh in the zone it is installed on.
func waitForMeshInZone(mesh string) InstallFunc {
	return func(cluster Cluster) error {
		return WaitForMesh(mesh, []Cluster{cluster})
	}
}

func (e *MeshEnv) zoneInstalls(zone *meshEnvZone) []InstallFunc {
	installs := append([]InstallFunc{}, zone.installs...)
	if e.noProxies {
		return installs
	}
	opts := append([]zoneproxy.DeploymentOptsFn{
		zoneproxy.WithMesh(e.Mesh),
		zoneproxy.WithNamespace(e.Namespace),
	}, e.proxyOpts...)
	return append(installs, zoneproxy.Install(opts...))
}

func (e *MeshEnv) clusters() []Cluster {
	clusters := make([]Cluster, 0, len(e.zones))
	for _, zone := range e.zones {
		clusters = append(clusters, zone.cluster)
	}
	return clusters
}

func (e *MeshEnv) trustDomains() []string {
	trustDomains := make([]string, 0, len(e.zones))
	for _, zone := range e.zones {
		trustDomains = append(trustDomains, MeshIdentityTrustDomain(e.Mesh, zone.cluster))
	}
	return trustDomains
}

// waitForZoneProxies blocks until every zone published the address of its
// ingress to Global. A ready Deployment is not enough: the other zones only
// route to the ingress once KDS carried its MeshZoneAddress up.
func (e *MeshEnv) waitForZoneProxies() error {
	for _, zone := range e.zones {
		zoneName := zone.cluster.ZoneName()
		if _, err := retry.DoWithRetryContextE(
			Global.GetTesting(), context.Background(),
			"wait for MeshZoneAddress of zone "+zoneName,
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				out, err := Global.GetKumactlOptions().
					RunKumactlAndGetOutput("get", "meshzoneaddresses", "-m", e.Mesh, "-ojson")
				if err != nil {
					return "", err
				}
				var list struct {
					Items []struct {
						Labels map[string]string `json:"labels"`
					} `json:"items"`
				}
				if err := json.Unmarshal([]byte(out), &list); err != nil {
					return "", err
				}
				for _, item := range list.Items {
					if item.Labels[mesh_proto.ZoneTag] == zoneName {
						return "", nil
					}
				}
				return "", errors.Errorf("mesh %s has no MeshZoneAddress of zone %s", e.Mesh, zoneName)
			},
		); err != nil {
			return err
		}
	}
	return nil
}
