/*
Copyright 2019 Kuma authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8s_test

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_registry "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"

	// +kubebuilder:scaffold:imports
	sample_v1alpha1 "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/test/api/sample/v1alpha1"
)

var k8sClient client.Client
var testEnv *envtest.Environment
var k8sClientScheme = runtime.NewScheme()

func TestKubernetes(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Kubernetes Resources Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("native", "test", "config", "crd", "bases"),
			filepath.Join("native", "config", "crd", "bases"),
		},
	}

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = sample_v1alpha1.AddToScheme(k8sClientScheme)
	Expect(err).NotTo(HaveOccurred())

	err = mesh_k8s.AddToScheme(k8sClientScheme)
	Expect(err).NotTo(HaveOccurred())

	err = kube_core.AddToScheme(k8sClientScheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: k8sClientScheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	err = k8s_registry.Global().RegisterObjectType(&v1alpha1.TrafficRoute{}, &sample_v1alpha1.SampleTrafficRoute{
		TypeMeta: kube_meta.TypeMeta{
			APIVersion: sample_v1alpha1.GroupVersion.String(),
			Kind:       "SampleTrafficRoute",
		},
	})
	Expect(err).ToNot(HaveOccurred())
	err = k8s_registry.Global().RegisterListType(&v1alpha1.TrafficRoute{}, &sample_v1alpha1.SampleTrafficRouteList{
		TypeMeta: kube_meta.TypeMeta{
			APIVersion: sample_v1alpha1.GroupVersion.String(),
			Kind:       "SampleTrafficRouteList",
		},
	})
	Expect(err).ToNot(HaveOccurred())

	Expect(k8sClient.Create(context.Background(), &kube_core.Namespace{ObjectMeta: kube_meta.ObjectMeta{Name: "demo"}})).To(Succeed())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
