package meshroute

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	util_k8s "github.com/kumahq/kuma/pkg/util/k8s"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type BackendRef struct {
	CoreName string
	Port     *uint32
	Kind     common_api.TargetRefKind
	Tags     map[string]string
	Weight   *uint
	Mesh     string
}

type BackendRefHash string

// Hash returns a hash of the BackendRef
func (in BackendRef) Hash() BackendRefHash {
	keys := maps.Keys(in.Tags)
	sort.Strings(keys)
	orderedTags := make([]string, 0, len(keys))
	for _, k := range keys {
		orderedTags = append(orderedTags, fmt.Sprintf("%s=%s", k, in.Tags[k]))
	}
	name := in.CoreName
	if in.Port != nil {
		name = fmt.Sprintf("%s_svc_%d", in.CoreName, *in.Port)
	}
	return BackendRefHash(fmt.Sprintf("%s/%s/%s/%s", in.Kind, name, strings.Join(orderedTags, "/"), in.Mesh))
}

func mapToSplitBackendRefs(refs []common_api.BackendRef) []BackendRef {
	var newRefs []BackendRef

	for _, ref := range refs {
		var coreName string

		if ref.Namespace != "" {
			coreName = util_k8s.K8sNamespacedNameToCoreName(ref.Name, ref.Namespace)
		} else {
			coreName = ref.Name
		}

		newRef := BackendRef{
			CoreName: coreName,
			Port:     ref.Port,
			Kind:     ref.Kind,
			Tags:     ref.Tags,
			Weight:   ref.Weight,
			Mesh:     ref.Mesh,
		}
		newRefs = append(newRefs, newRef)
	}

	return newRefs
}

