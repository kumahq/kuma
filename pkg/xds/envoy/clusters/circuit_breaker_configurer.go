package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/api/v2/cluster"
	"github.com/golang/protobuf/ptypes/wrappers"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
)

func OutlierDetection(circuitBreaker *mesh_core.CircuitBreakerResource) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&OutlierDetectionConfigurer{
			CircuitBreaker: circuitBreaker,
		})
	})
}

type OutlierDetectionConfigurer struct {
	CircuitBreaker *mesh_core.CircuitBreakerResource
}

func (c *OutlierDetectionConfigurer) Configure(cluster *envoy_api.Cluster) error {
	if c.CircuitBreaker == nil {
		return nil
	}

	cluster.OutlierDetection = &envoy_cluster.OutlierDetection{
		Interval:                       c.CircuitBreaker.Spec.GetConf().Interval,
		BaseEjectionTime:               c.CircuitBreaker.Spec.GetConf().BaseEjectionTime,
		MaxEjectionPercent:             c.CircuitBreaker.Spec.GetConf().MaxEjectionPercent,
		SplitExternalLocalOriginErrors: c.CircuitBreaker.Spec.GetConf().SplitExternalAndLocalErrors,
	}
	c.configureTotalErrorDetector(cluster.OutlierDetection)
	c.configureGatewayErrorDetector(cluster.OutlierDetection)
	c.configureLocalErrorDetector(cluster.OutlierDetection)
	c.configureStandardDeviationDetector(cluster.OutlierDetection)
	c.configureFailureDetector(cluster.OutlierDetection)
	return nil
}

func (c *OutlierDetectionConfigurer) configureTotalErrorDetector(outlierDetection *envoy_cluster.OutlierDetection) {
	if total := c.CircuitBreaker.Spec.GetConf().GetDetectors().GetTotalErrors(); total != nil {
		outlierDetection.Consecutive_5Xx = total.GetConsecutive()
		outlierDetection.EnforcingConsecutive_5Xx = &wrappers.UInt32Value{Value: 100}
	} else {
		outlierDetection.EnforcingConsecutive_5Xx = &wrappers.UInt32Value{Value: 0}
	}
}

func (c *OutlierDetectionConfigurer) configureGatewayErrorDetector(outlierDetection *envoy_cluster.OutlierDetection) {
	if gateway := c.CircuitBreaker.Spec.GetConf().GetDetectors().GetGatewayErrors(); gateway != nil {
		outlierDetection.ConsecutiveGatewayFailure = gateway.GetConsecutive()
		outlierDetection.EnforcingConsecutiveGatewayFailure = &wrappers.UInt32Value{Value: 100}
	} else {
		outlierDetection.EnforcingConsecutiveGatewayFailure = &wrappers.UInt32Value{Value: 0}
	}
}

func (c *OutlierDetectionConfigurer) configureLocalErrorDetector(outlierDetection *envoy_cluster.OutlierDetection) {
	if local := c.CircuitBreaker.Spec.GetConf().GetDetectors().GetLocalErrors(); local != nil {
		outlierDetection.ConsecutiveLocalOriginFailure = local.GetConsecutive()
		outlierDetection.EnforcingConsecutiveLocalOriginFailure = &wrappers.UInt32Value{Value: 100}
	} else {
		outlierDetection.EnforcingConsecutiveLocalOriginFailure = &wrappers.UInt32Value{Value: 0}
	}
}

func (c *OutlierDetectionConfigurer) configureStandardDeviationDetector(outlierDetection *envoy_cluster.OutlierDetection) {
	if stdev := c.CircuitBreaker.Spec.GetConf().GetDetectors().GetStandardDeviation(); stdev != nil {
		outlierDetection.SuccessRateRequestVolume = stdev.GetRequestVolume()
		outlierDetection.SuccessRateMinimumHosts = stdev.GetMinimumHosts()
		if factor := stdev.GetFactor(); factor != nil {
			outlierDetection.SuccessRateStdevFactor = &wrappers.UInt32Value{Value: uint32(factor.GetValue() * 1000)}
		}
		outlierDetection.EnforcingSuccessRate = &wrappers.UInt32Value{Value: 100}
		outlierDetection.EnforcingLocalOriginSuccessRate = &wrappers.UInt32Value{Value: 100} // takes effect only when SplitExternalLocalOriginErrors is true
	} else {
		outlierDetection.EnforcingSuccessRate = &wrappers.UInt32Value{Value: 0}
	}
}

func (c *OutlierDetectionConfigurer) configureFailureDetector(outlierDetection *envoy_cluster.OutlierDetection) {
	if failure := c.CircuitBreaker.Spec.GetConf().GetDetectors().GetFailure(); failure != nil {
		outlierDetection.FailurePercentageRequestVolume = failure.GetRequestVolume()
		outlierDetection.FailurePercentageMinimumHosts = failure.GetMinimumHosts()
		outlierDetection.FailurePercentageThreshold = failure.GetThreshold()

		outlierDetection.EnforcingFailurePercentage = &wrappers.UInt32Value{Value: 100}
		outlierDetection.EnforcingFailurePercentageLocalOrigin = &wrappers.UInt32Value{Value: 100} // takes effect only when SplitExternalLocalOriginErrors is true
	} else {
		outlierDetection.EnforcingFailurePercentage = &wrappers.UInt32Value{Value: 0}
	}
}
