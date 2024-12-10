package xds

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/shopspring/decimal"

	"github.com/kumahq/kuma/api/common/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type Configurer struct {
	Conf api.Conf
}

func NewConfigurer(conf api.Conf) *Configurer {
	return &Configurer{conf}
}

func (c *Configurer) ConfigureCluster(cluster *envoy_cluster.Cluster) error {
	configureCircuitBreakers(cluster, c.Conf.ConnectionLimits)
	return configureOutlierDetection(cluster, c.Conf.OutlierDetection)
}

func configureCircuitBreakers(cluster *envoy_cluster.Cluster, conf *api.ConnectionLimits) {
	if conf == nil {
		return
	}

	defaultThreshold := &envoy_cluster.CircuitBreakers_Thresholds{
		Priority:       envoy_config_core_v3.RoutingPriority_DEFAULT,
		TrackRemaining: true,
	}

	if conf.MaxConnectionPools != nil {
		defaultThreshold.MaxConnectionPools = util_proto.UInt32(*conf.MaxConnectionPools)
	}

	if conf.MaxConnections != nil {
		defaultThreshold.MaxConnections = util_proto.UInt32(*conf.MaxConnections)
	}

	if conf.MaxPendingRequests != nil {
		defaultThreshold.MaxPendingRequests = util_proto.UInt32(*conf.MaxPendingRequests)
	}

	if conf.MaxRetries != nil {
		defaultThreshold.MaxRetries = util_proto.UInt32(*conf.MaxRetries)
	}

	if conf.MaxRequests != nil {
		defaultThreshold.MaxRequests = util_proto.UInt32(*conf.MaxRequests)
	}

	cluster.CircuitBreakers = &envoy_cluster.CircuitBreakers{
		Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{defaultThreshold},
	}
}

func configureOutlierDetection(cluster *envoy_cluster.Cluster, conf *api.OutlierDetection) error {
	if conf == nil || pointer.Deref(conf.Disabled) {
		return nil
	}

	cluster.OutlierDetection = &envoy_cluster.OutlierDetection{}

	interval := conf.Interval
	baseEjectionTime := conf.BaseEjectionTime
	splitExternalAndLocalErrors := conf.SplitExternalAndLocalErrors
	maxEjectionPercent := conf.MaxEjectionPercent

	if interval != nil {
		cluster.OutlierDetection.Interval = util_proto.Duration(interval.Duration)
	}

	if baseEjectionTime != nil {
		cluster.OutlierDetection.BaseEjectionTime = util_proto.Duration(baseEjectionTime.Duration)
	}

	if splitExternalAndLocalErrors != nil {
		cluster.OutlierDetection.SplitExternalLocalOriginErrors = *splitExternalAndLocalErrors
	}

	if maxEjectionPercent != nil {
		cluster.OutlierDetection.MaxEjectionPercent = util_proto.UInt32(*maxEjectionPercent)
	}

	if err := configureDetectors(cluster.OutlierDetection, conf.Detectors); err != nil {
		return err
	}
	return nil
}

func configureDetectors(outlierDetection *envoy_cluster.OutlierDetection, detectors *api.Detectors) error {
	if detectors == nil {
		configureTotalFailures(outlierDetection, nil)
		configureLocalOriginFailures(outlierDetection, nil)
		configureGatewayFailures(outlierDetection, nil)
		return nil
	}

	configureTotalFailures(outlierDetection, detectors.TotalFailures)
	configureLocalOriginFailures(outlierDetection, detectors.LocalOriginFailures)
	configureGatewayFailures(outlierDetection, detectors.GatewayFailures)
	if err := configureSuccessRate(outlierDetection, detectors.SuccessRate); err != nil {
		return err
	}
	configureFailurePercentage(outlierDetection, detectors.FailurePercentage)
	return nil
}

