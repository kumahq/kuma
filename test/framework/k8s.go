package framework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"

	bootstrap_k8s "github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s"
)

func PodNameOfApp(cluster Cluster, name string, namespace string) (string, error) {
	pod, err := PodOfApp(cluster, name, namespace)
	if err != nil {
		return "", err
	}

	return pod.Name, nil
}

func PodOfApp(cluster Cluster, name string, namespace string) (v1.Pod, error) {
	pods, err := k8s.ListPodsE(
		cluster.GetTesting(),
		cluster.GetKubectlOptions(namespace),
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", name),
		},
	)
	if err != nil {
		return v1.Pod{}, err
	}
	if len(pods) != 1 {
		return v1.Pod{}, errors.Errorf("expected %d pods, got %d", 1, len(pods))
	}
	return pods[0], nil
}

func PodIPOfApp(cluster Cluster, name string, namespace string) (string, error) {
	pods, err := k8s.ListPodsE(
		cluster.GetTesting(),
		cluster.GetKubectlOptions(namespace),
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", name),
		},
	)
	if err != nil {
		return "", err
	}
	if len(pods) != 1 {
		return "", errors.Errorf("expected %d pods, got %d", 1, len(pods))
	}
	return pods[0].Status.PodIP, nil
}

func GatewayAPICRDs(cluster Cluster) error {
	return k8s.RunKubectlE(
		cluster.GetTesting(),
		cluster.GetKubectlOptions(),
		"apply", "-f", "https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.1.0/experimental-install.yaml")
}

func UpdateKubeObject(
	t testing.TestingT,
	k8sOpts *k8s.KubectlOptions,
	typeName string,
	objectName string,
	update func(object runtime.Object) runtime.Object,
) error {
	scheme, err := bootstrap_k8s.NewScheme()
	if err != nil {
		return err
	}
	codecs := serializer.NewCodecFactory(scheme)
	info, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
	if !ok {
		return errors.Errorf("no serializer for %q", runtime.ContentTypeYAML)
	}

	_, err = retry.DoWithRetryableErrorsE(t, "update object", map[string]string{"Error from server \\(Conflict\\)": "object conflict"}, 5, time.Second, func() (string, error) {
		out, err := k8s.RunKubectlAndGetOutputE(t, k8sOpts, "get", typeName, objectName, "-o", "yaml")
		if err != nil {
			return "", err
		}

		decoder := yaml.NewYAMLToJSONDecoder(bytes.NewReader([]byte(out)))
		into := map[string]interface{}{}

		if err := decoder.Decode(&into); err != nil {
			return "", err
		}

		u := unstructured.Unstructured{Object: into}
		obj, err := scheme.New(u.GroupVersionKind())
		if err != nil {
			return "", err
		}

		if err := scheme.Convert(&u, obj, nil); err != nil {
			return "", err
		}

		obj = update(obj)
		encoder := codecs.EncoderForVersion(info.Serializer, obj.GetObjectKind().GroupVersionKind().GroupVersion())
		yml, err := runtime.Encode(encoder, obj)
		if err != nil {
			return "", err
		}

		if err := k8s.KubectlApplyFromStringE(t, k8sOpts, string(yml)); err != nil {
			return "", err
		}
		return "", nil
	})
	if err != nil {
		return errors.Wrapf(err, "failed to update %s object %q", typeName, objectName)
	}
	return nil
}

type K8sDecoratedError struct {
	Err     error
	Details *ObjectDetails
}

func (e *K8sDecoratedError) Error() string {
	detailStr := MarshalObjectDetails(e.Details)
	return fmt.Sprintf("sourceError: %s K8sDetails: %q", e.Err.Error(), detailStr)
}

func MarshalObjectDetails(e *ObjectDetails) string {
	details := "none"
	if e != nil {
		b, err := json.Marshal(*e)
		if err != nil {
			details = fmt.Sprintf("failed to marshal details, err: %v", err)
		} else {
			details = string(b)
		}
	}
	return details
}

