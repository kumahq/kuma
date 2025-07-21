package system_names

import "github.com/kumahq/kuma/pkg/core/system_names"

var SystemResourceNameDynamicConfigListener = system_names.AsSystemName("dynamicconfig")

func SystemResourceNameDynamicConfigRoute(routeName string) string {
	return system_names.AsSystemName(system_names.Join("dynamicconfig", routeName))
}

func SystemResourceNameDynamicConfigRouteNotModified(routeName string) string {
	return system_names.AsSystemName(system_names.Join("dynamicconfig", routeName, "not_modified"))
}
