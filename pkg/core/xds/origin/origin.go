package origin

// Origin is a marker (string) used to tag generated xDS resources with their provenance.
//
// What it is:
// - A lightweight label that identifies which generator/feature produced a given xDS resource.
//
// Where it lives:
//   - Stored on core xDS resources as core/xds.Resource.Origin and propagated through the entire
//     xDS assembly pipeline (see pkg/core/xds/resource.go).
//
// How it is set and used:
//   - Generators and features set the origin when constructing resources. Common constants are defined in
//     pkg/xds/generator/origin (e.g. "inbound", "outbound", "dns", "egress", "ingress", "transparent",
//     "prometheus", "secrets", "tracing", "probe", "proxy-template-raw", "proxy-template-modifications")
//     and in feature-specific packages like pkg/plugins/runtime/gateway/origin ("gateway").
//   - Hooks, policy plugins, and reconcilers filter or select resources by origin. For example, see
//     pkg/xds/hooks/resource_set.go and NonGatewayResources in pkg/core/xds/resource.go.
//
// Adding a new origin:
//   - Define a constant of this type close to the code that produces the resources, for example:
//     const MyFeature Origin = "my-feature"
//   - Prefer colocating origins in the corresponding generator/feature package to keep ownership clear.
type Origin string
