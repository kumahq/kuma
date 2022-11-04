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

const (
	demoClientDeployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-client
  namespace: %s
  labels:
    app: demo-client
spec:
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: demo-client
  template:
    metadata:
      annotations:
        kuma.io/mesh: %s
      labels:
        app: demo-client
    spec:
      containers:
        - name: demo-client
          image: %s
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 3000
          command: [ "ncat" ]
          args:
            - -lk
            - -p
            - "3000"
          resources:
            limits:
              cpu: 50m
              memory: 64Mi
`
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
			func() (s string, err error) {
				return "", k8s.KubectlApplyFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), string(bytes))
			})
		return err
	}
}

func YamlK8s(yaml string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				return "", k8s.KubectlApplyFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), yaml)
			})
		return err
	}
}

func DeleteYamlK8s(yaml string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "delete yaml resource", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				return "", k8s.KubectlDeleteFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), yaml)
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

func YamlUniversal(yaml string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				kumactl := cluster.GetKumactlOptions()
				return "", kumactl.KumactlApplyFromString(yaml)
			})
		return err
	}
}

func ResourceUniversal(resource model.Resource) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install resource", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
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
			func() (s string, err error) {
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
		k8s.WaitUntilNumPodsCreated(c.GetTesting(), c.GetKubectlOptions(namespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", app),
			}, num, DefaultRetries, DefaultTimeout)
		return nil
	}
}

func WaitPodsAvailable(namespace, app string) InstallFunc {
	return func(c Cluster) error {
		pods, err := k8s.ListPodsE(c.GetTesting(), c.GetKubectlOptions(namespace),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", app)})
		if err != nil {
			return err
		}
		for _, p := range pods {
			err := k8s.WaitUntilPodAvailableE(c.GetTesting(), c.GetKubectlOptions(namespace), p.GetName(), DefaultRetries, DefaultTimeout)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WaitUntilJobSucceed(namespace, app string) InstallFunc {
	return func(c Cluster) error {
		return k8s.WaitUntilJobSucceedE(c.GetTesting(), c.GetKubectlOptions(namespace), app, DefaultRetries, DefaultTimeout)
	}
}

func zoneRelatedResource(
	tokenProvider func(zone string) (string, error),
	appType AppMode,
	resourceManifestFunc func(address string, port, advertisedPort int) string,
) func(cluster Cluster) error {
	dpName := string(appType)

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
		dpYAML := resourceManifestFunc(publicAddress, kdsPort, kdsPort)

		zone := uniCluster.name
		if uniCluster.controlplane.mode == core.Standalone {
			zone = ""
		}
		token, err := tokenProvider(zone)
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

func IngressUniversal(tokenProvider func(zone string) (string, error)) InstallFunc {
	manifestFunc := func(address string, port, advertisedPort int) string {
		return fmt.Sprintf(ZoneIngress, address, port, advertisedPort)
	}

	return zoneRelatedResource(tokenProvider, AppIngress, manifestFunc)
}

func EgressUniversal(tokenProvider func(zone string) (string, error)) InstallFunc {
	manifestFunc := func(_ string, port, _ int) string {
		return fmt.Sprintf(ZoneEgress, port)
	}

	return zoneRelatedResource(tokenProvider, AppEgress, manifestFunc)
}

func DemoClientK8sWithAffinity(mesh string, namespace string) InstallFunc {
	affinity := `      nodeSelector:
        second: "true"
`

	return DemoClientK8sCustomized(mesh, namespace, demoClientDeployment+affinity)
}

func DemoClientK8s(mesh string, namespace string) InstallFunc {
	return DemoClientK8sCustomized(mesh, namespace, demoClientDeployment)
}

func DemoClientK8sCustomized(mesh string, namespace string, deploymentYaml string) InstallFunc {
	const name = "demo-client"
	return Combine(
		YamlK8s(fmt.Sprintf(deploymentYaml, namespace, mesh, Config.GetUniversalImage())),
		WaitNumPods(namespace, 1, name),
		WaitPodsAvailable(namespace, name),
	)
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

// NamespaceWithSidecarInjectionOnAnnotation creates namespace with sidecar-injection annotation
// Since we still support annotations for backwards compatibility, we should also test it.
// Use NamespaceWithSidecarInjection unless you want to explicitly check backwards compatibility.
// https://github.com/kumahq/kuma/issues/4005
func NamespaceWithSidecarInjectionOnAnnotation(namespace string) InstallFunc {
	return YamlK8s(fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  annotations:
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
					Annotations: map[string]string{"kuma.io/mesh": mesh},
					Labels:      map[string]string{"app": name},
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
		if appYaml == "" {
			if transparent {
				appYaml = fmt.Sprintf(DemoClientDataplaneTransparentProxy, mesh, "3000", name, redirectPortInbound, redirectPortInboundV6, redirectPortOutbound, strings.Join(opts.reachableServices, ","))
			} else {
				if opts.serviceProbe {
					appYaml = fmt.Sprintf(DemoClientDataplaneWithServiceProbe, mesh, "13000", "3000", name, "80", "8080")
				} else {
					appYaml = fmt.Sprintf(DemoClientDataplane, mesh, "13000", "3000", name, "80", "8080")
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

func TestServerExternalServiceUniversal(name string, mesh string, port int, tls bool) InstallFunc {
	return func(cluster Cluster) error {
		containerName := fmt.Sprintf("%s.%s", name, mesh)
		args := []string{"test-server", "echo", "--instance", name, "--port", fmt.Sprintf("%d", port)}
		opt := []AppDeploymentOption{
			WithAppname(name),
			WithName(name),
			WithMesh(mesh),
			WithoutDataplane(),
			WithDockerContainerName(containerName),
		}
		if tls {
			path, err := DumpTempCerts("localhost", containerName)
			Logf("using temp dir: %s", path)
			if err != nil {
				return err
			}
			args = append(args, "--crt", "/certs/cert.pem", "--key", "/certs/key.pem", "--tls")
			opt = append(opt, WithDockerVolumes(path+":/certs"))
		}

		opt = append(opt, WithArgs(args))
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
    redirectPortInboundV6: %s
    redirectPortOutbound: %s
`, redirectPortInbound, redirectPortInboundV6, redirectPortOutbound)
		}
		token := opts.token
		var err error
		if token == "" {
			token, err = cluster.GetKuma().GenerateDpToken(mesh, opts.serviceName)
			if err != nil {
				return err
			}
		}

		serviceProbe := ""
		if opts.serviceProbe {
			serviceProbe =
				`    serviceProbe:
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
`, mesh, "80", "8080", serviceAddress, opts.serviceName, opts.protocol, opts.serviceVersion, opts.serviceInstance, serviceProbe, transparentProxy, opts.appendDataplaneConfig)

		opt = append(opt,
			WithName(name),
			WithMesh(mesh),
			WithAppname(opts.serviceName),
			WithTransparentProxy(transparent),
			WithToken(token),
			WithArgs(args),
			WithYaml(appYaml),
			WithIPv6(Config.IPV6))
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

func CreateCertsFor(names ...string) (cert, key string, err error) {
	keyPair, err := tls.NewSelfSignedCert("kuma", tls.ServerCertType, tls.DefaultKeyType, names...)
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
	if err := os.WriteFile(filepath.Join(path, "cert.pem"), []byte(fmt.Sprintf("---\n%s", cert)), os.ModePerm); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(path, "key.pem"), []byte(fmt.Sprintf("---\n%s", key)), os.ModePerm); err != nil {
		return "", err
	}
	return path, nil
}