func MakeTCPSplit(
	clusterCache map[BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []common_api.BackendRef,
	meshCtx xds_context.MeshContext,
) []envoy_common.Split {
	return makeSplit(
		map[core_mesh.Protocol]struct{}{
			core_mesh.ProtocolUnknown: {},
			core_mesh.ProtocolKafka:   {},
			core_mesh.ProtocolTCP:     {},
			core_mesh.ProtocolHTTP:    {},
			core_mesh.ProtocolHTTP2:   {},
			core_mesh.ProtocolGRPC:    {},
		},
		clusterCache,
		servicesAcc,
		mapToSplitBackendRefs(refs),
		meshCtx,
	)
}

func MakeHTTPSplit(
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []common_api.BackendRef,
	meshCtx xds_context.MeshContext,
) []envoy_common.Split {
	return makeSplit(
		map[core_mesh.Protocol]struct{}{
			core_mesh.ProtocolHTTP:  {},
			core_mesh.ProtocolHTTP2: {},
			core_mesh.ProtocolGRPC:  {},
		},
		clusterCache,
		servicesAcc,
		mapToSplitBackendRefs(refs),
		meshCtx,
	)
}

type DestinationService struct {
	Outbound    mesh_proto.OutboundInterface
	Protocol    core_mesh.Protocol
	ServiceName string
	BackendRef  common_api.BackendRef
}

func CollectServices(
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
) []DestinationService {
	var dests []DestinationService
	for _, outbound := range proxy.Dataplane.Spec.GetNetworking().GetOutbounds() {
		oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
		if outbound.BackendRef != nil {
			if outbound.GetAddress() == proxy.Dataplane.Spec.GetNetworking().GetAddress() {
				continue
			}
			ms, msOk := meshCtx.MeshServiceByName[outbound.BackendRef.Name]
			mes, mesOk := meshCtx.MeshExternalServiceByName[outbound.BackendRef.Name]
			if !msOk && !mesOk {
				// we want to ignore service which is not found. Logging might be excessive here.
				// We don't have other mechanism to bubble up warnings yet.
				continue
			}
			if msOk {
				port, ok := ms.FindPort(outbound.BackendRef.Port)
				if !ok {
					continue
				}
				protocol := core_mesh.Protocol(core_mesh.ProtocolTCP)
				if port.AppProtocol != "" {
					protocol = port.AppProtocol
				}
				ns := ""
				if ms.GetMeta().GetNameExtensions() != nil {
					ns = ms.GetMeta().GetNameExtensions()[model.K8sNamespaceComponent]
				}
				dests = append(dests, DestinationService{
					Outbound:    oface,
					Protocol:    protocol,
					ServiceName: ms.DestinationName(outbound.BackendRef.Port),
					BackendRef: common_api.BackendRef{
						TargetRef: common_api.TargetRef{
							Kind:      common_api.MeshService,
							Name:      ms.GetMeta().GetName(),
							Namespace: ns,
						},
						Port: &port.Port,
					},
				})
			}
			if mesOk {
				port := mes.Spec.Match.Port
				protocol := mes.Spec.Match.Protocol
				ns := ""
				if mes.GetMeta().GetNameExtensions() != nil {
					ns = mes.GetMeta().GetNameExtensions()[model.K8sNamespaceComponent]
				}
				dests = append(dests, DestinationService{
					Outbound:    oface,
					Protocol:    core_mesh.Protocol(protocol),
					ServiceName: mes.DestinationName(outbound.BackendRef.Port),
					BackendRef: common_api.BackendRef{
						TargetRef: common_api.TargetRef{
							Kind:      common_api.MeshExternalService,
							Name:      mes.GetMeta().GetName(),
							Namespace: ns,
						},
						Port: pointer.To(uint32(port)),
					},
				})
			}
		} else {
			serviceName := outbound.GetService()
			dests = append(dests, DestinationService{
				Outbound:    oface,
				Protocol:    meshCtx.GetServiceProtocol(serviceName),
				ServiceName: serviceName,
				BackendRef: common_api.BackendRef{
					TargetRef: common_api.TargetRef{
						Kind: common_api.MeshService,
						Name: serviceName,
						Tags: outbound.GetTags(),
					},
				},
			})
		}
	}
	return dests
}

func makeSplit[T ~string](
	protocols map[core_mesh.Protocol]struct{},
	clusterCache map[T]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []BackendRef,
	meshCtx xds_context.MeshContext,
) []envoy_common.Split {
	var split []envoy_common.Split

	for _, ref := range refs {
		switch ref.Kind {
		case common_api.MeshService, common_api.MeshExternalService, common_api.MeshServiceSubset:
		default:
			continue
		}

		var service string
		var protocol core_mesh.Protocol
		if pointer.DerefOr(ref.Weight, 1) == 0 {
			continue
		}
		var meshServiceName string
		switch {
		case ref.Kind == common_api.MeshExternalService:
			mes, ok := meshCtx.MeshExternalServiceByName[ref.CoreName]
			if !ok {
				continue
			}
			port := pointer.Deref(ref.Port)
			service = mes.DestinationName(port)
			protocol = meshCtx.GetServiceProtocol(service)
		case ref.Port != nil: // in this case, reference real MeshService instead of kuma.io/service tag
			ms, ok := meshCtx.MeshServiceByName[ref.CoreName]
			if !ok {
				continue
			}
			meshServiceName = ms.GetMeta().GetName()
			port, ok := ms.FindPort(*ref.Port)
			if !ok {
				continue
			}
			service = ms.DestinationName(*ref.Port)
			protocol = port.AppProtocol // todo(jakubdyszkiewicz): do we need to default to TCP or will this be done by MeshService defaulter?
		default:
			service = ref.CoreName
			protocol = meshCtx.GetServiceProtocol(service)
		}
		if _, ok := protocols[protocol]; !ok {
			return nil
		}

		var clusterName string
		if ref.Kind == common_api.MeshExternalService {
			clusterName = envoy_names.GetMeshExternalServiceName(ref.CoreName)
		} else {
			clusterName, _ = envoy_tags.Tags(ref.Tags).
				WithTags(mesh_proto.ServiceTag, service).
				DestinationClusterName(nil)
		}

		// The mesh tag is present here if this destination is generated
		// from a cross-mesh MeshGateway listener virtual outbound.
		// It is not part of the service tags.
		if mesh, ok := ref.Tags[mesh_proto.MeshTag]; ok {
			// The name should be distinct to the service & mesh combination
			clusterName = fmt.Sprintf("%s_%s", clusterName, mesh)
		}

		isExternalService := meshCtx.IsExternalService(service)
		refHash := T(ref.Hash())

		if existingClusterName, ok := clusterCache[refHash]; ok {
			// cluster already exists, so adding only split
			split = append(split, plugins_xds.NewSplitBuilder().
				WithClusterName(existingClusterName).
				WithWeight(uint32(pointer.DerefOr(ref.Weight, 1))).
				WithExternalService(isExternalService).
				Build())
			continue
		}

		clusterCache[refHash] = clusterName

		split = append(split, plugins_xds.NewSplitBuilder().
			WithClusterName(clusterName).
			WithWeight(uint32(pointer.DerefOr(ref.Weight, 1))).
			WithExternalService(isExternalService).
			Build())

		clusterBuilder := plugins_xds.NewClusterBuilder().
			WithService(service).
			WithName(clusterName).
			WithTags(envoy_tags.Tags(ref.Tags).
				WithTags(mesh_proto.ServiceTag, ref.CoreName).
				WithoutTags(mesh_proto.MeshTag)).
			WithExternalService(isExternalService)

		if mesh, ok := ref.Tags[mesh_proto.MeshTag]; ok {
			clusterBuilder.WithMesh(mesh)
		}

		if len(meshServiceName) > 0 {
			servicesAcc.AddMeshService(meshServiceName, *ref.Port, clusterBuilder.Build())
		} else {
			servicesAcc.Add(clusterBuilder.Build())
		}
	}

	return split
}
