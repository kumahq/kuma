package policies

import (
	meshaccesslog "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog"
	meshcircuitbreaker "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker"
	meshfaultinjection "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection"
	meshhealthcheck "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck"
	meshhttproute "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute"
	meshloadbalancingstrategy "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy"
	meshproxypatch "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch"
	meshratelimit "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit"
	meshretry "github.com/kumahq/kuma/pkg/plugins/policies/meshretry"
	meshtcproute "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute"
	meshtimeout "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout"
	meshtrace "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace"
	meshtrafficpermission "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission"
)

func init() {
	meshhttproute.Register()
	meshtcproute.Register()
	meshloadbalancingstrategy.Register()
	meshaccesslog.Register()
	meshtrace.Register()
	meshfaultinjection.Register()
	meshratelimit.Register()
	meshtimeout.Register()
	meshtrafficpermission.Register()
	meshcircuitbreaker.Register()
	meshhealthcheck.Register()
	meshretry.Register()
	meshproxypatch.Register()
}
