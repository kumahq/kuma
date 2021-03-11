package framework

import (
	"fmt"
	"os"
	"time"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/testing"

	"github.com/kumahq/kuma/pkg/tls"

	"github.com/go-errors/errors"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InstallFunc func(cluster Cluster) error

func YamlK8s(yaml string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				return "", k8s.KubectlApplyFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), yaml)
			})
		return err
	}
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

func Kuma(mode string, fs ...DeployOptionsFunc) InstallFunc {
	return func(cluster Cluster) error {
		err := cluster.DeployKuma(mode, fs...)
		return err
	}
}

func KumaDNS() InstallFunc {
	return func(cluster Cluster) error {
		err := cluster.InjectDNS(KumaNamespace)
		return err
	}
}

func WaitService(namespace, service string) InstallFunc {
	return func(c Cluster) error {
		k8s.WaitUntilServiceAvailable(c.GetTesting(), c.GetKubectlOptions(namespace), service, 10, 3*time.Second)
		return nil
	}
}

func WaitNumPods(num int, app string) InstallFunc {
	return func(c Cluster) error {
		k8s.WaitUntilNumPodsCreated(c.GetTesting(), c.GetKubectlOptions(),
			kube_meta.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", app),
			}, num, DefaultRetries, DefaultTimeout)
		return nil
	}
}

