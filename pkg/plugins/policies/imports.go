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
	"os"
	"strings"
)

var nameToModule = map[string]func(bool){
	"meshaccesslog":             meshaccesslog.InitPlugin,
	"meshcircuitbreaker":        meshcircuitbreaker.InitPlugin,
	"meshfaultinjection":        meshfaultinjection.InitPlugin,
	"meshhealthcheck":           meshhealthcheck.InitPlugin,
	"meshhttproute":             meshhttproute.InitPlugin,
	"meshloadbalancingstrategy": meshloadbalancingstrategy.InitPlugin,
	"meshmetric":                meshmetric.InitPlugin,
	"meshproxypatch":            meshproxypatch.InitPlugin,
	"meshratelimit":             meshratelimit.InitPlugin,
	"meshretry":                 meshretry.InitPlugin,
	"meshtcproute":              meshtcproute.InitPlugin,
	"meshtimeout":               meshtimeout.InitPlugin,
	"meshtrace":                 meshtrace.InitPlugin,
	"meshtrafficpermission":     meshtrafficpermission.InitPlugin,
}

func initAllPolicies() {
	for _, initializer := range nameToModule {
		initializer(true)
	}
}

func init() {
	// we read this manually otherwise we would have to wire in enabling plugins in every test
	rawEnabledPluginPolicies := os.Getenv("KUMA_PLUGIN_POLICIES_ENABLED")
	if rawEnabledPluginPolicies == "" {
		initAllPolicies()
	} else {
		enabledPluginPolicies := strings.Split(rawEnabledPluginPolicies, ",")
		for _, policy := range enabledPluginPolicies {
			initializer, ok := nameToModule[policy]
			if ok {
				initializer(true)
			}
		}
	}
}
