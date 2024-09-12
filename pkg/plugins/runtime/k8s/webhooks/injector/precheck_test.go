package injector

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

var _ = Describe("annotation deprecation", func() {
	type testCase struct {
		annotationKey   string
		annotationValue string

		expectedKeyDeprecated         bool
		expectedValueDeprecated       bool
		expectedKeyDeprecationMessage string
		expectedValueReplacement      string
	}

	i := &KumaInjector{
		client: dummyClient{},
	}

	DescribeTable("yes/no value deprecation",
		func(given testCase) {
			logSink := &fakeLogSink{root: &fakeLogSinkRoot{}}

			pod := createPod(given.annotationKey, given.annotationValue)
			_, err := i.preCheck(context.TODO(), pod, logr.New(logSink))

			Expect(err).ToNot(HaveOccurred())
			Expect(logSink.root.messages).To(ContainElement(logInfo{name: nil, tags: nil, msg: "skipping Kuma injection"}))

			if given.expectedValueDeprecated {
				message := fmt.Sprintf("WARNING: using '%s' for annotation '%s' is deprecated, please use '%s' instead",
					given.annotationValue, given.annotationKey, given.expectedValueReplacement)
				Expect(logSink.root.messages).To(ContainElement(
					logInfo{name: nil, tags: nil, msg: message},
				))
			} else {
				deprecateMessage := slices.IndexFunc(logSink.root.exportMessages(), func(logMsg string) bool {
					return strings.Contains(logMsg, fmt.Sprintf("for annotation '%s' is deprecated, please use", given.annotationKey))
				})
				Expect(deprecateMessage).To(Equal(-1))
			}
		},
		Entry("kuma.io/virtual-probes - yes", testCase{
			annotationKey:            metadata.KumaVirtualProbesAnnotation,
			annotationValue:          "yes",
			expectedValueDeprecated:  true,
			expectedValueReplacement: "true",
		}),
		Entry("kuma.io/sidecar-injection - yes", testCase{
			annotationKey:            metadata.KumaSidecarInjectionAnnotation,
			annotationValue:          "yes",
			expectedValueDeprecated:  true,
			expectedValueReplacement: "true",
		}),
		Entry("kuma.io/sidecar-injection - no", testCase{
			annotationKey:            metadata.KumaSidecarInjectionAnnotation,
			annotationValue:          "no",
			expectedValueDeprecated:  true,
			expectedValueReplacement: "false",
		}),
		Entry("kuma.io/sidecar-injection - enabled", testCase{
			annotationKey:           metadata.KumaSidecarInjectionAnnotation,
			annotationValue:         "enabled",
			expectedKeyDeprecated:   false,
			expectedValueDeprecated: false,
		}),
		Entry("kuma.io/gateway - enabled", testCase{
			annotationKey:           metadata.KumaGatewayAnnotation,
			annotationValue:         "enabled",
			expectedKeyDeprecated:   false,
			expectedValueDeprecated: false,
		}),
	)

	DescribeTable("annotation key deprecation",
		func(given testCase) {
			logSink := &fakeLogSink{root: &fakeLogSinkRoot{}}

			pod := createPod(given.annotationKey, given.annotationValue)
			_, err := i.preCheck(context.TODO(), pod, logr.New(logSink))

			Expect(err).ToNot(HaveOccurred())
			Expect(logSink.root.messages).To(ContainElement(logInfo{name: nil, tags: nil, msg: "skipping Kuma injection"}))

			if given.expectedKeyDeprecated {
				Expect(logSink.root.messages).To(ContainElement(
					logInfo{
						name: nil, tags: []interface{}{"key", given.annotationKey, "message", given.expectedKeyDeprecationMessage},
						msg: "WARNING: using deprecated pod annotation",
					},
				))
			}
		},
		Entry("kuma.io/virtual-probes - deprecated", testCase{
			annotationKey:                 metadata.KumaVirtualProbesAnnotation,
			annotationValue:               "enabled",
			expectedKeyDeprecated:         true,
			expectedKeyDeprecationMessage: fmt.Sprintf("'%s' will be removed in a future release", metadata.KumaVirtualProbesAnnotation),
		}),
		Entry("kuma.io/sidecar-injection - not deprecated", testCase{
			annotationKey:         metadata.KumaSidecarInjectionAnnotation,
			annotationValue:       "enabled",
			expectedKeyDeprecated: false,
		}),
	)
})

