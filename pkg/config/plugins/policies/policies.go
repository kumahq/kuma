package policies

func DefaultPoliciesConfig() *Config {
	return &Config{
		PluginPoliciesEnabled: []string{
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
		},
	}
}
