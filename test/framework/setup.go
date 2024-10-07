package framework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_rest "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	bootstrap_k8s "github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s"
	"github.com/kumahq/kuma/pkg/tls"
)

type InstallFunc func(cluster Cluster) error

var Serializer *k8sjson.Serializer

func init() {
	K8sScheme, err := bootstrap_k8s.NewScheme()
	if err != nil {
		panic(err)
	}

	Serializer = k8sjson.NewSerializerWithOptions(
		k8sjson.DefaultMetaFactory, K8sScheme, K8sScheme,
		k8sjson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)
}

func YamlK8sObject(obj runtime.Object) InstallFunc {
	return func(cluster Cluster) error {
		b := bytes.Buffer{}
		err := Serializer.Encode(obj, &b)
		if err != nil {
			return err
		}

		// Remove "status" from Kubernetes YAMLs
		// "status" is an element in Kubernetes Object that is filled by Kubernetes, not a user.
		// Encoder by default also serializes the Status object with default values (since those are not pointers)
		// that does not have omitempty, so for example in case of StatefulSet.Status it will be
		// status:
		//   replicas: 0
		//   availableReplicas: 0
		// However, availableReplicas is a beta field that is not available in previous version of Kubernetes.
		obj := map[string]interface{}{}
		if err := yaml.Unmarshal(b.Bytes(), &obj); err != nil {
			return err
		}
		delete(obj, "status")
		bytes, err := yaml.Marshal(obj)
		if err != nil {
			return err
		}

		_, err = retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", DefaultRetries, DefaultTimeout,
			func() (string, error) {
				return "", k8s.KubectlApplyFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), string(bytes))
			})
		return err
	}
}

func YamlK8s(yamls ...string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", DefaultRetries, DefaultTimeout,
			func() (string, error) {
				for _, yaml := range yamls {
					if err := k8s.KubectlApplyFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), yaml); err != nil {
						return "", err
					}
				}
				return "", nil
			})
		return err
	}
}

func DeleteYamlK8s(yamls ...string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "delete yaml resource", DefaultRetries, DefaultTimeout,
			func() (string, error) {
				for _, yaml := range yamls {
					if err := k8s.KubectlDeleteFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), yaml); err != nil {
						return "", err
					}
				}
				return "", nil
			})
		return err
	}
}

func MeshUniversal(name string) InstallFunc {
	mesh := fmt.Sprintf(`
type: Mesh
name: %s
`, name)
	return YamlUniversal(mesh)
}

func MeshKubernetes(name string) InstallFunc {
	mesh := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
`, name)
	return YamlK8s(mesh)
}

func MeshWithMeshServicesKubernetes(name string, meshServicesEnabled string) InstallFunc {
	mesh := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  meshServices:
    mode: %s
`, name, meshServicesEnabled)
	return YamlK8s(mesh)
}

func MTLSMeshKubernetes(name string) InstallFunc {
	mesh := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  mtls:
    enabledBackend: ca-1
    backends:
      - name: ca-1
        type: builtin
`, name)
	return YamlK8s(mesh)
}

func MTLSMeshKubernetesWithEgressRouting(name string) InstallFunc {
	mesh := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  routing:
    zoneEgress: true
  mtls:
    enabledBackend: ca-1
    backends:
      - name: ca-1
        type: builtin
`, name)
	return YamlK8s(mesh)
}

func MTLSMeshWithMeshServicesKubernetes(name string, meshServicesMode string) InstallFunc {
	mesh := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  meshServices:
    mode: %s
  mtls:
    enabledBackend: ca-1
    backends:
      - name: ca-1
        type: builtin
`, name, meshServicesMode)
	return YamlK8s(mesh)
}

func MeshTrafficPermissionAllowAllKubernetes(name string) InstallFunc {
	mtp := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  namespace: %[2]s
  name: allow-all-%[1]s.%[2]s
  labels:
    kuma.io/mesh: %[1]s
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        action: Allow`, name, Config.KumaNamespace)
	return YamlK8s(mtp)
}

func MeshWithMeshServicesUniversal(name string, meshServicesMode string) InstallFunc {
	mesh := fmt.Sprintf(`
type: Mesh
name: %s
meshServices:
  mode: %s
`, name, meshServicesMode)
	return YamlUniversal(mesh)
}

