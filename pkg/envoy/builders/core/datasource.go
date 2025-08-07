package core

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
)

func NewDataSource() *Builder[envoy_core.DataSource] {
	return &Builder[envoy_core.DataSource]{}
}

func InlineBytes(value []byte) Configurer[envoy_core.DataSource] {
	return func(ds *envoy_core.DataSource) error {
		ds.Specifier = &envoy_core.DataSource_InlineBytes{
			InlineBytes: value,
		}
		return nil
	}
}