func ExtractDeploymentDetails(testingT testing.TestingT,
	kubectlOptions *k8s.KubectlOptions, name string,
) *ObjectDetails {
	deploy, err := k8s.GetDeploymentE(testingT, kubectlOptions, name)
	if err != nil {
		// might not be a Deployment, let's ignore it
		return &ObjectDetails{RetrievalError: err}
	}

	deployDetails := ObjectDetails{
		Namespace:  kubectlOptions.Namespace,
		Kind:       "Deployment",
		Name:       name,
		Conditions: fromDeploymentCondition(deploy.Status.Conditions),
		Events:     getObjectEvents(testingT, kubectlOptions, "Deployment", name),
	}

	replicaSets, err := k8s.ListReplicaSetsE(testingT, kubectlOptions, metav1.ListOptions{
		LabelSelector: "app=" + name,
	})
	if err != nil {
		deployDetails.RetrievalError = err
	}
	for _, ars := range replicaSets {
		deployDetails.ReplicaSets = append(deployDetails.ReplicaSets, ObjectDetails{
			Name:       ars.Name,
			Conditions: fromReplicaSetCondition(ars.Status.Conditions),
			Events:     getObjectEvents(testingT, kubectlOptions, "ReplicaSet", ars.Name),
		})
	}

	pods, err := k8s.ListPodsE(testingT, kubectlOptions, metav1.ListOptions{
		LabelSelector: "app=" + name,
	})
	if err != nil {
		deployDetails.RetrievalError = err
	}
	for i := range pods {
		deployDetails.Pods = append(deployDetails.Pods, ObjectDetails{
			Name:       pods[i].Name,
			Conditions: fromPodCondition(pods[i].Status.Conditions),
			Events:     getObjectEvents(testingT, kubectlOptions, "Pod", pods[i].Name),
			Logs:       getPodLogs(testingT, kubectlOptions, &pods[i]),
		})
	}
	return &deployDetails
}

func ExtractPodDetails(testingT testing.TestingT,
	kubectlOptions *k8s.KubectlOptions, name string,
) *ObjectDetails {
	podObject, err := k8s.GetPodE(testingT, kubectlOptions, name)
	if err != nil {
		// might not be a Pod, let's ignore it
		return &ObjectDetails{RetrievalError: err}
	}
	return &ObjectDetails{
		Name:       podObject.Name,
		Namespace:  kubectlOptions.Namespace,
		Kind:       "Pod",
		Conditions: fromPodCondition(podObject.Status.Conditions),
		Events:     getObjectEvents(testingT, kubectlOptions, "Pod", podObject.Name),
		Phase:      string(podObject.Status.Phase),
		Logs:       getPodLogs(testingT, kubectlOptions, podObject),
	}
}

type ObjectDetails struct {
	RetrievalError error              `json:"retrievalError,omitempty"`
	Kind           string             `json:"kind,omitempty"`
	Namespace      string             `json:"namespace,omitempty"`
	Name           string             `json:"name,omitempty"`
	Phase          string             `json:"phase,omitempty"`
	Logs           map[string]string  `json:"logs,omitempty"`
	Conditions     []*objectCondition `json:"conditions,omitempty"`
	Events         []*simplifiedEvent `json:"events,omitempty"`
	ReplicaSets    []ObjectDetails    `json:"replicaSets,omitempty"`
	Pods           []ObjectDetails    `json:"pods,omitempty"`
}

type objectCondition struct {
	// Type of replica set condition.
	Type string `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type simplifiedEvent struct {
	LastSeen string `json:"lastSeen,omitempty"`
	Type     string `json:"type,omitempty"`
	Object   string `json:"object,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Message  string `json:"message,omitempty"`
}

func getObjectEvents(testingT testing.TestingT, kubectlOptions *k8s.KubectlOptions, kind string, name string) []*simplifiedEvent {
	events, _ := k8s.ListEventsE(testingT, kubectlOptions, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.kind=%s,involvedObject.name=%s", kind, name),
	})
	return simplifyK8sEvents(events)
}