func MTLSMeshUniversal(name string) InstallFunc {
	mesh := fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
    - name: ca-1
      type: builtin
`, name)
	return YamlUniversal(mesh)
}

func MTLSMeshWithMeshServicesUniversal(name string, meshServicesEnabled string) InstallFunc {
	mesh := fmt.Sprintf(`
type: Mesh
name: %s
meshServices:
  mode: %s
mtls:
  enabledBackend: ca-1
  backends:
    - name: ca-1
      type: builtin
`, name, meshServicesEnabled)
	return YamlUniversal(mesh)
}

func TrafficRouteKubernetes(name string) InstallFunc {
	tr := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: TrafficRoute
mesh: %[1]s
metadata:
  name: route-all-%[1]s
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: '*'
  conf:
    loadBalancer:
      roundRobin: {}
    destination:
      kuma.io/service: '*'`, name)
	return YamlK8s(tr)
}

func TrafficRouteUniversal(name string) InstallFunc {
	tr := fmt.Sprintf(`
type: TrafficRoute
name: route-all-%[1]s
mesh: %[1]s
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'
conf:
  loadBalancer:
    roundRobin: {}
  destination:
    kuma.io/service: '*'`, name)
	return YamlUniversal(tr)
}

func TrafficPermissionUniversal(name string) InstallFunc {
	tp := fmt.Sprintf(`
type: TrafficPermission
name: allow-all-%[1]s
mesh: %[1]s
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'`, name)
	return YamlUniversal(tp)
}

func TrafficPermissionKubernetes(name string) InstallFunc {
	tp := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: %[1]s
metadata:
  name: allow-all-%[1]s
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: '*'`, name)
	return YamlK8s(tp)
}

func TimeoutUniversal(name string) InstallFunc {
	timeout := fmt.Sprintf(`
type: Timeout
mesh: %[1]s
name: timeout-all-%[1]s
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'
conf:
  connectTimeout: 5s # all protocols
  tcp: # tcp, kafka
    idleTimeout: 1h
  http: # http, http2, grpc
    requestTimeout: 15s
    idleTimeout: 1h
    streamIdleTimeout: 30m
    maxStreamDuration: 0s`, name)
	return YamlUniversal(timeout)
}

func TimeoutKubernetes(name string) InstallFunc {
	timeout := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Timeout
mesh: %[1]s
metadata:
  name: timeout-all-%[1]s
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: '*'
  conf:
    connectTimeout: 5s # all protocols
    tcp: # tcp, kafka
      idleTimeout: 1h 
    http: # http, http2, grpc
      requestTimeout: 15s 
      idleTimeout: 1h
      streamIdleTimeout: 30m
      maxStreamDuration: 0s
`, name)
	return YamlK8s(timeout)
}

func CircuitBreakerUniversal(name string) InstallFunc {
	cb := fmt.Sprintf(`
type: CircuitBreaker
mesh: %[1]s
name: circuit-breaker-all-%[1]s
sources:
- match:
    kuma.io/service: '*'
destinations:
- match:
    kuma.io/service: '*'
conf:
  thresholds:
    maxConnections: 1024
    maxPendingRequests: 1024
    maxRequests: 1024
    maxRetries: 3`, name)
	return YamlUniversal(cb)
}