func createPod(annotationName, value string) *kube_core.Pod {
	return &kube_core.Pod{
		ObjectMeta: kube_meta.ObjectMeta{
			Name:   "test-pod",
			Labels: map[string]string{},
			Annotations: map[string]string{
				annotationName: value,
			},
		},
		Spec: kube_core.PodSpec{
			Containers: []kube_core.Container{
				{Name: "foo", Image: "busybox"},
			},
		},
	}
}

// taken from https://github.com/kubernetes-sigs/controller-runtime/blob/release-0.19/pkg/log/log_test.go#L31

// logInfo is the information for a particular fakeLogSink message.
type logInfo struct {
	name []string
	tags []interface{}
	msg  string
}

// fakeLogSinkRoot is the root object to which all fakeLoggers record their messages.
type fakeLogSinkRoot struct {
	messages []logInfo
}

func (r *fakeLogSinkRoot) exportMessages() []string {
	var result []string
	for _, m := range r.messages {
		result = append(result, m.msg)
	}
	return result
}

// fakeLogSink is a fake implementation of logr.Logger that records
// messages, tags, and names,
// just records the name.
type fakeLogSink struct {
	name []string
	tags []interface{}

	root *fakeLogSinkRoot
}

func (f *fakeLogSink) Init(info logr.RuntimeInfo) {
}

func (f *fakeLogSink) WithName(name string) logr.LogSink {
	names := append([]string(nil), f.name...)
	names = append(names, name)
	return &fakeLogSink{
		name: names,
		tags: f.tags,
		root: f.root,
	}
}

func (f *fakeLogSink) WithValues(vals ...interface{}) logr.LogSink {
	tags := append([]interface{}(nil), f.tags...)
	tags = append(tags, vals...)
	return &fakeLogSink{
		name: f.name,
		tags: tags,
		root: f.root,
	}
}

func (f *fakeLogSink) Error(err error, msg string, vals ...interface{}) {
	tags := append([]interface{}(nil), f.tags...)
	tags = append(tags, "error", err)
	tags = append(tags, vals...)
	f.root.messages = append(f.root.messages, logInfo{
		name: append([]string(nil), f.name...),
		tags: tags,
		msg:  msg,
	})
}

func (f *fakeLogSink) Info(level int, msg string, vals ...interface{}) {
	tags := append([]interface{}(nil), f.tags...)
	tags = append(tags, vals...)
	f.root.messages = append(f.root.messages, logInfo{
		name: append([]string(nil), f.name...),
		tags: tags,
		msg:  msg,
	})
}

func (f *fakeLogSink) Enabled(level int) bool { return true }

// Taken from https://github.com/kubernetes-sigs/controller-runtime/blob/release-0.19/pkg/client/interceptor/intercept_test.go#L335

type dummyClient struct{}

var _ kube_client.WithWatch = &dummyClient{}

func (d dummyClient) Get(ctx context.Context, key kube_client.ObjectKey, obj kube_client.Object, opts ...kube_client.GetOption) error {
	return nil
}

func (d dummyClient) List(ctx context.Context, list kube_client.ObjectList, opts ...kube_client.ListOption) error {
	return nil
}

func (d dummyClient) Create(ctx context.Context, obj kube_client.Object, opts ...kube_client.CreateOption) error {
	return nil
}

func (d dummyClient) Delete(ctx context.Context, obj kube_client.Object, opts ...kube_client.DeleteOption) error {
	return nil
}

func (d dummyClient) Update(ctx context.Context, obj kube_client.Object, opts ...kube_client.UpdateOption) error {
	return nil
}

func (d dummyClient) Patch(ctx context.Context, obj kube_client.Object, patch kube_client.Patch, opts ...kube_client.PatchOption) error {
	return nil
}

func (d dummyClient) DeleteAllOf(ctx context.Context, obj kube_client.Object, opts ...kube_client.DeleteAllOfOption) error {
	return nil
}

func (d dummyClient) Status() kube_client.SubResourceWriter {
	return d.SubResource("status")
}

func (d dummyClient) SubResource(subResource string) kube_client.SubResourceClient {
	return nil
}

func (d dummyClient) Scheme() *runtime.Scheme {
	return nil
}

func (d dummyClient) RESTMapper() meta.RESTMapper {
	return nil
}

func (d dummyClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, nil
}

func (d dummyClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return false, nil
}

func (d dummyClient) Watch(ctx context.Context, obj kube_client.ObjectList, opts ...kube_client.ListOption) (watch.Interface, error) {
	return nil, nil
}
