package framework

import (
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/kumactl"
)

type InstallationMode string

var (
	HelmInstallationMode    = InstallationMode("helm")
	KumactlInstallationMode = InstallationMode("kumactl")
)

// KumaDeploymentOptions are the options used for installing the Kuma
// control plane.
type KumaDeploymentOptions struct {
	// common
	isipv6  bool
	verbose *bool

	// cp specific
	ctlOpts                     map[string]string
	globalAddress               string
	installationMode            InstallationMode
	skipDefaultMesh             bool
	helmReleaseName             string
	helmChartPath               *string
	helmChartVersion            string
	helmOpts                    map[string]string
	noHelmOpts                  []string
	env                         map[string]string
	zoneIngress                 bool
	zoneIngressEnvoyAdminTunnel bool
	zoneEgress                  bool
	zoneEgressEnvoyAdminTunnel  bool
	cni                         bool
	cniNamespace                string
	cpReplicas                  int
	hdsDisabled                 bool
	runPostgresMigration        bool
	yamlConfig                  string
	apiHeaders                  []string
	zoneName                    string
	verifyKuma                  bool
	setupKumactl                bool

	// Functions to apply to each mesh after the control plane
	// is provisioned.
	meshUpdateFuncs map[string][]func(*mesh_proto.Mesh) *mesh_proto.Mesh
}

func (k *KumaDeploymentOptions) apply(opts ...KumaDeploymentOption) {
	// Set defaults.
	k.isipv6 = Config.IPV6
	k.installationMode = KumactlInstallationMode
	k.env = map[string]string{}
	k.meshUpdateFuncs = map[string][]func(*mesh_proto.Mesh) *mesh_proto.Mesh{}
	k.verifyKuma = true
	k.setupKumactl = true

	// Apply options.
	for _, o := range opts {
		o.ApplyKuma(k)
	}
}

type KumaDeploymentOption interface {
	ApplyKuma(options *KumaDeploymentOptions)
}

type KumaOptionFunc func(options *KumaDeploymentOptions)

func (f KumaOptionFunc) ApplyKuma(opts *KumaDeploymentOptions) {
	if f != nil {
		f(opts)
	}
}

type AppDeploymentOptions struct {
	// common
	isipv6  bool
	verbose *bool

	// app specific
	namespace             string
	appname               string
	name                  string
	appYaml               string
	appArgs               []string
	appLabel              string
	token                 string
	transparent           *bool
	builtindns            *bool // true by default
	protocol              string
	serviceName           string
	serviceVersion        string
	serviceInstance       string
	mesh                  string
	dpVersion             string
	kumactlFlow           bool
	concurrency           int
	omitDataplane         bool
	proxyOnly             bool
	serviceProbe          bool
	reachableServices     []string
	appendDataplaneConfig string
	boundToContainerIp    bool
	serviceAddress        string
	dpEnvs                map[string]string
	additionalTags        map[string]string

	dockerVolumes       []string
	dockerContainerName string
}

func (d *AppDeploymentOptions) apply(opts ...AppDeploymentOption) {
	// Set defaults.
	d.isipv6 = Config.IPV6

	// Apply options.
	for _, o := range opts {
		o.ApplyApp(d)
	}
}

type AppDeploymentOption interface {
	ApplyApp(*AppDeploymentOptions)
}

type AppOptionFunc func(*AppDeploymentOptions)

func (f AppOptionFunc) ApplyApp(opts *AppDeploymentOptions) {
	if f != nil {
		f(opts)
	}
}

type GenericOption struct {
	Kuma KumaOptionFunc
	App  AppOptionFunc
}

var (
	_ KumaDeploymentOption = &GenericOption{}
	_ AppDeploymentOption  = &GenericOption{}
)

func (o *GenericOption) ApplyApp(opts *AppDeploymentOptions) {
	o.App(opts)
}

func (o *GenericOption) ApplyKuma(opts *KumaDeploymentOptions) {
	o.Kuma(opts)
}

// WithVerbose enables verbose mode on the deployment.
func WithVerbose() *GenericOption {
	enabled := true

	return &GenericOption{
		App: func(o *AppDeploymentOptions) {
			o.verbose = &enabled
		},
		Kuma: func(o *KumaDeploymentOptions) {
			o.verbose = &enabled
		},
	}
}

