package policies

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshretry"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrace"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission"
)

var nameToModule = map[string]func(){
	"meshaccesslogs":              meshaccesslog.InitPlugin,
	"meshcircuitbreakers":         meshcircuitbreaker.InitPlugin,
	"meshfaultinjections":         meshfaultinjection.InitPlugin,
	"meshhealthchecks":            meshhealthcheck.InitPlugin,
	"meshhttproutes":              meshhttproute.InitPlugin,
	"meshloadbalancingstrategies": meshloadbalancingstrategy.InitPlugin,
	"meshmetrics":                 meshmetric.InitPlugin,
	"meshproxypatches":            meshproxypatch.InitPlugin,
	"meshratelimits":              meshratelimit.InitPlugin,
	"meshretries":                 meshretry.InitPlugin,
	"meshtcproutes":               meshtcproute.InitPlugin,
	"meshtimeouts":                meshtimeout.InitPlugin,
	"meshtraces":                  meshtrace.InitPlugin,
	"meshtrafficpermissions":      meshtrafficpermission.InitPlugin,
}

func InitAllPolicies() {
	for _, initializer := range nameToModule {
		initializer()
	}
}

func InitPolicies(enabledPluginPolicies []string) {
	for _, policy := range enabledPluginPolicies {
		initializer, ok := nameToModule[policy]
		if ok {
			initializer()
		} else {
			panic("policy " + policy + " not found")
		}
	}
}
