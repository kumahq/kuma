package merge_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	jsonpatch "github.com/evanphx/json-patch/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/merge"
	meshaccesslog_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshcircuitbreaker_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func baselineMerge[T any](confs []T) T {
	GinkgoHelper()
	var resultBytes []byte
	for _, conf := range confs {
		confBytes, err := json.Marshal(conf)
		Expect(err).ToNot(HaveOccurred())
		if len(resultBytes) == 0 {
			resultBytes = confBytes
			continue
		}
		resultBytes, err = jsonpatch.MergePatch(resultBytes, confBytes)
		Expect(err).ToNot(HaveOccurred())
	}
	var result T
	Expect(json.Unmarshal(resultBytes, &result)).To(Succeed())
	return result
}

func runEquivalence[T any](r *rand.Rand, n int, randomConf func(*rand.Rand) T) {
	GinkgoHelper()
	confs := []T{}
	anyConfs := []any{}
	for range n {
		conf := randomConf(r)
		confs = append(confs, conf)
		anyConfs = append(anyConfs, conf)
	}

	merged, err := merge.Confs(anyConfs)
	Expect(err).ToNot(HaveOccurred())
	Expect(merged).To(HaveLen(1))

	expected := baselineMerge(confs)
	Expect(merged[0]).To(Equal(expected))
}

var _ = Describe("mergeConfs equivalence with jsonpatch.MergePatch", func() {
	randomTimeoutConf := func(r *rand.Rand) meshtimeout_api.Conf {
		conf := meshtimeout_api.Conf{}
		if r.Intn(2) == 0 {
			conf.ConnectionTimeout = pointer.To(k8s.Duration{Duration: time.Duration(r.Intn(100)) * time.Second})
		}
		if r.Intn(2) == 0 {
			conf.IdleTimeout = pointer.To(k8s.Duration{Duration: time.Duration(r.Intn(100)) * time.Second})
		}
		if r.Intn(2) == 0 {
			http := &meshtimeout_api.Http{}
			if r.Intn(2) == 0 {
				http.RequestTimeout = pointer.To(k8s.Duration{Duration: time.Duration(r.Intn(100)) * time.Second})
			}
			if r.Intn(2) == 0 {
				http.StreamIdleTimeout = pointer.To(k8s.Duration{Duration: time.Duration(r.Intn(100)) * time.Second})
			}
			if r.Intn(4) == 0 {
				http.MaxStreamDuration = pointer.To(k8s.Duration{Duration: time.Duration(r.Intn(100)) * time.Second})
			}
			conf.Http = http
		}
		return conf
	}

	randomCircuitBreakerConf := func(r *rand.Rand) meshcircuitbreaker_api.Conf {
		conf := meshcircuitbreaker_api.Conf{}
		if r.Intn(2) == 0 {
			conf.ConnectionLimits = &meshcircuitbreaker_api.ConnectionLimits{
				MaxConnections: pointer.To(uint32(r.Intn(1000))),
			}
		}
		if r.Intn(2) == 0 {
			outlier := &meshcircuitbreaker_api.OutlierDetection{
				Disabled:           pointer.To(r.Intn(2) == 0),
				Interval:           pointer.To(k8s.Duration{Duration: time.Duration(r.Intn(100)) * time.Second}),
				MaxEjectionPercent: pointer.To(uint32(r.Intn(100))),
			}
			if r.Intn(2) == 0 {
				outlier.Detectors = &meshcircuitbreaker_api.Detectors{
					SuccessRate: &meshcircuitbreaker_api.DetectorSuccessRateFailures{
						MinimumHosts:            pointer.To(uint32(r.Intn(10))),
						StandardDeviationFactor: pointer.To(intstr.FromInt32(int32(r.Intn(200)))),
					},
				}
			}
			conf.OutlierDetection = outlier
		}
		return conf
	}

	randomAccessLogConf := func(r *rand.Rand) meshaccesslog_api.Conf {
		conf := meshaccesslog_api.Conf{}
		if r.Intn(2) == 0 {
			backend := meshaccesslog_api.Backend{
				Type: meshaccesslog_api.OtelTelemetryBackendType,
				OpenTelemetry: &meshaccesslog_api.OtelBackend{
					Attributes: &[]meshaccesslog_api.OtelAttribute{
						{Key: "mesh", Value: fmt.Sprintf("mesh-%d", r.Intn(10))},
					},
				},
			}
			if r.Intn(2) == 0 {
				body := fmt.Sprintf(`{"kvlistValue":{"values":[{"key":"k%d","value":{"stringValue":"v%d"}}]}}`, r.Intn(5), r.Intn(5))
				if r.Intn(3) == 0 {
					body = `{"kvlistValue":{"values":[{"key":null}]}}`
				}
				backend.OpenTelemetry.Body = &apiextensionsv1.JSON{
					Raw: []byte(body),
				}
			}
			conf.Backends = &[]meshaccesslog_api.Backend{backend}
		}
		return conf
	}

	for _, seed := range []int64{1, 2, 3, 4, 5} {
		for _, n := range []int{2, 3, 5, 10, 40} {
			seed, n := seed, n
			It(fmt.Sprintf("MeshTimeout should match baseline for seed=%d n=%d", seed, n), func() {
				runEquivalence(rand.New(rand.NewSource(seed)), n, randomTimeoutConf)
			})
			It(fmt.Sprintf("MeshCircuitBreaker should match baseline for seed=%d n=%d", seed, n), func() {
				runEquivalence(rand.New(rand.NewSource(seed)), n, randomCircuitBreakerConf)
			})
			It(fmt.Sprintf("MeshAccessLog should match baseline for seed=%d n=%d", seed, n), func() {
				runEquivalence(rand.New(rand.NewSource(seed)), n, randomAccessLogConf)
			})
		}
	}
})
