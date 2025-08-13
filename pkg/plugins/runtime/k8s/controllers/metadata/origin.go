// This file lives in a dedicated "metadata" subpackage, because the Origin
// type/values are imported by many components (generators, plugins,
// controllers, hooks, etc.). Keeping them in a tiny leaf package avoids
// import cycles and prevents pulling heavy transitive dependencies across
// the build graph. Per-feature constants live in their own metadata
// subpackages to keep ownership clear while keeping dependencies minimal
package metadata

import . "github.com/kumahq/kuma/pkg/core/xds/origin"

// OriginKube is the origin for Kubernetes-specific resources
const OriginKube Origin = "kubernetes"