func CircuitBreakerKubernetes(name string) InstallFunc {
	cb := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: CircuitBreaker
mesh: %[1]s
metadata:
  name: circuit-breaker-all-%[1]s
spec:
  sources:
  - match:
      kuma.io/service: '*'
  destinations:
  - match:
      kuma.io/service: '*'
  conf:
    thresholds:
      maxConnections: 1024
      maxPendingRequests: 1024
      maxRequests: 1024
      maxRetries: 3`, name)
	return YamlK8s(cb)
}

func RetryUniversal(name string) InstallFunc {
	retry := fmt.Sprintf(`
type: Retry
name: retry-all-%[1]s
mesh: %[1]s
sources:
- match:
    kuma.io/service: '*'
destinations:
- match:
    kuma.io/service: '*'
conf:
  http:
    numRetries: 5
    perTryTimeout: 16s
    backOff:
      baseInterval: 25ms
      maxInterval: 250s
  grpc:
    numRetries: 5
    perTryTimeout: 16s
    backOff:
      baseInterval: 25ms
      maxInterval: 250ms
  tcp:
    maxConnectAttempts: 5
`, name)
	return YamlUniversal(retry)
}

func RetryKubernetes(name string) InstallFunc {
	retry := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Retry
mesh: %[1]s
metadata:
  name: retry-all-%[1]s
spec:
  sources:
  - match:
      kuma.io/service: '*'
  destinations:
  - match:
      kuma.io/service: '*'
  conf:
    http:
      numRetries: 5
      perTryTimeout: 16s
      backOff:
        baseInterval: 25ms
        maxInterval: 250s
    grpc:
      numRetries: 5
      perTryTimeout: 16s
      backOff:
        baseInterval: 25ms
        maxInterval: 250ms
    tcp:
      maxConnectAttempts: 5
`, name)
	return YamlK8s(retry)
}

func MeshTrafficPermissionAllowAllUniversal(name string) InstallFunc {
	mtp := fmt.Sprintf(`
type: MeshTrafficPermission
name: allow-all-%[1]s
mesh: %[1]s
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        action: Allow`, name)
	return YamlUniversal(mtp)
}

func YamlUniversal(yaml string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", DefaultRetries, DefaultTimeout,
			func() (string, error) {
				kumactl := cluster.GetKumactlOptions()
				return "", kumactl.KumactlApplyFromString(yaml)
			})
		return err
	}
}

func ResourceUniversal(resource model.Resource) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install resource", DefaultRetries, DefaultTimeout,
			func() (string, error) {
				kumactl := cluster.GetKumactlOptions()

				res := core_rest.From.Resource(resource)
				jsonRes, err := json.Marshal(res)
				if err != nil {
					return "", err
				}
				yamlRes, err := yaml.JSONToYAML(jsonRes)
				if err != nil {
					return "", err
				}
				return "", kumactl.KumactlApplyFromString(string(yamlRes))
			})
		return err
	}
}

func YamlPathK8s(path string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource by path", DefaultRetries, DefaultTimeout,
			func() (string, error) {
				return "", k8s.KubectlApplyE(cluster.GetTesting(), cluster.GetKubectlOptions(), path)
			})
		return err
	}
}

func Kuma(mode core.CpMode, opt ...KumaDeploymentOption) InstallFunc {
	return func(cluster Cluster) error {
		opt = append(opt, WithIPv6(Config.IPV6))
		return cluster.DeployKuma(mode, opt...)
	}
}

func WaitService(namespace, service string) InstallFunc {
	return func(c Cluster) error {
		k8s.WaitUntilServiceAvailable(c.GetTesting(), c.GetKubectlOptions(namespace), service, 10, 3*time.Second)
		return nil
	}
}

func WaitNumPods(namespace string, num int, app string) InstallFunc {
	return func(c Cluster) error {
		ck8s := c.(*K8sCluster)
		k8s.WaitUntilNumPodsCreated(c.GetTesting(), c.GetKubectlOptions(namespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", app),
			}, num, ck8s.defaultRetries, ck8s.defaultTimeout)
		return nil
	}
}

func WaitPodsAvailable(namespace, app string) InstallFunc {
	return WaitPodsAvailableWithLabel(namespace, "app", app)
}

func WaitPodsAvailableWithLabel(namespace, labelKey, labelValue string) InstallFunc {
	return func(c Cluster) error {
		ck8s := c.(*K8sCluster)
		testingT := c.GetTesting()
		kubectlOptions := c.GetKubectlOptions(namespace)

		pods, err := k8s.ListPodsE(testingT, kubectlOptions,
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelKey, labelValue)})
		if err != nil {
			return err
		}

		var podError error
		for _, p := range pods {
			pod := p
			podError = k8s.WaitUntilPodAvailableE(testingT, kubectlOptions, pod.GetName(), ck8s.defaultRetries, ck8s.defaultTimeout)
			if podError != nil {
				podDetails := ExtractPodDetails(testingT, c.GetKubectlOptions(namespace), pod.Name)
				return &K8sDecoratedError{Err: podError, Details: podDetails}
			}
		}
		return nil
	}
}

