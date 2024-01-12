package policies

func DefaultPoliciesConfig() *Config {
	return &Config{
		PluginPoliciesEnabled: []string{
			"meshaccesslog",
			"meshcircuitbreaker",
			"meshfaultinjection",
			"meshhealthcheck",
			"meshhttproute",
			"meshloadbalancingstrategy",
			"meshmetric",
			"meshproxypatch",
			"meshratelimit",
			"meshretry",
			"meshtcproute",
			"meshtimeout",
			"meshtrace",
			"meshtrafficpermission",
		},
	}
}