func getPodLogs(testingT testing.TestingT, kubectlOptions *k8s.KubectlOptions, pod *v1.Pod) map[string]string {
	var allLogs map[string]string
	allContainers := append([]v1.ContainerStatus{}, pod.Status.InitContainerStatuses...)
	allContainers = append(allContainers, pod.Status.ContainerStatuses...)
	for _, c := range allContainers {
		logs, err := k8s.GetPodLogsE(testingT, kubectlOptions, pod, c.Name)
		if err != nil {
			continue
		}

		if allLogs == nil {
			allLogs = make(map[string]string)
		}
		allLogs[c.Name] = logs
	}

	return allLogs
}

func fromDeploymentCondition(deploymentConditions []appsv1.DeploymentCondition) []*objectCondition {
	objectConditions := make([]*objectCondition, len(deploymentConditions))
	for i, condition := range deploymentConditions {
		objectConditions[i] = &objectCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			LastTransitionTime: condition.LastTransitionTime.String(),
			Reason:             condition.Reason,
			Message:            condition.Message,
		}
	}
	return objectConditions
}

func fromReplicaSetCondition(replicaSetConditions []appsv1.ReplicaSetCondition) []*objectCondition {
	objectConditions := make([]*objectCondition, len(replicaSetConditions))
	for i, condition := range replicaSetConditions {
		objectConditions[i] = &objectCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			LastTransitionTime: condition.LastTransitionTime.String(),
			Reason:             condition.Reason,
			Message:            condition.Message,
		}
	}
	return objectConditions
}

func fromPodCondition(replicaSetConditions []v1.PodCondition) []*objectCondition {
	objectConditions := make([]*objectCondition, len(replicaSetConditions))
	for i, condition := range replicaSetConditions {
		objectConditions[i] = &objectCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			LastTransitionTime: condition.LastTransitionTime.String(),
			Reason:             condition.Reason,
			Message:            condition.Message,
		}
	}
	return objectConditions
}

func simplifyK8sEvents(v1Events []v1.Event) []*simplifiedEvent {
	if v1Events == nil {
		return nil
	}

	simplifiedEvents := make([]*simplifiedEvent, len(v1Events))
	for i, v1Event := range v1Events {
		simplifiedEvents[i] = simplifySingleK8sEvent(v1Event)
	}
	return simplifiedEvents
}

func simplifySingleK8sEvent(v1Event v1.Event) *simplifiedEvent {
	var lastSeen time.Time
	if !v1Event.LastTimestamp.IsZero() {
		lastSeen = v1Event.LastTimestamp.Time
	} else {
		lastSeen = v1Event.ObjectMeta.CreationTimestamp.Time
	}

	return &simplifiedEvent{
		LastSeen: lastSeen.String(),
		Type:     v1Event.Type,
		Object:   fmt.Sprintf("%s/%s", v1Event.InvolvedObject.Kind, v1Event.InvolvedObject.Name),
		Reason:   v1Event.Reason,
		Message:  v1Event.Message,
	}
}

func RestartCount(pods []v1.Pod) int {
	restartCount := 0
	for _, pod := range pods {
		for _, container := range pod.Status.ContainerStatuses {
			restartCount += int(container.RestartCount)
		}
	}
	return restartCount
}

func ScaleApp(cluster Cluster, app, namespace string, replicas int) error {
	if err := k8s.RunKubectlE(
		cluster.GetTesting(),
		cluster.GetKubectlOptions(namespace),
		"scale", "deployment", app, "--replicas", strconv.Itoa(replicas),
	); err != nil {
		return errors.Wrap(err, "could not scale")
	}
	if err := WaitNumPods(namespace, replicas, app)(cluster); err != nil {
		return errors.Wrap(err, "could not wait until app is scaled")
	}
	return nil
}
