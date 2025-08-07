package core

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
)

func NewConfigSource() *Builder[envoy_core.ConfigSource] {
	return &Builder[envoy_core.ConfigSource]{}
}

func Sds() Configurer[envoy_core.ConfigSource] {
	return func(cs *envoy_core.ConfigSource) error {
		cs.ResourceApiVersion = envoy_core.ApiVersion_V3
		cs.ConfigSourceSpecifier = &envoy_core.ConfigSource_Ads{}
		return nil
	}
}
