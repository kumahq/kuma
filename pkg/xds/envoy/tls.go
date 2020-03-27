package envoy

import (
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_grpc_credential "github.com/envoyproxy/go-control-plane/envoy/config/grpc_credential/v2alpha"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	core_xds "github.com/Kong/kuma/pkg/core/xds"
	sds_provider "github.com/Kong/kuma/pkg/sds/provider"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
)

func CreateDownstreamTlsContext(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) (*envoy_auth.DownstreamTlsContext, error) {
	if !ctx.Mesh.Resource.Spec.GetMtls().GetEnabled() {
		return nil, nil
	}
	commonTlsContext, err := CreateCommonTlsContext(ctx, metadata)
	if err != nil {
		return nil, err
	}
	return &envoy_auth.DownstreamTlsContext{
		CommonTlsContext:         commonTlsContext,
		RequireClientCertificate: &wrappers.BoolValue{Value: true},
	}, nil
}

func CreateUpstreamTlsContext(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) (*envoy_auth.UpstreamTlsContext, error) {
	if !ctx.Mesh.Resource.Spec.GetMtls().GetEnabled() {
		return nil, nil
	}
	commonTlsContext, err := CreateCommonTlsContext(ctx, metadata)
	if err != nil {
		return nil, err
	}
	return &envoy_auth.UpstreamTlsContext{
		CommonTlsContext: commonTlsContext,
	}, nil
}

func CreateCommonTlsContext(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) (*envoy_auth.CommonTlsContext, error) {
	meshCaSecret, err := sdsSecretConfig(ctx, sds_provider.MeshCaResource, metadata)
	if err != nil {
		return nil, err
	}
	identitySecret, err := sdsSecretConfig(ctx, sds_provider.IdentityCertResource, metadata)
	if err != nil {
		return nil, err
	}
	return &envoy_auth.CommonTlsContext{
		ValidationContextType: &envoy_auth.CommonTlsContext_ValidationContextSdsSecretConfig{
			ValidationContextSdsSecretConfig: meshCaSecret,
		},
		TlsCertificateSdsSecretConfigs: []*envoy_auth.SdsSecretConfig{
			identitySecret,
		},
	}, nil
}

func sdsSecretConfig(context xds_context.Context, name string, metadata *core_xds.DataplaneMetadata) (*envoy_auth.SdsSecretConfig, error) {
	googleGrpc := &envoy_core.GrpcService_GoogleGrpc{
		StatPrefix: util_xds.SanitizeMetric("sds_" + name),
	}
	if context.Mesh.Resource.HasVaultCA() {
		// In this case SDS is embedded into Kuma DP and exposed via unix socket therefore we don't need any TLS/auth.
		if metadata.DataplaneSdsAddress == "" {
			return nil, errors.New("dataplane.sds.path in Envoy metadata cannot be empty when Vault CA is set")
		}
		googleGrpc.TargetUri = metadata.DataplaneSdsAddress
	} else {
		googleGrpc.TargetUri = context.ControlPlane.SdsLocation
		googleGrpc.ChannelCredentials = &envoy_core.GrpcService_GoogleGrpc_ChannelCredentials{
			CredentialSpecifier: &envoy_core.GrpcService_GoogleGrpc_ChannelCredentials_SslCredentials{
				SslCredentials: &envoy_core.GrpcService_GoogleGrpc_SslCredentials{
					RootCerts: &envoy_core.DataSource{
						Specifier: &envoy_core.DataSource_InlineBytes{
							InlineBytes: context.ControlPlane.SdsTlsCert,
						},
					},
				},
			},
		}
		if metadata.GetDataplaneTokenPath() != "" { // Token path may not be set when token verification is off
			if err := grpcWithCallCredentials(googleGrpc, metadata.DataplaneTokenPath); err != nil {
				return nil, err
			}
		}
	}
	return &envoy_auth.SdsSecretConfig{
		Name: name,
		SdsConfig: &envoy_core.ConfigSource{
			ConfigSourceSpecifier: &envoy_core.ConfigSource_ApiConfigSource{
				ApiConfigSource: &envoy_core.ApiConfigSource{
					ApiType: envoy_core.ApiConfigSource_GRPC,
					GrpcServices: []*envoy_core.GrpcService{
						{
							TargetSpecifier: &envoy_core.GrpcService_GoogleGrpc_{
								GoogleGrpc: googleGrpc,
							},
						},
					},
				},
			},
		},
	}, nil
}

func grpcWithCallCredentials(grpc *envoy_core.GrpcService_GoogleGrpc, credentialPath string) error {
	if credentialPath == "" {
		return nil
	}

	config := &envoy_grpc_credential.FileBasedMetadataConfig{
		SecretData: &envoy_core.DataSource{
			Specifier: &envoy_core.DataSource_Filename{
				Filename: credentialPath,
			},
		},
	}
	typedConfig, err := ptypes.MarshalAny(config)
	if err != nil {
		return err
	}

	grpc.CallCredentials = append(grpc.CallCredentials, &envoy_core.GrpcService_GoogleGrpc_CallCredentials{
		CredentialSpecifier: &envoy_core.GrpcService_GoogleGrpc_CallCredentials_FromPlugin{
			FromPlugin: &envoy_core.GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin{
				Name: "envoy.grpc_credentials.file_based_metadata",
				ConfigType: &envoy_core.GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin_TypedConfig{
					TypedConfig: typedConfig,
				},
			},
		},
	})
	grpc.CredentialsFactoryName = "envoy.grpc_credentials.file_based_metadata"
	return nil
}
