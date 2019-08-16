package generator

import (
	"bytes"
	"fmt"

	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/envoy"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/template"
	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
)

type TemplateProxyGenerator struct {
	ProxyTemplate *konvoy_mesh.ProxyTemplate
}

func (g *TemplateProxyGenerator) Generate(proxy *model.Proxy) ([]*Resource, error) {
	resources := make([]*Resource, 0, len(g.ProxyTemplate.Conf))
	for i, source := range g.ProxyTemplate.Conf {
		var generator ResourceGenerator
		switch s := source.Type.(type) {
		case *konvoy_mesh.ProxyTemplateSource_Profile:
			generator = &ProxyTemplateProfileSource{Profile: s.Profile}
		case *konvoy_mesh.ProxyTemplateSource_Raw:
			generator = &ProxyTemplateRawSource{Raw: s.Raw}
		default:
			return nil, fmt.Errorf("sources[%d]{name=%q}: unknown source type", i, source.Name)
		}
		if rs, err := generator.Generate(proxy); err != nil {
			return nil, fmt.Errorf("sources[%d]{name=%q}: %s", i, source.Name, err)
		} else {
			resources = append(resources, rs...)
		}
	}
	return resources, nil
}

type ProxyTemplateRawSource struct {
	Raw *konvoy_mesh.ProxyTemplateRawSource
}

func (s *ProxyTemplateRawSource) Generate(proxy *model.Proxy) ([]*Resource, error) {
	resources := make([]*Resource, 0, len(s.Raw.Resources))
	for i, r := range s.Raw.Resources {
		json, err := yaml.YAMLToJSON([]byte(r.Resource))
		if err != nil {
			json = []byte(r.Resource)
		}

		var any types.Any
		if err := (&jsonpb.Unmarshaler{}).Unmarshal(bytes.NewReader(json), &any); err != nil {
			return nil, fmt.Errorf("raw.resources[%d]{name=%q}.resource: %s", i, r.Name, err)
		}
		var dyn types.DynamicAny
		if err := types.UnmarshalAny(&any, &dyn); err != nil {
			return nil, fmt.Errorf("raw.resources[%d]{name=%q}.resource: %s", i, r.Name, err)
		}
		p, ok := dyn.Message.(ResourcePayload)
		if !ok {
			return nil, fmt.Errorf("raw.resources[%d]{name=%q}.resource: xDS resource doesn't implement all required interfaces", i, r.Name)
		}
		if v, ok := p.(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return nil, fmt.Errorf("raw.resources[%d]{name=%q}.resource: %s", i, r.Name, err)
			}
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

func NewDefaultProxyProfile() ResourceGenerator {
	return CompositeResourceGenerator{TransparentProxyGenerator{}, InboundProxyGenerator{}}
}

func init() {
	predefinedProfiles[template.ProfileDefaultProxy] = NewDefaultProxyProfile()
}

type ProxyTemplateProfileSource struct {
	Profile *konvoy_mesh.ProxyTemplateProfileSource
}

func (s *ProxyTemplateProfileSource) Generate(proxy *model.Proxy) ([]*Resource, error) {
	g, ok := predefinedProfiles[s.Profile.Name]
	if !ok {
		return nil, fmt.Errorf("profile{name=%q}: unknown profile", s.Profile.Name)
	}
	return g.Generate(proxy)
}

type InboundProxyGenerator struct {
}

func (_ InboundProxyGenerator) Generate(proxy *model.Proxy) ([]*Resource, error) {
	endpoints, err := proxy.Dataplane.Spec.Networking.GetInboundInterfaces()
	if err != nil {
		return nil, err
	}
	if len(endpoints) == 0 {
		return nil, nil
	}
	virtual := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetRedirectPort() != 0
	resources := make([]*Resource, 0, len(endpoints))
	names := make(map[string]bool)
	for _, endpoint := range endpoints {
		localClusterName := fmt.Sprintf("localhost:%d", endpoint.WorkloadPort)
		if used := names[localClusterName]; !used {
			resources = append(resources, &Resource{
				Name:     localClusterName,
				Version:  proxy.Dataplane.Meta.GetVersion(),
				Resource: envoy.CreateLocalCluster(localClusterName, "127.0.0.1", endpoint.WorkloadPort),
			})
			names[localClusterName] = true
		}

		inboundListenerName := fmt.Sprintf("inbound:%s:%d", endpoint.DataplaneIP, endpoint.DataplanePort)
		if used := names[inboundListenerName]; !used {
			resources = append(resources, &Resource{
				Name:     inboundListenerName,
				Version:  proxy.Dataplane.Meta.GetVersion(),
				Resource: envoy.CreateInboundListener(inboundListenerName, endpoint.DataplaneIP, endpoint.DataplanePort, localClusterName, virtual),
			})
			names[inboundListenerName] = true
		}
	}
	return resources, nil
}

type TransparentProxyGenerator struct {
}

func (_ TransparentProxyGenerator) Generate(proxy *model.Proxy) ([]*Resource, error) {
	redirectPort := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetRedirectPort()
	if redirectPort == 0 {
		return nil, nil
	}
	return []*Resource{
		&Resource{
			Name:     "catch_all",
			Version:  proxy.Dataplane.Meta.GetVersion(),
			Resource: envoy.CreateCatchAllListener("catch_all", "0.0.0.0", redirectPort, "pass_through"),
		},
		&Resource{
			Name:     "pass_through",
			Version:  proxy.Dataplane.Meta.GetVersion(),
			Resource: envoy.CreatePassThroughCluster("pass_through"),
		},
	}, nil
}
