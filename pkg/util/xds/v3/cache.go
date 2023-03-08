// Copyright 2018 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package v3

import (
	"context"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

type Snapshot interface {
	envoy_cache.ResourceSnapshot

	Consistent() error
	// WithVersion creates a new snapshot with a different version for a given resource type.
	WithVersion(typ string, version string) Snapshot
}

type SnapshotGenerator interface {
	GenerateSnapshot(context.Context, *envoy_config_core_v3.Node) (Snapshot, error)
}