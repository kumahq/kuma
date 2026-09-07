package zone

import "slices"

// FullScope lists the scopes a Zone Token can be issued for. The zone proxy
// scopes are gone: a zone proxy is an ordinary Dataplane and authenticates with
// a dataplane token. Distributions that issue Zone Tokens for their own
// components append their scopes here during initialization.
var FullScope []string

func InScope(scope []string, s string) bool {
	return slices.Contains(scope, s)
}