func WithIPv6(enabled bool) *GenericOption {
	return &GenericOption{
		Kuma: func(o *KumaDeploymentOptions) {
			o.isipv6 = enabled
		},
		App: func(o *AppDeploymentOptions) {
			o.isipv6 = enabled
		},
	}
}

func WithPostgres(envVars map[string]string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.runPostgresMigration = true

		if o.env == nil {
			o.env = map[string]string{}
		}

		for key, value := range envVars {
			o.env[key] = value
		}
	})
}

func WithHDS(enabled bool) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.hdsDisabled = !enabled
	})
}

// WithCPReplicas works only with HELM for now.
func WithCPReplicas(cpReplicas int) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.cpReplicas = cpReplicas
	})
}

// WithSkipDefaultMesh works only with HELM now.
func WithSkipDefaultMesh(skip bool) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.skipDefaultMesh = skip
	})
}

func WithInstallationMode(mode InstallationMode) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.installationMode = mode
	})
}

func WithHelmReleaseName(name string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.helmReleaseName = name
	})
}

func WithHelmChartPath(path string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.helmChartPath = &path
	})
}

func WithHelmChartVersion(version string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.helmChartVersion = version
	})
}

func WithHelmOpt(name, value string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		if o.helmOpts == nil {
			o.helmOpts = map[string]string{}
		}
		o.helmOpts[name] = value
	})
}

func WithoutHelmOpt(name string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.noHelmOpts = append(o.noHelmOpts, name)
	})
}

func ClearNoHelmOpts() KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.noHelmOpts = nil
	})
}

func WithEnv(name, value string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.env[name] = value
	})
}

func WithEnvs(entries map[string]string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		for k, v := range entries {
			o.env[k] = v
		}
	})
}

func WithYamlConfig(cfg string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.yamlConfig = cfg
	})
}

func WithIngressEnvoyAdminTunnel() KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.zoneIngressEnvoyAdminTunnel = true
	})
}

func WithIngress() KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.zoneIngress = true
	})
}

func WithEgressEnvoyAdminTunnel() KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.zoneEgressEnvoyAdminTunnel = true
	})
}

type CNIVersion string

const (
	CNIVersion1 CNIVersion = "v1"
	CNIVersion2 CNIVersion = "v2"
)

func WithEgress() KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.zoneEgress = true
	})
}

func WithCNI() KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.cni = true
	})
}

func WithCNINamespace(namespace string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.cniNamespace = namespace
	})
}

func WithGlobalAddress(address string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.globalAddress = address
	})
}

// WithCtlOpts allows arbitrary options to be passed to kuma, which is important
// for using test/framework in other libraries where additional options may have
// been added.
func WithCtlOpts(opts map[string]string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		if o.ctlOpts == nil {
			o.ctlOpts = map[string]string{}
		}
		for name, value := range opts {
			o.ctlOpts[name] = value
		}
	})
}

type MeshUpdateFunc func(mesh *mesh_proto.Mesh) *mesh_proto.Mesh

// WithMeshUpdate registers a function to update the specification
// for the named mesh. When the control plane implementation creates the
// mesh, it invokes the function and applies configuration changes to the
// mesh object.
func WithMeshUpdate(mesh string, u MeshUpdateFunc) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.meshUpdateFuncs[mesh] = append(o.meshUpdateFuncs[mesh], u)
	})
}

func WithApiHeaders(headers ...string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.apiHeaders = headers
	})
}

func WithZoneName(zoneName string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.zoneName = zoneName
	})
}

func WithoutVerifyingKuma() KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.verifyKuma = false
	})
}

func WithoutConfiguringKumactl() KumaDeploymentOption {
	return KumaOptionFunc(func(o *KumaDeploymentOptions) {
		o.setupKumactl = false
	})
}

// WithoutDataplane suppresses the automatic configuration of kuma-dp
// in the application container. This is useful when the test requires a
// container that is not bound to the mesh.
func WithoutDataplane() AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.omitDataplane = true
	})
}

// WithKumactlFlow creates the Dataplane resource in the control plane before the actual data plane proxy start.
// If proxy is disconnected, resource won't be deleted from the storage and will be displayed as Offline.
func WithKumactlFlow() AppDeploymentOption {
	return AppOptionFunc(func(options *AppDeploymentOptions) {
		options.kumactlFlow = true
	})
}

