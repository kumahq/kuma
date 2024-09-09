package hostname

import (
	"context"
	"reflect"

	"github.com/pkg/errors"

	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/hostname"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type MeshServiceHostnameGenerator struct {
	resManager manager.ResourceManager
}

var _ hostname.HostnameGenerator = &MeshServiceHostnameGenerator{}

func NewMeshServiceHostnameGenerator(
	resManager manager.ResourceManager,
) *MeshServiceHostnameGenerator {
	return &MeshServiceHostnameGenerator{
		resManager: resManager,
	}
}

func (g *MeshServiceHostnameGenerator) GetResources(ctx context.Context) (model.ResourceList, error) {
	resources := &meshservice_api.MeshServiceResourceList{}
	if err := g.resManager.List(ctx, resources); err != nil {
		return nil, errors.Wrap(err, "could not list MeshServices")
	}
	return resources, nil
}

func (g *MeshServiceHostnameGenerator) UpdateResourceStatus(ctx context.Context, resource model.Resource, statuses []hostnamegenerator_api.HostnameGeneratorStatus, addresses []hostnamegenerator_api.Address) error {
	service, ok := resource.(*meshservice_api.MeshServiceResource)
	if !ok {
		return errors.Errorf("invalid resource type: expected=%T, got=%T", (*meshservice_api.MeshServiceResource)(nil), resource)
	}
	service.Status.Addresses = addresses
	service.Status.HostnameGenerators = statuses
	if err := g.resManager.Update(ctx, resource); err != nil {
		return errors.Wrap(err, "couldn't update MeshService status")
	}
	return nil
}

func (g *MeshServiceHostnameGenerator) HasStatusChanged(resource model.Resource, generatorStatuses []hostnamegenerator_api.HostnameGeneratorStatus, addresses []hostnamegenerator_api.Address) (bool, error) {
	service, ok := resource.(*meshservice_api.MeshServiceResource)
	if !ok {
		return false, errors.Errorf("invalid resource type: expected=%T, got=%T", (*meshservice_api.MeshServiceResource)(nil), resource)
	}

	return !reflect.DeepEqual(addresses, service.Status.Addresses) || !reflect.DeepEqual(generatorStatuses, service.Status.HostnameGenerators), nil
}

func (g *MeshServiceHostnameGenerator) GenerateHostname(localZone string, generator *hostnamegenerator_api.HostnameGeneratorResource, resource model.Resource) (string, error) {
	service, ok := resource.(*meshservice_api.MeshServiceResource)
	if !ok {
		return "", errors.Errorf("invalid resource type: expected=%T, got=%T", (*meshservice_api.MeshServiceResource)(nil), resource)
	}
	if generator.Spec.Selector.MeshService == nil {
		return "", nil
	}
	if !generator.Spec.Selector.MeshService.Matches(service.Meta.GetLabels()) {
		return "", nil
	}
	return hostname.EvaluateTemplate(localZone, generator.Spec.Template, service.GetMeta())
}
