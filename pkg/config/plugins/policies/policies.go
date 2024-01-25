package policies

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
	return &Config{
		PluginPoliciesEnabled: DefaultPluginPoliciesEnabled,
	}
}