func WaitPodsAvailable(namespace, app string) InstallFunc {
	return func(c Cluster) error {
		pods, err := k8s.ListPodsE(c.GetTesting(), c.GetKubectlOptions(namespace),
			kube_meta.ListOptions{LabelSelector: fmt.Sprintf("app=%s", app)})
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

func WaitUntilPodCompleteE(t testing.TestingT, options *k8s.KubectlOptions, podName string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for pod %s to be provisioned.", podName)
	message, err := retry.DoWithRetryE(
		t,
		statusMsg,
		retries,
		sleepBetweenRetries,
		func() (string, error) {
			pod, err := k8s.GetPodE(t, options, podName)
			if err != nil {
				return "", err
			}
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Terminated == nil || cs.State.Terminated.ExitCode != 0 {
					return "", errors.Errorf("Pod is not complete yet")
				}
			}
			return "Pod is now complete", nil
		},
	)
	if err != nil {
		logger.Logf(t, "Timedout waiting for Pod to be completed: %s", err)
		return err
	}
	logger.Logf(t, message)
	return nil
}

func WaitPodsComplete(namespace, app string) InstallFunc {
	return func(c Cluster) error {
		pods, err := k8s.ListPodsE(c.GetTesting(), c.GetKubectlOptions(namespace),
			kube_meta.ListOptions{LabelSelector: fmt.Sprintf("app=%s", app)})
		if err != nil {
			return err
		}
		for _, p := range pods {
			err := WaitUntilPodCompleteE(c.GetTesting(), c.GetKubectlOptions(namespace), p.GetName(), DefaultRetries, DefaultTimeout)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WaitPodsNotAvailable(namespace, app string) InstallFunc {
	return func(c Cluster) error {
		pods, err := k8s.ListPodsE(c.GetTesting(), c.GetKubectlOptions(namespace),
			kube_meta.ListOptions{LabelSelector: fmt.Sprintf("app=%s", app)})
		if err != nil {
			return err
		}

		for _, p := range pods {
			_, _ = retry.DoWithRetryE(
				c.GetTesting(),
				"Wait pod deletion",
				DefaultRetries,
				DefaultTimeout,
				func() (string, error) {
					pod, err := k8s.GetPodE(c.GetTesting(), c.GetKubectlOptions(namespace), p.GetName())
					if err == nil {
						return "", err
					}
					if !k8s.IsPodAvailable(pod) {
						return "Pod is not available", nil
					}
					return "", errors.Errorf("Pod is still available")
				},
			)
		}
		return nil
	}
}

func EchoServerK8s(mesh string) InstallFunc {
	image := "kuma-universal"

	if i := os.Getenv("KUMA_UNIVERSAL_IMAGE"); i != "" {
		image = i
	}

	const name = "echo-server"
	service := `
apiVersion: v1
kind: Service
metadata:
  name: echo-server
  namespace: kuma-test
  annotations:
    80.service.kuma.io/protocol: http
spec:
  ports:
    - port: 80
      name: http
  selector:
    app: echo-server
`
	deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo-server
  namespace: kuma-test
  labels:
    app: echo-server
spec:
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: echo-server
  template:
    metadata:
      annotations:
        kuma.io/mesh: %s
      labels:
        app: echo-server
    spec:
      containers:
        - name: echo-server
          image: ` + image + `
          imagePullPolicy: IfNotPresent
          readinessProbe:
            httpGet:
              path: /
              port: 80
            initialDelaySeconds: 3
            periodSeconds: 3
          ports:
            - containerPort: 80
          command: [ "ncat" ]
          args:
            - -lk
            - -p
            - "80"
            - --sh-exec
            - '/usr/bin/printf "HTTP/1.1 200 OK\n\n Echo\n"'
          resources:
            limits:
              cpu: 50m
              memory: 128Mi
`
	return Combine(
		YamlK8s(service),
		YamlK8s(fmt.Sprintf(deployment, mesh)),
		WaitService(TestNamespace, name),
		WaitNumPods(1, name),
		WaitPodsAvailable(TestNamespace, name),
	)
}

func EchoServerUniversal(name, mesh, echo, token string, fs ...DeployOptionsFunc) InstallFunc {
	return func(cluster Cluster) error {
		opts := newDeployOpt(fs...)
		args := []string{"ncat", "-lk", "-p", "80", "--sh-exec", "'echo \"HTTP/1.1 200 OK\n\n Echo " + echo + "\n\"'"}
		appYaml := ""
		if opts.protocol == "" {
			opts.protocol = "http"
		}
		switch {
		case opts.transparent:
			appYaml = fmt.Sprintf(EchoServerDataplaneTransparentProxy, mesh, "8080", "80", "8080", redirectPortInbound, redirectPortOutbound)
		case opts.serviceProbe:
			appYaml = fmt.Sprintf(EchoServerDataplaneWithServiceProbe, mesh, "8080", "80", "8080", opts.protocol)
		default:
			appYaml = fmt.Sprintf(EchoServerDataplane, mesh, "8080", "80", "8080", opts.protocol)
		}
		fs = append(fs, WithName(name), WithMesh(mesh), WithAppname(AppModeEchoServer), WithToken(token), WithArgs(args), WithYaml(appYaml))
		return cluster.DeployApp(fs...)
	}
}

func IngressUniversal(mesh, token string) InstallFunc {
	return func(cluster Cluster) error {
		uniCluster := cluster.(*UniversalCluster)
		app, err := NewUniversalApp(cluster.GetTesting(), uniCluster.name, AppIngress, AppIngress, true, []string{})
		if err != nil {
			return err
		}

		app.CreateMainApp([]string{}, []string{})

		err = app.mainApp.Start()
		if err != nil {
			return err
		}
		uniCluster.apps[AppIngress] = app

		publicAddress := uniCluster.apps[AppIngress].ip
		dpyaml := fmt.Sprintf(IngressDataplane, mesh, publicAddress, kdsPort, kdsPort)
		return uniCluster.CreateDP(app, "ingress", app.ip, dpyaml, token)
	}
}

func DemoClientK8s(mesh string) InstallFunc {
	image := "kuma-universal"

	if i := os.Getenv("KUMA_UNIVERSAL_IMAGE"); i != "" {
		image = i
	}

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
          image: ` + image + `
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
		YamlK8s(fmt.Sprintf(deployment, mesh)),
		WaitNumPods(1, name),
		WaitPodsAvailable(TestNamespace, name),
	)
}

func DemoClientJobK8s(mesh, destination string) InstallFunc {
	image := "kuma-universal"

	if i := os.Getenv("KUMA_UNIVERSAL_IMAGE"); i != "" {
		image = i
	}

	const name = "demo-job-client"
	deployment := `
apiVersion: batch/v1
kind: Job
metadata:
  name: demo-job-client
  namespace: kuma-test
  labels:
    app: demo-job-client
spec:
  template:
    metadata:
      annotations:
        kuma.io/mesh: %s
      labels:
        app: demo-job-client
    spec:
      containers:
      - name: demo-job-client
        image: ` + image + `
        imagePullPolicy: IfNotPresent
        command: [ "curl" ]
        args:
          - -v
          - -m
          - "3"
          - --fail
          - %s
      restartPolicy: OnFailure
`
	return Combine(
		YamlK8s(fmt.Sprintf(deployment, mesh, destination)),
		WaitNumPods(1, name),
		WaitPodsComplete(TestNamespace, name),
	)
}

func DemoClientUniversal(name, mesh, token string, fs ...DeployOptionsFunc) InstallFunc {
	return func(cluster Cluster) error {
		opts := newDeployOpt(fs...)
		args := []string{"ncat", "-lvk", "-p", "3000"}
		appYaml := ""
		if opts.transparent {
			appYaml = fmt.Sprintf(DemoClientDataplaneTransparentProxy, mesh, "3000", redirectPortInbound, redirectPortOutbound)
		} else {
			if opts.serviceProbe {
				appYaml = fmt.Sprintf(DemoClientDataplaneWithServiceProbe, mesh, "13000", "3000", "80", "8080")
			} else {
				appYaml = fmt.Sprintf(DemoClientDataplane, mesh, "13000", "3000", "80", "8080")
			}
		}
		fs = append(fs, WithName(name), WithMesh(mesh), WithAppname(AppModeDemoClient), WithToken(token), WithArgs(args), WithYaml(appYaml))
		return cluster.DeployApp(fs...)
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

func CreateCertsForIP(ip string) (cert, key string, err error) {
	keyPair, err := tls.NewSelfSignedCert("kuma", tls.ServerCertType, "localhost", ip)
	if err != nil {
		return "", "", err
	}

	return string(keyPair.CertPEM), string(keyPair.KeyPEM), nil
}
