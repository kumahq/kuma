package v1alpha1

import (
	"fmt"

	"github.com/kumahq/kuma/pkg/plugins/policies/core/jsonpatch/validators"
)

func (t *MeshLoadBalancingStrategyResource) Deprecations() []string {
	deprecations := validateHashPoliciesType(t.Spec.To)
	deprecations = append(deprecations, validators.TopLevelTargetRefDeprecations(t.Spec.TargetRef)...)
	return deprecations
}

func validateHashPoliciesType(confs []To) []string {
	deprecations := []string{}
	for ruleIdx, conf := range confs {
		if conf.Default.LoadBalancer == nil {
			continue
		}
	
		switch conf.Default.LoadBalancer.Type {
		case RingHashType:
			if conf.Default.LoadBalancer.RingHash == nil || conf.Default.LoadBalancer.RingHash.HashPolicies == nil {
				continue
			}
			for lbIdx, lbConf := range *conf.Default.LoadBalancer.RingHash.HashPolicies {
				if lbConf.Type == SourceIPType {
					deprecations = append(deprecations, fmt.Sprintf("%s type for 'spec.to[%d].default.loadBalancer.ringHash.hashPolicies[%d].type' is deprecated, use %s instead", SourceIPType, ruleIdx, lbIdx, ConnectionType))
				}
			}
		case MaglevType:
			if conf.Default.LoadBalancer.Maglev == nil || conf.Default.LoadBalancer.Maglev.HashPolicies == nil {
				continue
			}
			for lbIdx, lbConf := range *conf.Default.LoadBalancer.Maglev.HashPolicies {
				if lbConf.Type == SourceIPType {
					deprecations = append(deprecations, fmt.Sprintf("%s type for 'spec.to[%d].default.loadBalancer.maglev.hashPolicies[%d].type' is deprecated, use %s instead", SourceIPType, ruleIdx, lbIdx, ConnectionType))
				}
			}
		}
	}
	return deprecations
}