func configureFailurePercentage(
	outlierDetection *envoy_cluster.OutlierDetection,
	conf *api.DetectorFailurePercentageFailures,
) {
	if conf == nil {
		outlierDetection.EnforcingFailurePercentage = util_proto.UInt32(0)
		// takes effect only when SplitExternalLocalOriginErrors is true
		outlierDetection.EnforcingFailurePercentageLocalOrigin = util_proto.UInt32(0)

		return
	}

	outlierDetection.EnforcingFailurePercentage = util_proto.UInt32(100)
	// takes effect only when SplitExternalLocalOriginErrors is true
	outlierDetection.EnforcingFailurePercentageLocalOrigin = util_proto.UInt32(100)

	if conf.MinimumHosts != nil {
		outlierDetection.FailurePercentageMinimumHosts = util_proto.UInt32(*conf.MinimumHosts)
	}

	if conf.RequestVolume != nil {
		outlierDetection.FailurePercentageRequestVolume = util_proto.UInt32(*conf.RequestVolume)
	}

	if conf.Threshold != nil {
		outlierDetection.FailurePercentageThreshold = util_proto.UInt32(*conf.Threshold)
	}
}

func configureSuccessRate(outlierDetection *envoy_cluster.OutlierDetection, conf *api.DetectorSuccessRateFailures) error {
	if conf == nil {
		outlierDetection.EnforcingSuccessRate = util_proto.UInt32(0)
		// outlierDetection.EnforcingLocalOriginSuccessRate takes effect only
		// when SplitExternalLocalOriginErrors is true
		outlierDetection.EnforcingLocalOriginSuccessRate = util_proto.UInt32(0)
		return nil
	}

	outlierDetection.EnforcingSuccessRate = util_proto.UInt32(100)
	// outlierDetection.EnforcingLocalOriginSuccessRate takes effect only when
	// SplitExternalLocalOriginErrors is true
	outlierDetection.EnforcingLocalOriginSuccessRate = util_proto.UInt32(100)

	if conf.MinimumHosts != nil {
		outlierDetection.SuccessRateMinimumHosts = util_proto.UInt32(*conf.MinimumHosts)
	}

	if conf.RequestVolume != nil {
		outlierDetection.SuccessRateRequestVolume = util_proto.UInt32(*conf.RequestVolume)
	}

	if conf.StandardDeviationFactor != nil {
		dec, err := v1alpha1.NewDecimalFromIntOrString(*conf.StandardDeviationFactor)
		if err != nil {
			return err
		}
		outlierDetection.SuccessRateStdevFactor = util_proto.UInt32(uint32(dec.Mul(decimal.NewFromInt(1000)).IntPart()))
	}
	return nil
}

func configureGatewayFailures(
	outlierDetection *envoy_cluster.OutlierDetection,
	gatewayFailures *api.DetectorGatewayFailures,
) {
	if gatewayFailures == nil || gatewayFailures.Consecutive == nil {
		outlierDetection.EnforcingConsecutiveGatewayFailure = util_proto.UInt32(0)
		return
	}

	outlierDetection.EnforcingConsecutiveGatewayFailure = util_proto.UInt32(100)
	outlierDetection.ConsecutiveGatewayFailure = util_proto.UInt32(*gatewayFailures.Consecutive)
}

func configureLocalOriginFailures(
	outlierDetection *envoy_cluster.OutlierDetection,
	conf *api.DetectorLocalOriginFailures,
) {
	if conf == nil || conf.Consecutive == nil {
		outlierDetection.EnforcingConsecutiveLocalOriginFailure = util_proto.UInt32(0)
		return
	}

	outlierDetection.EnforcingConsecutiveLocalOriginFailure = util_proto.UInt32(100)
	outlierDetection.ConsecutiveLocalOriginFailure = util_proto.UInt32(*conf.Consecutive)
}

func configureTotalFailures(
	outlierDetection *envoy_cluster.OutlierDetection,
	conf *api.DetectorTotalFailures,
) {
	if conf == nil || conf.Consecutive == nil {
		outlierDetection.EnforcingConsecutive_5Xx = util_proto.UInt32(0)
		return
	}

	outlierDetection.EnforcingConsecutive_5Xx = util_proto.UInt32(100)
	outlierDetection.Consecutive_5Xx = util_proto.UInt32(*conf.Consecutive)
}
