package framework

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/kumahq/kuma/pkg/config/core"
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

func MeshUniversal(name string) InstallFunc {
	mesh := fmt.Sprintf(`
type: Mesh
name: %s
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

func KumaDNS() InstallFunc {
	return func(cluster Cluster) error {
		err := cluster.InjectDNS(Config.KumaNamespace)
		return err
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

func DemoClientK8s(mesh string) InstallFunc {
	const name = "demo-client"
	deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-client
  namespace: kuma-test
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
              memory: 128Mi
`
	return Combine(
		YamlK8s(fmt.Sprintf(deployment, mesh, Config.GetUniversalImage())),
		WaitNumPods(TestNamespace, 1, name),
		WaitPodsAvailable(TestNamespace, name),
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
		appYaml := ""
		if opts.transparent {
			appYaml = fmt.Sprintf(DemoClientDataplaneTransparentProxy, mesh, "3000", name, redirectPortInbound, redirectPortInboundV6, redirectPortOutbound, strings.Join(opts.reachableServices, ","))
		} else {
			if opts.serviceProbe {
				appYaml = fmt.Sprintf(DemoClientDataplaneWithServiceProbe, mesh, "13000", "3000", name, "80", "8080")
			} else {
				appYaml = fmt.Sprintf(DemoClientDataplane, mesh, "13000", "3000", name, "80", "8080")
			}
		}

		token := opts.token
		var err error
		if token == "" {
			token, err = cluster.GetKuma().GenerateDpToken(mesh, name)
			if err != nil {
				return err
			}
		}

		opt = append(opt,
			WithName(name),
			WithMesh(mesh),
			WithAppname(AppModeDemoClient),
			WithToken(token),
			WithArgs(args),
			WithYaml(appYaml),
			WithIPv6(Config.IPV6))
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

		args = append(args, "--port", "8080")
		appYaml := fmt.Sprintf(`
type: Dataplane
mesh: %s
name: {{ name }}
networking:
  address:  {{ address }}
  inbound:
  - port: %s
    servicePort: %s
    tags:
      kuma.io/service: %s
      kuma.io/protocol: %s
      version: %s
      instance: '%s'
      team: server-owners
%s
  transparentProxying:
    redirectPortInbound: %s
    redirectPortInboundV6: %s
    redirectPortOutbound: %s
`, mesh, "80", "8080", opts.serviceName, opts.protocol, opts.serviceVersion, opts.serviceInstance, serviceProbe, redirectPortInbound, redirectPortInboundV6, redirectPortOutbound)

		opt = append(opt,
			WithName(name),
			WithMesh(mesh),
			WithAppname("test-server"),
			WithTransparentProxy(true), // test server is always meant to be used with transparent proxy
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
