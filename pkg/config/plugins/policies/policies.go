package policies

import "golang.org/x/exp/slices"

var DefaultPluginPoliciesEnabled = []string{
	"meshaccesslogs",
	"meshcircuitbreakers",
	"meshfaultinjections",
	"meshhealthchecks",
	"meshhttproutes",
	"meshloadbalancingstrategies",
	"meshmetrics",
	"meshproxypatches",
	"meshratelimits",
	"meshretries",
	"meshtcproutes",
	"meshtimeouts",
	"meshtraces",
	"meshtrafficpermissions",
}

func DefaultPoliciesConfig() *Config {
	slices.Sort(DefaultPluginPoliciesEnabled)
	return &Config{
		PluginPoliciesEnabled: DefaultPluginPoliciesEnabled,
	}
}
