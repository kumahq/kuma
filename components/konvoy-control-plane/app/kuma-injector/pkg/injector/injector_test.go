package injector_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	inject "github.com/Kong/konvoy/components/konvoy-control-plane/app/kuma-injector/pkg/injector"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	conf "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/kuma-injector"

	"github.com/ghodss/yaml"
	kube_core "k8s.io/api/core/v1"
)

var _ = Describe("Injector", func() {

	var injector *inject.KumaInjector

	BeforeEach(func() {
		var cfg conf.Injector
		Expect(config.Load(filepath.Join("testdata", "inject.config.yaml"), &cfg)).To(Succeed())
		injector = inject.New(cfg)
	})

	type testCase struct {
		num string
	}

	DescribeTable("should inject Kuma into a Pod",
		func(given testCase) {
			// setup
			inputFile := filepath.Join("testdata", fmt.Sprintf("inject.%s.input.yaml", given.num))
			goldenFile := filepath.Join("testdata", fmt.Sprintf("inject.%s.golden.yaml", given.num))

			// given
			pod := &kube_core.Pod{}

			By("loading input Pod")
			// when
			input, err := ioutil.ReadFile(inputFile)
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			err = yaml.Unmarshal(input, pod)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("injecting Kuma")
			// when
			err = injector.InjectKuma(pod)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("loading golden Pod")
			// when
			actual, err := yaml.Marshal(pod)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("comparing actual against golden")
			// when
			expected, err := ioutil.ReadFile(goldenFile)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(expected))
		},
		Entry("01. Pod without init containers and annotations", testCase{
			num: "01",
		}),
		Entry("02. Pod with init containers and annotations", testCase{
			num: "02",
		}),
		Entry("03. Pod without Namespace and Name", testCase{
			num: "03",
		}),
		Entry("04. Pod with explicitly selected Mesh", testCase{
			num: "04",
		}),
	)
})
