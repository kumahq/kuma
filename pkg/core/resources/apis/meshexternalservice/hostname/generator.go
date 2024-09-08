package hostname

import (
	"context"
	"reflect"

	"github.com/pkg/errors"

	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/hostname"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type MeshExternalServiceHostnameGenerator struct {
	resManager manager.ResourceManager
}

var _ hostname.HostnameGenerator = &MeshExternalServiceHostnameGenerator{}

func NewMeshExternalServiceHostnameGenerator(
	resManager manager.ResourceManager,
) *MeshExternalServiceHostnameGenerator {
	return &MeshExternalServiceHostnameGenerator{
		resManager: resManager,
	}
}

func (g *MeshExternalServiceHostnameGenerator) GetResources(ctx context.Context) (model.ResourceList, error) {
	resources := &meshexternalservice_api.MeshExternalServiceResourceList{}
	if err := g.resManager.List(ctx, resources); err != nil {
		return nil, errors.Wrap(err, "could not list MeshExternalServices")
	}
	return resources, nil
}

func (g *MeshExternalServiceHostnameGenerator) UpdateResourceStatus(ctx context.Context, resource model.Resource, statuses []hostnamegenerator_api.HostnameGeneratorStatus, addresses []hostnamegenerator_api.Address) error {
	externalService, ok := resource.(*meshexternalservice_api.MeshExternalServiceResource)
	if !ok {
		return errors.Errorf("invalid resource type: expected=%T, got=%T", (*meshexternalservice_api.MeshExternalServiceResource)(nil), resource)
	}
	externalService.Status.Addresses = addresses
	externalService.Status.HostnameGenerators = statuses
	if err := g.resManager.Update(ctx, externalService); err != nil {
		return errors.Wrap(err, "couldn't update MeshExternalService status")
	}
	return nil
}

func (g *MeshExternalServiceHostnameGenerator) HasStatusChanged(resource model.Resource, generatorStatuses []hostnamegenerator_api.HostnameGeneratorStatus, addresses []hostnamegenerator_api.Address) (bool, error) {
	es, ok := resource.(*meshexternalservice_api.MeshExternalServiceResource)
	if !ok {
		return false, errors.Errorf("invalid resource type: expected=%T, got=%T", (*meshexternalservice_api.MeshExternalServiceResource)(nil), resource)
	}
	return !reflect.DeepEqual(addresses, es.Status.Addresses) || !reflect.DeepEqual(generatorStatuses, es.Status.HostnameGenerators), nil
}

func (g *MeshExternalServiceHostnameGenerator) GenerateHostname(localZone string, generator *hostnamegenerator_api.HostnameGeneratorResource, resource model.Resource) (string, error) {
	es, ok := resource.(*meshexternalservice_api.MeshExternalServiceResource)
	if !ok {
		return "", errors.Errorf("invalid resource type: expected=%T, got=%T", (*meshexternalservice_api.MeshExternalServiceResource)(nil), resource)
	}
	if generator.Spec.Selector.MeshExternalService == nil {
		return "", nil
	}
	if !generator.Spec.Selector.MeshExternalService.Matches(es.Meta.GetLabels()) {
		return "", nil
	}
	return hostname.EvaluateTemplate(localZone, generator.Spec.Template, es.GetMeta())
}
