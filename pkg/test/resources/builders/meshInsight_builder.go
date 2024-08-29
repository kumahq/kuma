package builders

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type MeshInsightBuilder struct {
	res *mesh.MeshInsightResource
}

func MeshInsight() *MeshInsightBuilder {
	return &MeshInsightBuilder{
		res: &mesh.MeshInsightResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.NoMesh,
				Name: "meshInsight-1",
			},
			Spec: &mesh_proto.MeshInsight{
				DataplanesByType: &mesh_proto.MeshInsight_DataplanesByType{
					Standard:         &mesh_proto.MeshInsight_DataplaneStat{},
					GatewayBuiltin:   &mesh_proto.MeshInsight_DataplaneStat{},
					GatewayDelegated: &mesh_proto.MeshInsight_DataplaneStat{},
				},
				Policies:  map[string]*mesh_proto.MeshInsight_PolicyStat{},
				Resources: map[string]*mesh_proto.MeshInsight_ResourceStat{},
			},
		},
	}
}

func (mi *MeshInsightBuilder) Build() *mesh.MeshInsightResource {
	return mi.res
}

func (mi *MeshInsightBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), mi.Build(), store.CreateBy(mi.Key()))
}

func (mi *MeshInsightBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(mi.res.GetMeta())
}

func (mi *MeshInsightBuilder) WithName(name string) *MeshInsightBuilder {
	mi.res.Meta.(*test_model.ResourceMeta).Name = name
	return mi
}

func (mi *MeshInsightBuilder) WithStandardDataplaneStats(
	online uint32,
	offline uint32,
	partiallyDegraded uint32,
	total uint32,
) *MeshInsightBuilder {
	mi.res.Spec.DataplanesByType.Standard.Online = online
	mi.res.Spec.DataplanesByType.Standard.Offline = offline
	mi.res.Spec.DataplanesByType.Standard.PartiallyDegraded = partiallyDegraded
	mi.res.Spec.DataplanesByType.Standard.Total = total
	return mi
}

func (mi *MeshInsightBuilder) WithBuiltinGatewayDataplaneStats(
	online uint32,
	offline uint32,
	partiallyDegraded uint32,
	total uint32,
) *MeshInsightBuilder {
	mi.res.Spec.DataplanesByType.GatewayBuiltin.Online = online
	mi.res.Spec.DataplanesByType.GatewayBuiltin.Offline = offline
	mi.res.Spec.DataplanesByType.GatewayBuiltin.PartiallyDegraded = partiallyDegraded
	mi.res.Spec.DataplanesByType.GatewayBuiltin.Total = total
	return mi
}

func (mi *MeshInsightBuilder) WithDelegatedGatewayDataplaneStats(
	online uint32,
	offline uint32,
	partiallyDegraded uint32,
	total uint32,
) *MeshInsightBuilder {
	mi.res.Spec.DataplanesByType.GatewayDelegated.Online = online
	mi.res.Spec.DataplanesByType.GatewayDelegated.Offline = offline
	mi.res.Spec.DataplanesByType.GatewayDelegated.PartiallyDegraded = partiallyDegraded
	mi.res.Spec.DataplanesByType.GatewayDelegated.Total = total
	return mi
}

func (mi *MeshInsightBuilder) AddPolicyStats(policy string, total uint32) *MeshInsightBuilder {
	mi.res.Spec.Policies[policy] = &mesh_proto.MeshInsight_PolicyStat{Total: total}
	return mi
}

func (mi *MeshInsightBuilder) AddResourceStats(resource string, total uint32) *MeshInsightBuilder {
	mi.res.Spec.Resources[resource] = &mesh_proto.MeshInsight_ResourceStat{Total: total}
	return mi
}
