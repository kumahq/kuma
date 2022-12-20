package policies

import (
	_ "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog"
	_ "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker"
	_ "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck"
	_ "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit"
	_ "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout"
	_ "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace"
	_ "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission"
)
