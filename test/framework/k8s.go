package framework

import (
	"bytes"
	"encoding/json"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"

	bootstrap_k8s "github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s"
)

func PodNameOfApp(cluster Cluster, name string, namespace string) (string, error) {
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
	return pods[0].Name, nil
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
		"apply", "-f", "https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/experimental-install.yaml")
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

type K8sObjectDetailPrinter struct {
	testingT        testing.TestingT
	kubectlOptions  *k8s.KubectlOptions
	objectNamespace string
	objectName      string

	isDeployKind bool
	isPodKind    bool
	podObject    *v1.Pod
}

func NewK8sDeploymentDetailPrinter(testingT testing.TestingT,
	kubectlOptions *k8s.KubectlOptions, namespace string, name string) *K8sObjectDetailPrinter {
	return &K8sObjectDetailPrinter{
		testingT:        testingT,
		kubectlOptions:  kubectlOptions,
		objectNamespace: namespace,
		objectName:      name,
		isDeployKind:    true,
	}
}
func NewK8sPodDetailPrinter(testingT testing.TestingT,
	kubectlOptions *k8s.KubectlOptions, podObject *v1.Pod) *K8sObjectDetailPrinter {
	return &K8sObjectDetailPrinter{
		testingT:        testingT,
		kubectlOptions:  kubectlOptions,
		objectNamespace: podObject.Namespace,
		objectName:      podObject.Name,
		podObject:       podObject,
		isPodKind:       true,
	}
}

func (p *K8sObjectDetailPrinter) Print() string {
	if p.isDeployKind {
		return p.printDeploymentDetails()
	}

	if p.isPodKind {
		return p.printPodDetails()
	}

	return ""
}

func (p *K8sObjectDetailPrinter) printDeploymentDetails() string {
	deploy, err := k8s.GetDeploymentE(p.testingT, p.kubectlOptions, p.objectName)
	if err != nil {
		// might not be a Deployment, let's ignore it
		return ""
	}

	deployDetails := deploymentDetails{Namespace: p.objectNamespace,
		Kind: "Deployment",
		objectDetails: newObjectDetails(p.objectName,
			fromDeploymentCondition(deploy.Status.Conditions),
			getObjectEvents(p.testingT, p.kubectlOptions, "Deployment", p.objectName)),
	}

	replicaSets, _ := k8s.ListReplicaSetsE(p.testingT, p.kubectlOptions, metav1.ListOptions{
		LabelSelector: "app=" + p.objectName,
	})
	for _, ars := range replicaSets {
		rsDetails := newObjectDetails(ars.Name,
			fromReplicaSetCondition(ars.Status.Conditions),
			getObjectEvents(p.testingT, p.kubectlOptions, "ReplicaSet", ars.Name))
		deployDetails.ReplicaSets = append(deployDetails.ReplicaSets, rsDetails)
	}

	pods, _ := k8s.ListPodsE(p.testingT, p.kubectlOptions, metav1.ListOptions{
		LabelSelector: "app=" + p.objectName,
	})
	for _, pod := range pods {
		pDetail := &podDetails{
			objectDetails: newObjectDetails(pod.Name,
				fromPodCondition(pod.Status.Conditions),
				getObjectEvents(p.testingT, p.kubectlOptions, "Pod", pod.Name)),
		}
		deployDetails.Pods = append(deployDetails.Pods, pDetail)
	}

	deployDetailsJson, err := json.Marshal(deployDetails)
	if err != nil {
		deployDetailsJson = []byte(fmt.Sprintf("error marshaling deployment details: '%s'", err.Error()))
	}
	return string(deployDetailsJson)
}

func (p *K8sObjectDetailPrinter) printPodDetails() string {
	pDetails := &podDetails{
		objectDetails: newObjectDetails(p.objectName,
			fromPodCondition(p.podObject.Status.Conditions),
			getObjectEvents(p.testingT, p.kubectlOptions, "Pod", p.objectName)),
	}
	podDetailsJson, err := json.Marshal(pDetails)
	if err != nil {
		podDetailsJson = []byte(fmt.Sprintf("error marshaling pod details: '%s'", err.Error()))
	}

	return string(podDetailsJson)
}

type objectDetails struct {
	Name       string             `json:"name,omitempty"`
	Conditions []*objectCondition `json:"conditions,omitempty"`
	Events     []*simplifiedEvent `json:"events,omitempty"`
}

type deploymentDetails struct {
	*objectDetails `json:",inline"`
	Kind           string           `json:"kind,omitempty"`
	Namespace      string           `json:"namespace,omitempty"`
	ReplicaSets    []*objectDetails `json:"replicaSets,omitempty"`
	Pods           []*podDetails    `json:"pods,omitempty"`
}

type podDetails struct {
	*objectDetails `json:",inline"`
	Phase          string `json:"phase,omitempty"`
	Logs           string `json:"logs,omitempty"`
}

type objectCondition struct {
	// Type of replica set condition.
	Type string `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type simplifiedEvent struct {
	LastSeen *time.Time `json:"lastSeen,omitempty"`
	Type     string     `json:"type,omitempty"`
	Object   string     `json:"object,omitempty"`
	Reason   string     `json:"reason,omitempty"`
	Message  string     `json:"message,omitempty"`
}

func getObjectEvents(testingT testing.TestingT, kubectlOptions *k8s.KubectlOptions, kind string, name string) []*simplifiedEvent {
	events, _ := k8s.ListEventsE(testingT, kubectlOptions, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.kind=%s,involvedObject.name=%s", kind, name)})
	return simplifyK8sEvents(events)
}

func newObjectDetails(name string, conditions []*objectCondition, events []*simplifiedEvent) *objectDetails {
	return &objectDetails{
		Name:       name,
		Conditions: conditions,
		Events:     events,
	}
}

func fromDeploymentCondition(deploymentConditions []appsv1.DeploymentCondition) []*objectCondition {
	objectConditions := make([]*objectCondition, len(deploymentConditions))
	for i, condition := range deploymentConditions {
		objectConditions[i] = &objectCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			LastTransitionTime: condition.LastTransitionTime,
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
			LastTransitionTime: condition.LastTransitionTime,
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
			LastTransitionTime: condition.LastTransitionTime,
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
		simplifiedEvents[i] = simplifySingleK8sEvent(&v1Event)
	}
	return simplifiedEvents
}

func simplifySingleK8sEvent(v1Event *v1.Event) *simplifiedEvent {
	var lastSeen *time.Time
	if !v1Event.LastTimestamp.IsZero() {
		lastSeen = &v1Event.LastTimestamp.Time
	} else {
		lastSeen = &v1Event.ObjectMeta.CreationTimestamp.Time
	}

	return &simplifiedEvent{
		LastSeen: lastSeen,
		Type:     v1Event.Type,
		Object:   fmt.Sprintf("%s/%s", v1Event.InvolvedObject.Kind, v1Event.InvolvedObject.Name),
		Reason:   v1Event.Reason,
		Message:  v1Event.Message,
	}
}
