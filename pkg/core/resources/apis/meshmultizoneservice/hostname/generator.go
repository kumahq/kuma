package hostname

import (
	"context"
	"reflect"

	"github.com/pkg/errors"

	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/hostname"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type MeshMultiZoneServiceHostnameGenerator struct {
	resManager manager.ResourceManager
}

var _ hostname.HostnameGenerator = &MeshMultiZoneServiceHostnameGenerator{}

func NewMeshMultiZoneServiceHostnameGenerator(
	resManager manager.ResourceManager,
) *MeshMultiZoneServiceHostnameGenerator {
	return &MeshMultiZoneServiceHostnameGenerator{
		resManager: resManager,
	}
}

func (g *MeshMultiZoneServiceHostnameGenerator) GetResources(ctx context.Context) (model.ResourceList, error) {
	resources := &meshmzservice_api.MeshMultiZoneServiceResourceList{}
	if err := g.resManager.List(ctx, resources); err != nil {
		return nil, errors.Wrap(err, "could not list MeshMultiZoneServices")
	}
	return resources, nil
}

func (g *MeshMultiZoneServiceHostnameGenerator) UpdateResourceStatus(ctx context.Context, resource model.Resource, statuses []hostnamegenerator_api.HostnameGeneratorStatus, addresses []hostnamegenerator_api.Address) error {
	service, ok := resource.(*meshmzservice_api.MeshMultiZoneServiceResource)
	if !ok {
		return errors.Errorf("invalid resource type: expected=%T, got=%T", (*meshmzservice_api.MeshMultiZoneServiceResource)(nil), resource)
	}
	service.Status.Addresses = addresses
	service.Status.HostnameGenerators = statuses
	if err := g.resManager.Update(ctx, resource); err != nil {
		return errors.Wrap(err, "couldn't update MeshMultiZoneService status")
	}
	return nil
}

func (g *MeshMultiZoneServiceHostnameGenerator) HasStatusChanged(resource model.Resource, generatorStatuses []hostnamegenerator_api.HostnameGeneratorStatus, addresses []hostnamegenerator_api.Address) (bool, error) {
	service, ok := resource.(*meshmzservice_api.MeshMultiZoneServiceResource)
	if !ok {
		return false, errors.Errorf("invalid resource type: expected=%T, got=%T", (*meshmzservice_api.MeshMultiZoneServiceResource)(nil), resource)
	}

	return !reflect.DeepEqual(addresses, service.Status.Addresses) || !reflect.DeepEqual(generatorStatuses, service.Status.HostnameGenerators), nil
}

func (g *MeshMultiZoneServiceHostnameGenerator) GenerateHostname(localZone string, generator *hostnamegenerator_api.HostnameGeneratorResource, resource model.Resource) (string, error) {
	service, ok := resource.(*meshmzservice_api.MeshMultiZoneServiceResource)
	if !ok {
		return "", errors.Errorf("invalid resource type: expected=%T, got=%T", (*meshmzservice_api.MeshMultiZoneServiceResource)(nil), resource)
	}
	if generator.Spec.Selector.MeshMultiZoneService == nil {
		return "", nil
	}
	if !generator.Spec.Selector.MeshMultiZoneService.Matches(service.Meta.GetLabels()) {
		return "", nil
	}
	return hostname.EvaluateTemplate(localZone, generator.Spec.Template, service.GetMeta())
}
