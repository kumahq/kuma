package testserver

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kumahq/kuma/test/framework"
)

type DeploymentOpts struct {
	Name                string
	Namespace           string
	Mesh                string
	ReachableServices   []string
	ReachableBackends   string
	WithStatefulSet     bool
	ServiceAccount      string
	echoArgs            []string
	args                []string
	probes              []probeParams
	Replicas            int32
	WaitingToBeReady    bool
	EnableProbes        bool
	EnableService       bool
	HeadlessService     bool
	PodAnnotations      map[string]string
	PodLabels           map[string]string
	NodeSelector        map[string]string
	protocol            string
	tlsKey              string
	tlsCrt              string
	initContainersToAdd []corev1.Container
}

func DefaultDeploymentOpts() DeploymentOpts {
	return DeploymentOpts{
		Mesh:             "default",
		Name:             "test-server",
		Namespace:        framework.TestNamespace,
		Replicas:         1,
		WaitingToBeReady: true,
		PodAnnotations:   map[string]string{},
		EnableProbes:     true,
		EnableService:    true,
		protocol:         "http",
	}
}

type DeploymentOptsFn = func(*DeploymentOpts)

func WithMesh(mesh string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Mesh = mesh
	}
}

func WithName(name string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Name = name
	}
}

func WithTLS(key, crt string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.tlsKey = key
		opts.tlsCrt = crt
	}
}

func WithReachableServices(services ...string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.ReachableServices = services
	}
}

func WithReachableBackends(config string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.ReachableBackends = config
	}
}

func WithNamespace(namespace string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Namespace = namespace
	}
}

func WithReplicas(n int32) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Replicas = n
	}
}

func WithStatefulSet() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.WithStatefulSet = true
	}
}

func WithServiceAccount(serviceAccountName string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.ServiceAccount = serviceAccountName
	}
}

func WithoutWaitingToBeReady() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.WaitingToBeReady = false
	}
}

// WithEchoArgs sets the arguments for the echo server, values will be appended the default echo arguments
func WithEchoArgs(args ...string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.echoArgs = args
	}
}

// WithArgs sets the arguments for the test server, they take precedence over the echo arguments
func WithArgs(args ...string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.args = args
	}
}

type ProbeType string

const (
	ReadinessProbe ProbeType        = "readiness"
	StartupProbe   ProbeType        = "startup"
	LivenessProbe  ProbeType        = "liveness"
	ProbeHttpGet   ProbeHandlerType = "httpGet"
	ProbeTcpSocket ProbeHandlerType = "tcpSocket"
	ProbeGRPC      ProbeHandlerType = "grpc"
)

type (
	ProbeHandlerType string
	probeParams      struct {
		ProbeType   ProbeType
		HandlerType ProbeHandlerType
		Port        uint32
		HttpGetPath string
	}
)

func (p probeParams) toKubeProbe() *corev1.Probe {
	switch p.HandlerType {
	case ProbeHttpGet:
		return &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: p.HttpGetPath,
					Port: intstr.FromInt32(int32(p.Port)),
				},
			},
			InitialDelaySeconds: 3,
			PeriodSeconds:       5,
			TimeoutSeconds:      3,
			FailureThreshold:    60,
		}
	case ProbeTcpSocket:
		return &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt32(int32(p.Port)),
				},
			},
			InitialDelaySeconds: 3,
			PeriodSeconds:       3,
		}
	case ProbeGRPC:
		return &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				GRPC: &corev1.GRPCAction{
					Port: int32(p.Port),
				},
			},
			InitialDelaySeconds: 3,
			PeriodSeconds:       3,
		}
	}
	return nil
}

// WithProbe adds a probe to the deployment, this only works when the arguments are customize using WithArgs
func WithProbe(probeType ProbeType, handlerType ProbeHandlerType, port uint32, httpGetPath string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		if opts.probes == nil {
			opts.probes = []probeParams{}
		}
		opts.probes = append(opts.probes, probeParams{
			ProbeType:   probeType,
			HandlerType: handlerType,
			Port:        port,
			HttpGetPath: httpGetPath,
		})
	}
}

func WithPodAnnotations(annotations map[string]string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.PodAnnotations = annotations
	}
}

func WithPodLabels(labels map[string]string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.PodLabels = labels
	}
}

func WithoutProbes() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.EnableProbes = false
	}
}

func WithoutService() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.EnableService = false
	}
}

func WithHeadlessService() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.HeadlessService = true
	}
}

func WithNodeSelector(selector map[string]string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.NodeSelector = selector
	}
}

func WithServicePortAppProtocol(protocol string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.protocol = protocol
	}
}

func AddInitContainer(initContainer corev1.Container) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.initContainersToAdd = append(opts.initContainersToAdd, initContainer)
	}
}

type TestServer interface{}

type Deployment interface {
	framework.Deployment
	TestServer
}

func Install(fn ...DeploymentOptsFn) framework.InstallFunc {
	opts := DefaultDeploymentOpts()
	for _, f := range fn {
		f(&opts)
	}

	return func(cluster framework.Cluster) error {
		var deployment Deployment
		switch cluster.(type) {
		case *framework.K8sCluster:
			deployment = &k8SDeployment{
				opts: opts,
			}
		default:
			return errors.New("invalid cluster")
		}
		return cluster.Deploy(deployment)
	}
}
