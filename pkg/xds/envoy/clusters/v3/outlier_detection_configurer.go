package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type OutlierDetectionConfigurer struct {
	CircuitBreaker *core_mesh.CircuitBreakerResource
}

var _ ClusterConfigurer = &OutlierDetectionConfigurer{}

func (c *OutlierDetectionConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
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
		outlierDetection.EnforcingConsecutive_5Xx = &wrapperspb.UInt32Value{Value: 100}
	} else {
		outlierDetection.EnforcingConsecutive_5Xx = &wrapperspb.UInt32Value{Value: 0}
	}
}

func (c *OutlierDetectionConfigurer) configureGatewayErrorDetector(outlierDetection *envoy_cluster.OutlierDetection) {
	if gateway := c.CircuitBreaker.Spec.GetConf().GetDetectors().GetGatewayErrors(); gateway != nil {
		outlierDetection.ConsecutiveGatewayFailure = gateway.GetConsecutive()
		outlierDetection.EnforcingConsecutiveGatewayFailure = &wrapperspb.UInt32Value{Value: 100}
	} else {
		outlierDetection.EnforcingConsecutiveGatewayFailure = &wrapperspb.UInt32Value{Value: 0}
	}
}

func (c *OutlierDetectionConfigurer) configureLocalErrorDetector(outlierDetection *envoy_cluster.OutlierDetection) {
	if local := c.CircuitBreaker.Spec.GetConf().GetDetectors().GetLocalErrors(); local != nil {
		outlierDetection.ConsecutiveLocalOriginFailure = local.GetConsecutive()
		outlierDetection.EnforcingConsecutiveLocalOriginFailure = &wrapperspb.UInt32Value{Value: 100}
	} else {
		outlierDetection.EnforcingConsecutiveLocalOriginFailure = &wrapperspb.UInt32Value{Value: 0}
	}
}

func (c *OutlierDetectionConfigurer) configureStandardDeviationDetector(outlierDetection *envoy_cluster.OutlierDetection) {
	if stdev := c.CircuitBreaker.Spec.GetConf().GetDetectors().GetStandardDeviation(); stdev != nil {
		outlierDetection.SuccessRateRequestVolume = stdev.GetRequestVolume()
		outlierDetection.SuccessRateMinimumHosts = stdev.GetMinimumHosts()
		if factor := stdev.GetFactor(); factor != nil {
			outlierDetection.SuccessRateStdevFactor = &wrapperspb.UInt32Value{Value: uint32(factor.GetValue() * 1000)}
		}
		outlierDetection.EnforcingSuccessRate = &wrapperspb.UInt32Value{Value: 100}
		outlierDetection.EnforcingLocalOriginSuccessRate = &wrapperspb.UInt32Value{Value: 100} // takes effect only when SplitExternalLocalOriginErrors is true
	} else {
		outlierDetection.EnforcingSuccessRate = &wrapperspb.UInt32Value{Value: 0}
	}
}

func (c *OutlierDetectionConfigurer) configureFailureDetector(outlierDetection *envoy_cluster.OutlierDetection) {
	if failure := c.CircuitBreaker.Spec.GetConf().GetDetectors().GetFailure(); failure != nil {
		outlierDetection.FailurePercentageRequestVolume = failure.GetRequestVolume()
		outlierDetection.FailurePercentageMinimumHosts = failure.GetMinimumHosts()
		outlierDetection.FailurePercentageThreshold = failure.GetThreshold()

		outlierDetection.EnforcingFailurePercentage = &wrapperspb.UInt32Value{Value: 100}
		outlierDetection.EnforcingFailurePercentageLocalOrigin = &wrapperspb.UInt32Value{Value: 100} // takes effect only when SplitExternalLocalOriginErrors is true
	} else {
		outlierDetection.EnforcingFailurePercentage = &wrapperspb.UInt32Value{Value: 0}
	}
}
