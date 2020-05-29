package framework

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/log"
)

type TestFramework struct {
	testing.T
	k8sclusters []string
	kumactl     string
	verbose     bool
}

// NewK8sTest gets the number of the clusters to use in the tests, and the pattern
// to locate the KUBECONFIG for them. The second argument can be empty
func NewK8sTest(numClusters int, kubeConfigPathPattern string, verbose bool) *TestFramework {
	if numClusters < 1 || numClusters > maxClusters {
		log.Error("Invalid cluster number. Should be in the range [1,3], but it is ", numClusters)
		return nil
	}
	if kubeConfigPathPattern == "" {
		kubeConfigPathPattern = defaultKubeConfigPathPattern
	}
	kubeconfigs := []string{""} // have an empty cluster at [0]
	for i := 1; i <= numClusters; i++ {
		kubeconfigs = append(kubeconfigs,
			os.ExpandEnv(fmt.Sprintf(kubeConfigPathPattern, i)))
	}

	kumactl := os.Getenv(envKUMACTLBIN)
	if kumactl == "" {
		log.Error("Unable to find kumactl, please supply valid KUMACTL environment variable.")
		return nil
	}

	return &TestFramework{
		k8sclusters: kubeconfigs,
		kumactl:     kumactl,
		verbose:     verbose,
	}
}

func (t *TestFramework) ApplyOnK8sCluster(idx int, namespace string, yaml string) error {
	options := k8s.NewKubectlOptions("", "", namespace)
	defer k8s.KubectlDelete(t, options, t.k8sclusters[idx])
	k8s.KubectlApply(t, options, yaml)
	return nil
}

func (t *TestFramework) ApplyAndWaitServiceOnK8sCluster(idx int, namespace string, yaml string, service string) error {
	options := k8s.NewKubectlOptions("", "", namespace)
	defer k8s.KubectlDelete(t, options, t.k8sclusters[idx])
	k8s.KubectlApply(t, options, yaml)
	k8s.WaitUntilServiceAvailable(t, options, service, defaultRetries, defaultTiemout)
	return nil
}

func (t *TestFramework) DeployKumaOnK8sCluster(idx int) error {
	options := NewKumactlOptions("", "", t.verbose)

	yaml, err := KumactlInstallCP(t, options)
	if err != nil {
		return err
	}

	err = k8s.KubectlApplyFromStringE(t,
		&k8s.KubectlOptions{
			ConfigPath: t.k8sclusters[idx],
		},
		yaml)
	if err != nil {
		return err
	}

	k8s.WaitUntilServiceAvailable(t,
		&k8s.KubectlOptions{
			ConfigPath: t.k8sclusters[idx],
			Namespace:  kumaNamespace,
		},
		kumaServiceName, defaultRetries, defaultTiemout)

	k8s.WaitUntilNumPodsCreated(t,
		&k8s.KubectlOptions{
			ConfigPath: t.k8sclusters[idx],
			Namespace:  kumaNamespace,
		},
		metav1.ListOptions{
			LabelSelector: "app=" + kumaServiceName,
		},
		1, defaultRetries, defaultTiemout)

	kumacp_pods := k8s.ListPods(t, &k8s.KubectlOptions{
		ConfigPath: t.k8sclusters[idx],
		Namespace:  kumaNamespace,
	},
		metav1.ListOptions{
			LabelSelector: "app=" + kumaServiceName,
		},
	)

	Expect(len(kumacp_pods)).To(Equal(1))

	k8s.WaitUntilPodAvailable(t,
		&k8s.KubectlOptions{
			ConfigPath: t.k8sclusters[idx],
			Namespace:  kumaNamespace,
		},
		kumacp_pods[0].Name,
		10, 3*time.Second)

	t.PortForwardServiceOnK8sCluster(idx, kumaNamespace, kumaServiceName, 5681, 5681)

	return nil
}

func (t *TestFramework) DeleteKumaOnK8sCluster(idx int) error {
	options := NewKumactlOptions("", "", t.verbose)

	yaml, err := KumactlInstallCP(t, options)
	if err != nil {
		return err
	}

	err = k8s.KubectlDeleteFromStringE(t,
		&k8s.KubectlOptions{
			ConfigPath: t.k8sclusters[idx],
		},
		yaml)

	return err
}

func (t *TestFramework) DeleteKumaNamespaceOnK8sCluster(idx int) error {
	return k8s.DeleteNamespaceE(t,
		&k8s.KubectlOptions{
			ConfigPath: t.k8sclusters[idx],
		}, kumaNamespace)
}

func (t *TestFramework) PortForwardServiceOnK8sCluster(idx int, namespace string, service string, localPort, remotePort int) {
	options := k8s.NewKubectlOptions("", t.k8sclusters[idx], namespace)
	go func() {
		_ = k8s.RunKubectlE(t, options, "port-forward", "service/"+service,
			strconv.Itoa(localPort)+":"+strconv.Itoa(remotePort))
	}()
}

func (t *TestFramework) VerifyKumaOnK8sCluster() error {
	return http_helper.HttpGetWithRetryWithCustomValidationE(
		t,
		"http://localhost:5681",
		&tls.Config{},
		defaultRetries/2,
		defaultTiemout,
		func(statusCode int, body string) bool {
			return statusCode == http.StatusOK
		},
	)
}

func (t *TestFramework) GetPodLogs(pod v1.Pod) string {
	podLogOpts := v1.PodLogOptions{}
	config, err := rest.InClusterConfig()
	if err != nil {
		return "error in getting config"
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "error in getting access to K8S"
	}
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return "error in opening stream"
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "error in copy information from podLogs to buf"
	}
	str := buf.String()

	return str
}

func IsK8sClustersStarted() bool {
	_, found := os.LookupEnv(envK8SCLUSTERS)
	return found
}
