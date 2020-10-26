package v1alpha1

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

func (ds *DataSource) ConvertToEnvoy() *envoy_core.DataSource {
	if ds == nil {
		return nil
	}

	envoyDataSource := &envoy_core.DataSource{}
	switch ds.GetType().(type) {
	case *DataSource_Inline:
		envoyDataSource.Specifier = &envoy_core.DataSource_InlineBytes{
			InlineBytes: ds.GetInline().GetValue(),
		}
	case *DataSource_File:
		envoyDataSource.Specifier = &envoy_core.DataSource_Filename{
			Filename: ds.GetFile(),
		}
	case *DataSource_Secret:
		return nil
	default:
		return nil
	}

	return envoyDataSource
}
