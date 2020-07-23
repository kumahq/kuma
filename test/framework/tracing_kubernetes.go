package framework

import (
	"encoding/json"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sort"
)

type K8SJaeger struct {
	port uint32
}

func (j *K8SJaeger) ZipkinCollectorURL() string {
	return "http://jaeger-collector.kuma-tracing:9411/api/v2/spans"
}

func (j *K8SJaeger) TracedServices() ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/api/services", j.port))
	if err != nil {
		return nil, err
	}
	output := &jaegerServicesOutput{}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return nil, err
	}
	sort.Strings(output.Data)
	return output.Data, nil
}

func DeployTracingK8S(cluster *K8sCluster) (*K8SJaeger, error) {
	kumactl, _ := NewKumactlOptions(cluster.GetTesting(), cluster.GetKuma().GetName(), true)
	yaml, err := kumactl.KumactlInstallTracing()
	if err != nil {
		return nil, err
	}
	err = k8s.KubectlApplyFromStringE(cluster.t,
		cluster.GetKubectlOptions(),
		yaml)
	if err != nil {
		return nil, err
	}

	k8s.WaitUntilNumPodsCreated(cluster.t,
		cluster.GetKubectlOptions(tracingNamespace),
		metav1.ListOptions{
			LabelSelector: "app=jaeger",
		},
		1,
		DefaultRetries,
		DefaultTimeout)

	pods := k8s.ListPods(cluster.t,
		cluster.GetKubectlOptions(tracingNamespace),
		metav1.ListOptions{
			LabelSelector: "app=jaeger",
		},
	)
	if len(pods) != 1 {
		return nil, errors.Errorf("counting Jaeger pods. Got: %d. Expected: 1", len(pods))
	}

	k8s.WaitUntilPodAvailable(cluster.t,
		cluster.GetKubectlOptions(tracingNamespace),
		pods[0].Name,
		DefaultRetries,
		DefaultTimeout)

	port, err := test.FindFreePort("")
	if err != nil {
		return nil, err
	}

	cluster.PortForwardPod(tracingNamespace, pods[0].Name, port, 16686)

	return &K8SJaeger{port}, nil
}

func DeleteTracingK8S(cluster *K8sCluster) error {
	kumactl, _ := NewKumactlOptions(cluster.GetTesting(), cluster.GetKuma().GetName(), true)
	yaml, err := kumactl.KumactlInstallTracing()
	if err != nil {
		return err
	}

	err = k8s.KubectlDeleteFromStringE(cluster.GetTesting(),
		cluster.GetKubectlOptions(),
		yaml)

	cluster.WaitNamespaceDelete(tracingNamespace)

	return err
}
