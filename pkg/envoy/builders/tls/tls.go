package tls

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	common_tls "github.com/kumahq/kuma/api/common/v1alpha1/tls"
	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

func NewUpstreamTLSContext() *Builder[envoy_tls.UpstreamTlsContext] {
	return &Builder[envoy_tls.UpstreamTlsContext]{}
}

func NewDownstreamTLSContext() *Builder[envoy_tls.DownstreamTlsContext] {
	return &Builder[envoy_tls.DownstreamTlsContext]{}
}

func SNI(sni string) Configurer[envoy_tls.UpstreamTlsContext] {
	return func(c *envoy_tls.UpstreamTlsContext) error {
		c.Sni = sni
		return nil
	}
}

func UpstreamCommonTlsContext(commonCtx *Builder[envoy_tls.CommonTlsContext]) Configurer[envoy_tls.UpstreamTlsContext] {
	return func(c *envoy_tls.UpstreamTlsContext) error {
		config, err := commonCtx.Build()
		if err != nil {
			return err
		}
		c.CommonTlsContext = config
		return nil
	}
}

func DownstreamCommonTlsContext(commonCtx *Builder[envoy_tls.CommonTlsContext]) Configurer[envoy_tls.DownstreamTlsContext] {
	return func(c *envoy_tls.DownstreamTlsContext) error {
		config, err := commonCtx.Build()
		if err != nil {
			return err
		}
		c.CommonTlsContext = config
		return nil
	}
}

func RequireClientCertificate(require bool) Configurer[envoy_tls.DownstreamTlsContext] {
	return func(c *envoy_tls.DownstreamTlsContext) error {
		c.RequireClientCertificate = util_proto.Bool(require)
		return nil
	}
}

func NewCommonTlsContext() *Builder[envoy_tls.CommonTlsContext] {
	return &Builder[envoy_tls.CommonTlsContext]{}
}

func TlsCertificateSdsSecretConfigs(builders []*Builder[envoy_tls.SdsSecretConfig]) Configurer[envoy_tls.CommonTlsContext] {
	return func(c *envoy_tls.CommonTlsContext) error {
		for _, builder := range builders {
			config, err := builder.Build()
			if err != nil {
				return err
			}
			c.TlsCertificateSdsSecretConfigs = append(c.TlsCertificateSdsSecretConfigs, config)
		}
		return nil
	}
}

func KumaAlpnProtocol() Configurer[envoy_tls.CommonTlsContext] {
	return func(c *envoy_tls.CommonTlsContext) error {
		c.AlpnProtocols = xds_tls.KumaALPNProtocols
		return nil
	}
}

func CipherSuites(cipherSuites []common_tls.TlsCipher) Configurer[envoy_tls.CommonTlsContext] {
	return func(c *envoy_tls.CommonTlsContext) error {
		if c.TlsParams == nil {
			c.TlsParams = &envoy_tls.TlsParameters{}
		}
		ciphers := []string{}
		for _, cipher := range cipherSuites {
			ciphers = append(ciphers, string(cipher))
		}
		c.TlsParams.CipherSuites = ciphers
		return nil
	}
}

func TlsMaxVersion(version *common_tls.TlsVersion) Configurer[envoy_tls.CommonTlsContext] {
	return func(c *envoy_tls.CommonTlsContext) error {
		if c.TlsParams == nil {
			c.TlsParams = &envoy_tls.TlsParameters{}
		}
		c.TlsParams.TlsMaximumProtocolVersion = common_tls.ToTlsVersion(version)
		return nil
	}
}

func TlsMinVersion(version *common_tls.TlsVersion) Configurer[envoy_tls.CommonTlsContext] {
	return func(c *envoy_tls.CommonTlsContext) error {
		if c.TlsParams == nil {
			c.TlsParams = &envoy_tls.TlsParameters{}
		}
		c.TlsParams.TlsMinimumProtocolVersion = common_tls.ToTlsVersion(version)
		return nil
	}
}

func CombinedCertificateValidationContext(builder *Builder[envoy_tls.CommonTlsContext_CombinedCertificateValidationContext]) Configurer[envoy_tls.CommonTlsContext] {
	return func(c *envoy_tls.CommonTlsContext) error {
		config, err := builder.Build()
		if err != nil {
			return err
		}
		c.ValidationContextType = &envoy_tls.CommonTlsContext_CombinedValidationContext{
			CombinedValidationContext: config,
		}
		return nil
	}
}

func NewCombinedCertificateValidationContext() *Builder[envoy_tls.CommonTlsContext_CombinedCertificateValidationContext] {
	return &Builder[envoy_tls.CommonTlsContext_CombinedCertificateValidationContext]{}
}

func NewDefaultValidationContext() *Builder[envoy_tls.CertificateValidationContext] {
	return &Builder[envoy_tls.CertificateValidationContext]{}
}

func SANs(builders []*Builder[envoy_tls.SubjectAltNameMatcher]) Configurer[envoy_tls.CertificateValidationContext] {
	return func(c *envoy_tls.CertificateValidationContext) error {
		for _, builder := range builders {
			config, err := builder.Build()
			if err != nil {
				return nil
			}
			c.MatchTypedSubjectAltNames = append(c.MatchTypedSubjectAltNames, config)
		}
		return nil
	}
}

func NewSubjectAltNameMatcher() *Builder[envoy_tls.SubjectAltNameMatcher] {
	return &Builder[envoy_tls.SubjectAltNameMatcher]{}
}

func URI(matcher *Builder[envoy_type_matcher.StringMatcher]) Configurer[envoy_tls.SubjectAltNameMatcher] {
	return func(c *envoy_tls.SubjectAltNameMatcher) error {
		config, err := matcher.Build()
		if err != nil {
			return err
		}
		c.SanType = envoy_tls.SubjectAltNameMatcher_URI
		c.Matcher = config
		return nil
	}
}

func ValidationContextSdsSecretConfig(builder *Builder[envoy_tls.SdsSecretConfig]) Configurer[envoy_tls.CommonTlsContext_CombinedCertificateValidationContext] {
	return func(c *envoy_tls.CommonTlsContext_CombinedCertificateValidationContext) error {
		config, err := builder.Build()
		if err != nil {
			return err
		}
		c.ValidationContextSdsSecretConfig = config
		return nil
	}
}

func DefaultValidationContext(builder *Builder[envoy_tls.CertificateValidationContext]) Configurer[envoy_tls.CommonTlsContext_CombinedCertificateValidationContext] {
	return func(c *envoy_tls.CommonTlsContext_CombinedCertificateValidationContext) error {
		config, err := builder.Build()
		if err != nil {
			return err
		}
		c.DefaultValidationContext = config
		return nil
	}
}

func NewTlsCertificateSdsSecretConfigs() *Builder[envoy_tls.SdsSecretConfig] {
	return &Builder[envoy_tls.SdsSecretConfig]{}
}

func SdsSecretConfigSource(secretName string, configSource *Builder[envoy_core.ConfigSource]) Configurer[envoy_tls.SdsSecretConfig] {
	return func(c *envoy_tls.SdsSecretConfig) error {
		cs, err := configSource.Build()
		if err != nil {
			return err
		}
		c.Name = secretName
		c.SdsConfig = cs
		return nil
	}
}
