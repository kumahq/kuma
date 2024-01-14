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

type pluginInitializer struct {
	initFn      func()
	initialized bool
}

var nameToModule = map[string]*pluginInitializer{
	"meshaccesslogs":              {initFn: meshaccesslog.InitPlugin, initialized: false},
	"meshcircuitbreakers":         {initFn: meshcircuitbreaker.InitPlugin, initialized: false},
	"meshfaultinjections":         {initFn: meshfaultinjection.InitPlugin, initialized: false},
	"meshhealthchecks":            {initFn: meshhealthcheck.InitPlugin, initialized: false},
	"meshhttproutes":              {initFn: meshhttproute.InitPlugin, initialized: false},
	"meshloadbalancingstrategies": {initFn: meshloadbalancingstrategy.InitPlugin, initialized: false},
	"meshmetrics":                 {initFn: meshmetric.InitPlugin, initialized: false},
	"meshproxypatches":            {initFn: meshproxypatch.InitPlugin, initialized: false},
	"meshratelimits":              {initFn: meshratelimit.InitPlugin, initialized: false},
	"meshretries":                 {initFn: meshretry.InitPlugin, initialized: false},
	"meshtcproutes":               {initFn: meshtcproute.InitPlugin, initialized: false},
	"meshtimeouts":                {initFn: meshtimeout.InitPlugin, initialized: false},
	"meshtraces":                  {initFn: meshtrace.InitPlugin, initialized: false},
	"meshtrafficpermissions":      {initFn: meshtrafficpermission.InitPlugin, initialized: false},
}

func InitAllPolicies() {
	for _, initializer := range nameToModule {
		if !initializer.initialized {
			initializer.initFn()
			initializer.initialized = true
		}
	}
}

func InitPolicies(enabledPluginPolicies []string) {
	for _, policy := range enabledPluginPolicies {
		initializer, ok := nameToModule[policy]
		if ok && !initializer.initialized {
			initializer.initFn()
			initializer.initialized = true
		} else {
			panic("policy " + policy + " not found")
		}
	}
}
