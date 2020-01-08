package xds

import (
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"
)

// ResourcePayload is a convenience type alias.
type ResourcePayload = envoy_cache.Resource

// Resource represents a generic xDS resource with name and version.
type Resource struct {
	Name     string
	Version  string
	Resource ResourcePayload
}
