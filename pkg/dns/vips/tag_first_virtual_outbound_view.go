package vips

import (
	"maps"
	"reflect"
	"strings"

	"github.com/asaskevich/govalidator"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

// TagFirstVirtualOutboundView was designed to compress VirtualOutbound configuration.
// It uses shortened field names and changes aggregation of outbounds.
// Now we don't duplicate tags, especially kuma.io/service tag.
// This model produces configuration which is around 50% characters smaller than default one
type TagFirstVirtualOutboundView struct {
	PerService map[string]*TagFirstOutbound `json:"v,omitempty"`
}

type TagFirstOutbound struct {
	Outbounds []*Outbound `json:"oe,omitempty"`
}

type Outbound struct {
	Tags         map[string]string `json:"t,omitempty"`
	AddressPorts []AddressPort     `json:"ap,omitempty"`
}

type AddressPort struct {
	Address *string `json:"a,omitempty"`
	Domain  *string `json:"d,omitempty"`
	Port    *uint32 `json:"p,omitempty"`
	Origin  *string `json:"o,omitempty"`
}

func (t *TagFirstOutbound) lookup(tags map[string]string) (*Outbound, bool) {
	for _, i := range t.Outbounds {
		if reflect.DeepEqual(i.Tags, tags) {
			return i, true
		}
	}
	return nil, false
}

func NewEmptyTagFirstOutboundView() *TagFirstVirtualOutboundView {
	return &TagFirstVirtualOutboundView{
		PerService: map[string]*TagFirstOutbound{},
	}
}

func NewTagFirstOutboundView(virtualOutboundMeshView *VirtualOutboundMeshView) *TagFirstVirtualOutboundView {
	if virtualOutboundMeshView == nil {
		return nil
	}

	view := NewEmptyTagFirstOutboundView()

	for he, vo := range virtualOutboundMeshView.byHostname {
		var domain *string
		if he.Type == Host || he.Type == FullyQualifiedDomain {
			if govalidator.IsDNSName(he.Name) {
				domain = pointer.To(he.Name)
			}
		}
		for _, oe := range vo.Outbounds {
			svc := oe.TagSet[mesh_proto.ServiceTag]
			port := getOrNil(oe.Port)
			origin := minifyOrigin(oe.Origin)
			tagFirst, ok := view.PerService[svc]
			if !ok {
				outbound := &Outbound{
					Tags: copyTagsWithoutSvc(oe.TagSet),
					AddressPorts: []AddressPort{{
						Address: &vo.Address,
						Domain:  domain,
						Port:    port,
						Origin:  pointer.To(origin),
					}},
				}
				view.PerService[svc] = &TagFirstOutbound{Outbounds: []*Outbound{outbound}}
			} else {
				cTags := copyTagsWithoutSvc(oe.TagSet)
				outbound, ok := tagFirst.lookup(cTags)
				if ok {
					outbound.AddressPorts = append(outbound.AddressPorts, AddressPort{
						Address: &vo.Address,
						Domain:  domain,
						Port:    port,
						Origin:  pointer.To(origin),
					})
				} else {
					tagFirst.Outbounds = append(tagFirst.Outbounds, &Outbound{
						Tags: cTags,
						AddressPorts: []AddressPort{{
							Address: &vo.Address,
							Domain:  domain,
							Port:    port,
							Origin:  pointer.To(origin),
						}},
					})
				}
			}
		}
	}

	return view
}

func (t *TagFirstVirtualOutboundView) ToVirtualOutboundView() *VirtualOutboundMeshView {
	view := NewEmptyVirtualOutboundView()

	for svc, tagFirstOutbound := range t.PerService {
		for _, outbound := range tagFirstOutbound.Outbounds {
			for _, ap := range outbound.AddressPorts {
				entryType := originToEntryType(pointer.Deref(ap.Origin))
				hostnameEntry := HostnameEntry{
					Type: entryType,
					Name: hostname(entryType, svc, ap),
				}

				if view.byHostname[hostnameEntry] == nil {
					view.byHostname[hostnameEntry] = &VirtualOutbound{
						Address: pointer.Deref(ap.Address),
						Outbounds: []OutboundEntry{{
							Port:   pointer.Deref(ap.Port),
							TagSet: copyTagsWithSvc(outbound.Tags, svc),
							Origin: deMinifyOrigin(pointer.Deref(ap.Origin)),
						}},
					}
				} else {
					view.byHostname[hostnameEntry].Outbounds = append(view.byHostname[hostnameEntry].Outbounds, OutboundEntry{
						Port:   pointer.Deref(ap.Port),
						TagSet: copyTagsWithSvc(outbound.Tags, svc),
						Origin: deMinifyOrigin(pointer.Deref(ap.Origin)),
					})
				}
			}
		}
	}

	return view
}

func copyTagsWithoutSvc(tags map[string]string) map[string]string {
	if tags == nil {
		return nil
	}
	stringMap := map[string]string{}
	maps.Copy(stringMap, tags)
	delete(stringMap, mesh_proto.ServiceTag)
	return stringMap
}

func copyTagsWithSvc(tags map[string]string, svc string) map[string]string {
	stringMap := map[string]string{}
	stringMap[mesh_proto.ServiceTag] = svc
	if tags == nil {
		return stringMap
	}
	maps.Copy(stringMap, tags)
	return stringMap
}

func minifyOrigin(origin string) string {
	switch origin {
	case OriginKube:
		return "k8s"
	case OriginService:
		return "svc"
	default:
		return origin
	}
}

func deMinifyOrigin(origin string) string {
	switch origin {
	case "k8s":
		return OriginKube
	case "svc":
		return OriginService
	default:
		return origin
	}
}

func originToEntryType(origin string) EntryType {
	if origin == "svc" {
		return Service
	}
	if origin == "k8s" || strings.HasPrefix(origin, HostPrefix) {
		return Host
	}
	if strings.HasPrefix(origin, VirtualOutboundPrefix) || strings.HasPrefix(origin, GatewayPrefix) {
		return FullyQualifiedDomain
	}
	return Service
}

func hostname(entryType EntryType, svc string, ap AddressPort) string {
	switch entryType {
	case Service:
		return svc
	case Host:
		return pointer.DerefOr(ap.Domain, pointer.Deref(ap.Address))
	case FullyQualifiedDomain:
		return pointer.DerefOr(ap.Domain, pointer.Deref(ap.Address))
	default:
		return svc
	}
}

func getOrNil(num uint32) *uint32 {
	if num == 0 {
		return nil
	} else {
		return pointer.To(num)
	}
}
