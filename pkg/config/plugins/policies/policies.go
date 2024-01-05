package policies

func DefaultPoliciesConfig() *PoliciesConfig {
	return &PoliciesConfig{
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
