// Package metadata provides lightweight, import-cycle-safe constants shared by
// multiple components (generators, plugins, controllers, hooks, etc.).
// Keeping per-feature constants in a tiny leaf package helps avoid pulling
// heavy transitive dependencies across the build graph and keeps ownership clear
package metadata

import . "github.com/kumahq/kuma/pkg/core/xds/origin"

// OriginMeshTrust marks xDS resources that were generated or
// modified by the MeshTrust policy processing pipeline
const OriginMeshTrust Origin = "mesh-identity-spire"