func WaitUntilJobSucceed(namespace, app string) InstallFunc {
	return func(c Cluster) error {
		ck8s := c.(*K8sCluster)
		return k8s.WaitUntilJobSucceedE(c.GetTesting(), c.GetKubectlOptions(namespace), app, ck8s.defaultRetries, ck8s.defaultTimeout)
	}
}

func universalZoneProxyRelatedResource(
	tokenProvider func(zone string) (string, error),
	dpName string,
	appType AppMode,
	resourceManifestFunc func(address string, port int) string,
	concurrency int,
) func(cluster Cluster) error {
	return func(cluster Cluster) error {
		uniCluster := cluster.(*UniversalCluster)

		app, err := NewUniversalApp(
			cluster.GetTesting(),
			uniCluster.name,
			dpName,
			"",
			appType,
			Config.IPV6,
			false,
			[]string{},
			[]string{},
			"",
			concurrency,
		)
		if err != nil {
			return err
		}

		app.CreateMainApp(nil, []string{})

		err = app.mainApp.Start()
		if err != nil {
			return err
		}

		uniCluster.apps[dpName] = app
		publicAddress := app.ip
		dpYAML := resourceManifestFunc(publicAddress, UniversalZoneIngressPort)

		token, err := tokenProvider(uniCluster.name)
		if err != nil {
			return err
		}

		switch appType {
		case AppIngress:
			return uniCluster.CreateZoneIngress(app, dpName, publicAddress, dpYAML, token, false)
		case AppEgress:
			return uniCluster.CreateZoneEgress(app, dpName, publicAddress, dpYAML, token, false)
		default:
			return errors.Errorf("unsupported appType: %s", appType)
		}
	}
}

func IngressUniversal(tokenProvider func(zone string) (string, error), opt ...AppDeploymentOption) InstallFunc {
	manifestFunc := func(address string, port int) string {
		return fmt.Sprintf(ZoneIngress, AppIngress, address, UniversalZoneIngressPort, port)
	}

	var opts appDeploymentOptions
	opts.apply(opt...)

	return universalZoneProxyRelatedResource(tokenProvider, AppIngress, AppIngress, manifestFunc, opts.concurrency)
}

func MultipleIngressUniversal(advertisedPort int, tokenProvider func(zone string) (string, error), opt ...AppDeploymentOption) InstallFunc {
	name := fmt.Sprintf("%s-%d", AppIngress, advertisedPort)
	manifestFunc := func(address string, port int) string {
		return fmt.Sprintf(ZoneIngress, name, address, advertisedPort, port)
	}

	var opts appDeploymentOptions
	opts.apply(opt...)

	return universalZoneProxyRelatedResource(tokenProvider, name, AppIngress, manifestFunc, opts.concurrency)
}

func EgressUniversal(tokenProvider func(zone string) (string, error), opt ...AppDeploymentOption) InstallFunc {
	manifestFunc := func(_ string, port int) string {
		return fmt.Sprintf(ZoneEgress, port)
	}

	var opts appDeploymentOptions
	opts.apply(opt...)

	return universalZoneProxyRelatedResource(tokenProvider, AppEgress, AppEgress, manifestFunc, opts.concurrency)
}

