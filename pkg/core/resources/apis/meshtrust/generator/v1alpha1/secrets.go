package v1alpha1

import (
	"bytes"
	"sort"

	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/types/known/anypb"

	core_plugins "github.com/kumahq/kuma/v2/pkg/core/plugins"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshtrust/metadata"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	bldrs_auth "github.com/kumahq/kuma/v2/pkg/envoy/builders/auth"
	bldrs_core "github.com/kumahq/kuma/v2/pkg/envoy/builders/core"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	"github.com/kumahq/kuma/v2/pkg/xds/generator/system_names"
)

var _ core_plugins.CoreResourcePlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.CoreResourcePlugin {
	return &plugin{}
}

func (p *plugin) Generate(rs *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) error {
	// When using SPIRE, we skip ValidationContext generation for the dataplane,
	// since SPIRE is responsible for delivering the validation context.
	// We should investigate whether it's possible to support both mechanisms simultaneously.
	// TODO: https://github.com/kumahq/kuma/issues/14685
	externallyManaged := pointer.Deref(proxy.WorkloadIdentity).ManagementMode == core_xds.ExternalManagementMode
	hasTrustDomains := len(xdsCtx.Mesh.CAsByTrustDomain) > 0

	if externallyManaged || !hasTrustDomains {
		return nil
	}

	config, err := validationCtx(xdsCtx)
	if err != nil {
		return err
	}
	rs.Add(&core_xds.Resource{
		Name:     config.Name,
		Origin:   metadata.OriginMeshTrust,
		Resource: config,
	})

	return nil
}

func validationCtx(xdsCtx xds_context.Context) (*envoy_auth.Secret, error) {
	validatorsPerTrustDomain := []*envoy_auth.SPIFFECertValidatorConfig_TrustDomain{}
	for domain, trusts := range xdsCtx.Mesh.CAsByTrustDomain {
		// concatenate multiple CAs
		allCAs := [][]byte{}
		for _, ca := range trusts {
			allCAs = append(allCAs, []byte(ca))
		}
		concatenatedCA := bytes.Join(allCAs, []byte("\n"))
		validator, err := bldrs_auth.NewSPIFFECertValidator().
			Configure(
				bldrs_auth.TrustDomainBundle(domain,
					bldrs_core.NewDataSource().Configure(bldrs_core.InlineBytes(concatenatedCA)))).Build()
		if err != nil {
			return nil, err
		}
		validatorsPerTrustDomain = append(validatorsPerTrustDomain, validator)
	}
	// Order by trustdomain name to return in consistent order
	sort.Slice(validatorsPerTrustDomain, func(i, j int) bool {
		return validatorsPerTrustDomain[i].Name < validatorsPerTrustDomain[j].Name
	})

	typedConfig, err := anypb.New(&envoy_auth.SPIFFECertValidatorConfig{
		TrustDomains: validatorsPerTrustDomain,
	})
	if err != nil {
		return nil, err
	}
	ca, err := bldrs_auth.NewSecret().
		Configure(bldrs_auth.Name(system_names.SystemResourceNameCABundle)).
		Configure(bldrs_auth.ValidationContext(
			bldrs_auth.NewValidationContext().
				Configure(
					bldrs_auth.CertificateValidationContext(
						bldrs_auth.NewCertificateValidationContext().
							Configure(bldrs_auth.SpiffeCustomValidator(typedConfig)))))).Build()
	if err != nil {
		return nil, err
	}
	return ca, nil
}
