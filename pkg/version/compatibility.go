package version

type DataplaneCompatibility struct {
	Envoy string `json:"envoy"`
}

type Compatibility struct {
	KumaDP map[string]DataplaneCompatibility `json:"kumaDp"`
}

var CompatibilityMatrix = Compatibility{
	KumaDP: map[string]DataplaneCompatibility{
		"1.0.0": {
			Envoy: "1.16.0",
		},
		"1.0.1": {
			Envoy: "1.16.0",
		},
		"1.0.2": {
			Envoy: "1.16.1",
		},
		"1.0.3": {
			Envoy: "1.16.1",
		},
		"1.0.4": {
			Envoy: "1.16.1",
		},
		"1.0.5": {
			Envoy: "1.16.2",
		},
		"1.0.6": {
			Envoy: "1.16.2",
		},
		"1.0.7": {
			Envoy: "1.16.2",
		},
		"1.0.8": {
			Envoy: "1.16.2",
		},
		"~1.1.0": {
			Envoy: "~1.17.0",
		},
		"~1.2.0": {
			Envoy: "~1.18.0",
		},
		"~1.3.0": {
			Envoy: "~1.18.4",
		},
		"~1.4.0": {
			Envoy: "~1.18.4",
		},
		"~1.5.0": {
			Envoy: "~1.21.1",
		},
		"~1.6.0": {
			Envoy: "~1.21.1",
		},
		// This includes all dev versions branched from the first release
		// candidate (i.e. both master and release-1.4)
		// and all 1.4 releases and RCs. See Masterminds/semver#21
		"~1.6.1-anyprerelease": {
			Envoy: "~1.21.1",
		},
	},
}

var DevVersionPrefix = "dev"
