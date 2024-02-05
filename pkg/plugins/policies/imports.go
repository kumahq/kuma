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

type PluginInitializer struct {
	InitFn      func()
	Initialized bool
}

var NameToModule = map[string]*PluginInitializer{
	"meshaccesslogs":              {InitFn: meshaccesslog.InitPlugin, Initialized: false},
	"meshcircuitbreakers":         {InitFn: meshcircuitbreaker.InitPlugin, Initialized: false},
	"meshfaultinjections":         {InitFn: meshfaultinjection.InitPlugin, Initialized: false},
	"meshhealthchecks":            {InitFn: meshhealthcheck.InitPlugin, Initialized: false},
	"meshhttproutes":              {InitFn: meshhttproute.InitPlugin, Initialized: false},
	"meshloadbalancingstrategies": {InitFn: meshloadbalancingstrategy.InitPlugin, Initialized: false},
	"meshmetrics":                 {InitFn: meshmetric.InitPlugin, Initialized: false},
	"meshproxypatches":            {InitFn: meshproxypatch.InitPlugin, Initialized: false},
	"meshratelimits":              {InitFn: meshratelimit.InitPlugin, Initialized: false},
	"meshretries":                 {InitFn: meshretry.InitPlugin, Initialized: false},
	"meshtcproutes":               {InitFn: meshtcproute.InitPlugin, Initialized: false},
	"meshtimeouts":                {InitFn: meshtimeout.InitPlugin, Initialized: false},
	"meshtraces":                  {InitFn: meshtrace.InitPlugin, Initialized: false},
	"meshtrafficpermissions":      {InitFn: meshtrafficpermission.InitPlugin, Initialized: false},
}

func InitAllPolicies() {
	for _, initializer := range NameToModule {
		if !initializer.Initialized {
			initializer.InitFn()
			initializer.Initialized = true
		}
	}
}

func InitPolicies(enabledPluginPolicies []string) {
	for _, policy := range enabledPluginPolicies {
		initializer, ok := NameToModule[policy]
		if ok && !initializer.Initialized {
			initializer.InitFn()
			initializer.Initialized = true
		} else {
			panic("policy " + policy + " not found")
		}
	}
}