func WithYaml(appYaml string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.appYaml = appYaml
	})
}

func WithProtocol(protocol string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.protocol = protocol
	})
}

func WithArgs(appArgs []string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.appArgs = appArgs
	})
}

func WithServiceName(name string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.serviceName = name
	})
}

func WithAppLabel(app string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.appLabel = app
	})
}

func WithAppendDataplaneYaml(config string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.appendDataplaneConfig = config
	})
}

func WithServiceVersion(version string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.serviceVersion = version
	})
}

func WithServiceInstance(instance string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.serviceInstance = instance
	})
}

func ProxyOnly() AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.proxyOnly = true
	})
}

func ServiceProbe() AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.serviceProbe = true
	})
}

func WithServiceAddress(serviceAddress string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.serviceAddress = serviceAddress
	})
}

// WithDPVersion only works with Universal now
func WithDPVersion(version string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.dpVersion = version
	})
}

func WithNamespace(namespace string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.namespace = namespace
	})
}

func WithMesh(mesh string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.mesh = mesh
	})
}

// BoundToContainerIp only works with Universal
func BoundToContainerIp() AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.boundToContainerIp = true
	})
}

func WithAppname(appname string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.appname = appname
	})
}

func WithName(name string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.name = name
	})
}

func WithToken(token string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.token = token
	})
}

func WithTransparentProxy(transparent bool) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.transparent = &transparent
	})
}

func WithAdditionalTags(tags map[string]string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.additionalTags = tags
	})
}

func WithDpEnvs(envs map[string]string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.dpEnvs = envs
	})
}

func WithBuiltinDNS(builtindns bool) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.builtindns = &builtindns
	})
}

func WithConcurrency(concurrency int) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.concurrency = concurrency
	})
}

func WithReachableServices(services ...string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.reachableServices = services
	})
}

func WithDockerVolumes(volumes ...string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.dockerVolumes = append(o.dockerVolumes, volumes...)
	})
}

func WithDockerContainerName(name string) AppDeploymentOption {
	return AppOptionFunc(func(o *AppDeploymentOptions) {
		o.dockerContainerName = name
	})
}

type NamespaceDeleteHookFunc func(c Cluster, namespace string) error

type Deployment interface {
	Name() string
	Deploy(cluster Cluster) error
	Delete(cluster Cluster) error
}

type Cluster interface {
	// Cluster
	Name() string
	DismissCluster() error
	// Generic
	DeployKuma(mode core.CpMode, opts ...KumaDeploymentOption) error
	GetKuma() ControlPlane
	// GetKumaCPLogs returns a set of logs (depending on env it can be stdout, stderr or current and previous)...
	GetKumaCPLogs() map[string]string
	VerifyKuma() error
	DeleteKuma() error
	GetKumactlOptions() *kumactl.KumactlOptions
	Deployment(name string) Deployment
	Deploy(deployment Deployment) error
	DeleteDeployment(name string) error
	WithTimeout(timeout time.Duration) Cluster
	WithRetries(retries int) Cluster
	GetZoneEgressEnvoyTunnel() envoy_admin.Tunnel
	GetZoneIngressEnvoyTunnel() envoy_admin.Tunnel
	Verbose() bool
	Install(fn InstallFunc) error
	ZoneName() string

	// K8s
	GetKubectlOptions(namespace ...string) *k8s.KubectlOptions
	CreateNamespace(namespace string) error
	DeleteNamespace(namespace string, fns ...NamespaceDeleteHookFunc) error
	DeployApp(fs ...AppDeploymentOption) error
	Exec(namespace, podName, containerName string, cmd ...string) (string, string, error)

	// Testing
	GetTesting() testing.TestingT
}

type ControlPlane interface {
	GetName() string
	GetMetrics() (string, error)
	GetMonitoringAssignment(clientId string) (string, error)
	GetKDSServerAddress() string
	GetKDSInsecureServerAddress() string
	GetXDSServerAddress() string
	GetGlobalStatusAPI() string
	GetAPIServerAddress() string
	GenerateDpToken(mesh, serviceName string) (string, error)
	GenerateZoneIngressToken(zone string) (string, error)
	GenerateZoneEgressToken(zone string) (string, error)
	GenerateZoneToken(zone string, scope []string) (string, error)
	Exec(cmd ...string) (string, string, error)
}