func NamespaceWithSidecarInjection(namespace string) InstallFunc {
	return YamlK8s(fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    kuma.io/sidecar-injection: "enabled"
`, namespace))
}

func DemoClientJobK8s(namespace, mesh, destination string) InstallFunc {
	const name = "demo-job-client"
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: batchv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{"app": name},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
					Labels:      map[string]string{"app": name, "kuma.io/mesh": mesh},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           Config.GetUniversalImage(),
							ImagePullPolicy: "IfNotPresent",
							Command:         []string{"curl"},
							Args:            []string{"-v", "-m", "3", "--fail", destination},
						},
					},
					RestartPolicy: "OnFailure",
				},
			},
		},
	}
	return Combine(
		YamlK8sObject(job),
		WaitUntilJobSucceed(namespace, name),
	)
}

func DemoClientUniversal(name string, mesh string, opt ...AppDeploymentOption) InstallFunc {
	return func(cluster Cluster) error {
		var opts appDeploymentOptions
		opts.apply(opt...)
		args := []string{"ncat", "-lvk", "-p", "3000"}
		appYaml := opts.appYaml
		transparent := opts.transparent != nil && *opts.transparent // default false
		additionalTags := ""
		for key, val := range opts.additionalTags {
			additionalTags += fmt.Sprintf(`
      %s: %s`, key, val)
		}
		if appYaml == "" {
			if transparent {
				appYaml = fmt.Sprintf(DemoClientDataplaneTransparentProxy, mesh, "3000", name, additionalTags, redirectPortInbound, redirectPortOutbound, strings.Join(opts.reachableServices, ","))
			} else {
				if opts.serviceProbe {
					appYaml = fmt.Sprintf(DemoClientDataplaneWithServiceProbe, mesh, "13000", "3000", name, additionalTags, "80", "8080")
				} else {
					appYaml = fmt.Sprintf(DemoClientDataplane, mesh, "13000", "3000", name, additionalTags, "80", "8080")
				}
			}
		}

		if !opts.omitDataplane {
			token := opts.token
			var err error
			if token == "" {
				token, err = cluster.GetKuma().GenerateDpToken(mesh, name)
				if err != nil {
					return err
				}
			}
			opt = append(opt, WithToken(token))
		}

		opt = append(
			opt,
			WithName(name),
			WithMesh(mesh),
			WithAppname(AppModeDemoClient),
			WithArgs(args),
			WithYaml(appYaml),
			WithIPv6(Config.IPV6),
		)
		return cluster.DeployApp(opt...)
	}
}

func TcpSinkUniversal(name string, opt ...AppDeploymentOption) InstallFunc {
	return func(cluster Cluster) error {
		var opts appDeploymentOptions
		opts.apply(opt...)
		args := []string{"ncat", "-lk", "9999", ">", "/nc.out"}
		opt = append(
			opt,
			WithName(name),
			WithAppname(AppModeTcpSink),
			WithArgs(args),
			WithoutDataplane(),
			WithTransparentProxy(false),
			WithIPv6(Config.IPV6),
		)
		return cluster.DeployApp(opt...)
	}
}

func TestServerExternalServiceUniversal(name string, port int, tls bool, opt ...AppDeploymentOption) InstallFunc {
	return func(cluster Cluster) error {
		var opts appDeploymentOptions
		opts.apply(opt...)
		args := []string{"test-server", "echo", "--instance", name, "--port", fmt.Sprintf("%d", port)}
		if tls {
			path, err := DumpTempCerts("localhost", opts.dockerContainerName)
			Logf("using temp dir: %s", path)
			if err != nil {
				return err
			}
			args = append(args, "--crt", "/certs/cert.pem", "--key", "/certs/key.pem", "--tls")
			opts.dockerVolumes = append(opts.dockerVolumes, fmt.Sprintf("%s:/certs", path))
		}
		opt = append(opt,
			WithAppname(name),
			WithName(name),
			WithoutDataplane(),
			WithArgs(args),
			WithDockerVolumes(opts.dockerVolumes...),
		)
		return cluster.DeployApp(opt...)
	}
}

func TestServerUniversal(name string, mesh string, opt ...AppDeploymentOption) InstallFunc {
	return func(cluster Cluster) error {
		var opts appDeploymentOptions
		opts.apply(opt...)
		if len(opts.protocol) == 0 {
			opts.protocol = "http"
		}
		if opts.serviceVersion == "" {
			opts.serviceVersion = "v1"
		}
		args := []string{"test-server"}
		if len(opts.appArgs) > 0 {
			args = append(args, opts.appArgs...)
		}
		if opts.serviceName == "" {
			opts.serviceName = "test-server"
		}
		if opts.serviceInstance == "" {
			opts.serviceInstance = "1"
		}
		transparent := opts.transparent == nil || *opts.transparent // default true
		transparentProxy := ""
		if transparent {
			transparentProxy = fmt.Sprintf(`
  transparentProxying:
    redirectPortInbound: %s
    redirectPortOutbound: %s
`, redirectPortInbound, redirectPortOutbound)
		}
		token := opts.token
		var err error
		if token == "" {
			token, err = cluster.GetKuma().GenerateDpToken(mesh, opts.serviceName)
			if err != nil {
				return err
			}
		}

		additionalTags := ""
		for key, val := range opts.additionalTags {
			additionalTags += fmt.Sprintf(`
      %s: %s`, key, val)
		}

		serviceProbe := ""
		if opts.serviceProbe {
			serviceProbe = `    serviceProbe:
      tcp: {}`
		}

		serviceAddress := ""
		if opts.serviceAddress != "" {
			serviceAddress = fmt.Sprintf(`    serviceAddress: %s`, opts.serviceAddress)
		}

		if len(args) < 2 || args[1] != "grpc" { // grpc client does not have port
			args = append(args, "--port", "8080")
		}
		appYaml := fmt.Sprintf(`
type: Dataplane
mesh: %s
name: {{ name }}
networking:
  address:  {{ address }}
  inbound:
  - port: %s
    servicePort: %s
%s
    tags:
      kuma.io/service: %s
      kuma.io/protocol: %s
      version: %s
      instance: '%s'
      team: server-owners
%s
%s
%s
%s
`, mesh, "80", "8080", serviceAddress, opts.serviceName, opts.protocol, opts.serviceVersion, opts.serviceInstance, additionalTags, serviceProbe, transparentProxy, opts.appendDataplaneConfig)

		opt = append(opt,
			WithName(name),
			WithMesh(mesh),
			WithAppname(opts.serviceName),
			WithTransparentProxy(transparent),
			WithToken(token),
			WithArgs(args),
			WithYaml(appYaml),
			WithIPv6(Config.IPV6),
			WithDpEnvs(opts.dpEnvs),
		)
		return cluster.DeployApp(opt...)
	}
}

func Combine(fs ...InstallFunc) InstallFunc {
	return func(cluster Cluster) error {
		for _, f := range fs {
			if err := f(cluster); err != nil {
				return err
			}
		}
		return nil
	}
}

func CombineWithRetries(maxRetries int, fs ...InstallFunc) InstallFunc {
	return func(cluster Cluster) error {
		for _, f := range fs {
			_, err := retry.DoWithRetryE(
				cluster.GetTesting(),
				"installing component to cluster",
				maxRetries,
				0,
				func() (string, error) {
					return "", f(cluster)
				})
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func Namespace(name string) InstallFunc {
	return func(cluster Cluster) error {
		return k8s.CreateNamespaceE(cluster.GetTesting(), cluster.GetKubectlOptions(), name)
	}
}

type ClusterSetup struct {
	installFuncs []InstallFunc
}

func NewClusterSetup() *ClusterSetup {
	return &ClusterSetup{}
}

func (cs *ClusterSetup) Install(fn InstallFunc) *ClusterSetup {
	cs.installFuncs = append(cs.installFuncs, fn)
	return cs
}

func (cs *ClusterSetup) Setup(cluster Cluster) error {
	return Combine(cs.installFuncs...)(cluster)
}

func (cs *ClusterSetup) SetupWithRetries(cluster Cluster, maxRetries int) error {
	return CombineWithRetries(maxRetries, cs.installFuncs...)(cluster)
}

func CreateCertsFor(names ...string) (string, string, error) {
	keyPair, err := tls.NewSelfSignedCert(tls.ServerCertType, tls.DefaultKeyType, names...)
	if err != nil {
		return "", "", err
	}

	return string(keyPair.CertPEM), string(keyPair.KeyPEM), nil
}

func DumpTempCerts(names ...string) (string, error) {
	cert, key, err := CreateCertsFor(names...)
	if err != nil {
		return "", err
	}
	path, err := os.MkdirTemp("", "cert-*")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(
		filepath.Join(path, "cert.pem"),
		[]byte(fmt.Sprintf("---\n%s", cert)),
		os.ModePerm, // #nosec G306
	); err != nil {
		return "", err
	}
	if err := os.WriteFile(
		filepath.Join(path, "key.pem"),
		[]byte(fmt.Sprintf("---\n%s", key)),
		os.ModePerm, // #nosec G306
	); err != nil {
		return "", err
	}
	return path, nil
}
