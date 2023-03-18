package framework

import (
	"bytes"
	"fmt"
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
		"apply", "-f", "https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.6.0/experimental-install.yaml")
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
