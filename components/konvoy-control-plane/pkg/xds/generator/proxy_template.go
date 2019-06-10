package generator

import (
	"bytes"
	"fmt"

	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/model/api/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/envoy"
	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
)

type ProxyTemplateGenerator struct {
	ProxyTemplate *konvoy_mesh.ProxyTemplate
}

func (g *ProxyTemplateGenerator) Generate(proxy *Proxy) ([]*Resource, error) {
	resources := make([]*Resource, 0, len(g.ProxyTemplate.Spec.Sources))
	for _, source := range g.ProxyTemplate.Spec.Sources {
		var generator ResourceGenerator
		switch {
		case source.Profile != nil:
			generator = &ProxyTemplateProfileSource{Profile: source.Profile}
		case source.Raw != nil:
			generator = &ProxyTemplateRawSource{Raw: source.Raw}
		default:
			return nil, fmt.Errorf("unknown source type: %#v", source)
		}
		if rs, err := generator.Generate(proxy); err != nil {
			return []*Resource{}, err
		} else {
			resources = append(resources, rs...)
		}

	}
	return resources, nil
}

type ProxyTemplateRawSource struct {
	Raw *konvoy_mesh.ProxyTemplateRawSource
}

func (s *ProxyTemplateRawSource) Generate(proxy *Proxy) ([]*Resource, error) {
	resources := make([]*Resource, 0, len(s.Raw.Resources))
	for i := range s.Raw.Resources {
		r := &s.Raw.Resources[i]

		json, err := yaml.YAMLToJSON([]byte(r.Resource))
		if err != nil {
			json = []byte(r.Resource)
		}

		var any types.Any
		if err := (&jsonpb.Unmarshaler{}).Unmarshal(bytes.NewReader(json), &any); err != nil {
			return nil, err
		}
		var dyn types.DynamicAny
		if err := types.UnmarshalAny(&any, &dyn); err != nil {
			return nil, err
		}
		p, ok := dyn.Message.(ResourcePayload)
		if !ok {
			return nil, fmt.Errorf("resource %q doesn't implement all required interfaces", r.Name)
		}

		resources = append(resources, &Resource{
			Name:     r.Name,
			Version:  r.Version,
			Resource: p,
		})
	}
	return resources, nil
}

var predefinedProfiles = make(map[string]ResourceGenerator)

func init() {
	predefinedProfiles["transparent-inbound-proxy"] = &TransparentInboundProxyProfile{}
	predefinedProfiles["transparent-outbound-proxy"] = &TransparentOutboundProxyProfile{}
}

type ProxyTemplateProfileSource struct {
	Profile *konvoy_mesh.ProxyTemplateProfileSource
}

func (s *ProxyTemplateProfileSource) Generate(proxy *Proxy) ([]*Resource, error) {
	g, ok := predefinedProfiles[s.Profile.Name]
	if !ok {
		return nil, fmt.Errorf("unknown profile: %s", s.Profile.Name)
	}
	return g.Generate(proxy)
}

type TransparentInboundProxyProfile struct {
}

func (p *TransparentInboundProxyProfile) Generate(proxy *Proxy) ([]*Resource, error) {
	if len(proxy.Workload.Addresses) == 0 || len(proxy.Workload.Ports) == 0 {
		return nil, nil
	}
	resources := make([]*Resource, 0, len(proxy.Workload.Addresses)*len(proxy.Workload.Ports))
	for _, port := range proxy.Workload.Ports {
		localClusterName := fmt.Sprintf("localhost:%d", port)
		resources = append(resources, &Resource{
			Name:     localClusterName,
			Version:  proxy.Workload.Version,
			Resource: envoy.CreateLocalCluster(localClusterName, "127.0.0.1", port),
		})

		for _, address := range proxy.Workload.Addresses {
			inboundListenerName := fmt.Sprintf("inbound:%s:%d", address, port)

			resources = append(resources, &Resource{
				Name:     inboundListenerName,
				Version:  proxy.Workload.Version,
				Resource: envoy.CreateInboundListener(inboundListenerName, address, port, localClusterName),
			})
		}
	}
	return resources, nil
}

type TransparentOutboundProxyProfile struct {
}

func (p *TransparentOutboundProxyProfile) Generate(proxy *Proxy) ([]*Resource, error) {
	return []*Resource{
		&Resource{
			Name:     "catch_all",
			Version:  proxy.Workload.Version,
			Resource: envoy.CreateCatchAllListener("catch_all", "0.0.0.0", 15001, "pass_through"),
		},
		&Resource{
			Name:     "pass_through",
			Version:  proxy.Workload.Version,
			Resource: envoy.CreatePassThroughCluster("pass_through"),
		},
	}, nil
}
