package auth

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/types/known/anypb"

	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
)

func NewSecret() *Builder[envoy_auth.Secret] {
	return &Builder[envoy_auth.Secret]{}
}

func Name(name string) Configurer[envoy_auth.Secret] {
	return func(s *envoy_auth.Secret) error {
		s.Name = name
		return nil
	}
}

func ValidationContext(tls *Builder[envoy_auth.Secret_ValidationContext]) Configurer[envoy_auth.Secret] {
	return func(s *envoy_auth.Secret) error {
		conf, err := tls.Build()
		if err != nil {
			return err
		}
		s.Type = conf
		return nil
	}
}

func TlsCertificate(tls *Builder[envoy_auth.Secret_TlsCertificate]) Configurer[envoy_auth.Secret] {
	return func(s *envoy_auth.Secret) error {
		conf, err := tls.Build()
		if err != nil {
			return err
		}
		s.Type = conf
		return nil
	}
}

func NewValidationContext() *Builder[envoy_auth.Secret_ValidationContext] {
	return &Builder[envoy_auth.Secret_ValidationContext]{}
}

func CertificateValidationContext(validationCtx *Builder[envoy_auth.CertificateValidationContext]) Configurer[envoy_auth.Secret_ValidationContext] {
	return func(s *envoy_auth.Secret_ValidationContext) error {
		conf, err := validationCtx.Build()
		if err != nil {
			return err
		}
		s.ValidationContext = conf
		return nil
	}
}

func NewCertificateValidationContext() *Builder[envoy_auth.CertificateValidationContext] {
	return &Builder[envoy_auth.CertificateValidationContext]{}
}

func SpiffeCustomValidator(typedConf *anypb.Any) Configurer[envoy_auth.CertificateValidationContext] {
	return func(s *envoy_auth.CertificateValidationContext) error {
		s.CustomValidatorConfig = &envoy_core.TypedExtensionConfig{
			Name:        "envoy.tls.cert_validator.spiffe",
			TypedConfig: typedConf,
		}
		return nil
	}
}

func NewTlsCertificate() *Builder[envoy_auth.Secret_TlsCertificate] {
	return &Builder[envoy_auth.Secret_TlsCertificate]{}
}

func CertificateChain(datasource *Builder[envoy_core.DataSource]) Configurer[envoy_auth.Secret_TlsCertificate] {
	return func(s *envoy_auth.Secret_TlsCertificate) error {
		ds, err := datasource.Build()
		if err != nil {
			return nil
		}
		if s.TlsCertificate == nil {
			s.TlsCertificate = &envoy_auth.TlsCertificate{}
		}
		s.TlsCertificate.CertificateChain = ds
		return nil
	}
}

func PrivateKey(datasource *Builder[envoy_core.DataSource]) Configurer[envoy_auth.Secret_TlsCertificate] {
	return func(s *envoy_auth.Secret_TlsCertificate) error {
		ds, err := datasource.Build()
		if err != nil {
			return nil
		}
		if s.TlsCertificate == nil {
			s.TlsCertificate = &envoy_auth.TlsCertificate{}
		}
		s.TlsCertificate.PrivateKey = ds
		return nil
	}
}

func NewSPIFFECertValidator() *Builder[envoy_auth.SPIFFECertValidatorConfig_TrustDomain] {
	return &Builder[envoy_auth.SPIFFECertValidatorConfig_TrustDomain]{}
}

func TrustDomainBundle(trustDomain string, datasource *Builder[envoy_core.DataSource]) Configurer[envoy_auth.SPIFFECertValidatorConfig_TrustDomain] {
	return func(s *envoy_auth.SPIFFECertValidatorConfig_TrustDomain) error {
		ds, err := datasource.Build()
		if err != nil {
			return nil
		}
		s.Name = trustDomain
		s.TrustBundle = ds
		return nil
	}
}
