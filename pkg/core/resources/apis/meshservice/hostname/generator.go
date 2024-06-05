package hostname

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"text/template"

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

func (g *MeshServiceHostnameGenerator) GenerateHostname(generator *hostnamegenerator_api.HostnameGeneratorResource, resource model.Resource) (string, error) {
	service, ok := resource.(*meshservice_api.MeshServiceResource)
	if !ok {
		return "", errors.Errorf("invalid resource type: expected=%T, got=%T", (*meshservice_api.MeshServiceResource)(nil), resource)
	}
	if !generator.Spec.Selector.MeshService.Matches(service.Meta.GetLabels()) {
		return "", nil
	}
	sb := strings.Builder{}
	tmpl := template.New("").Funcs(
		map[string]any{
			"label": func(key string) (string, error) {
				val, ok := service.GetMeta().GetLabels()[key]
				if !ok {
					return "", errors.Errorf("label %s not found", key)
				}
				return val, nil
			},
		},
	)
	tmpl, err := tmpl.Parse(generator.Spec.Template)
	if err != nil {
		return "", fmt.Errorf("failed compiling gotemplate error=%q", err.Error())
	}
	type meshedName struct {
		Name      string
		Namespace string
		Mesh      string
	}
	name := service.GetMeta().GetNameExtensions()[model.K8sNameComponent]
	if name == "" {
		name = service.GetMeta().GetName()
	}
	err = tmpl.Execute(&sb, meshedName{
		Name:      name,
		Namespace: service.GetMeta().GetNameExtensions()[model.K8sNamespaceComponent],
		Mesh:      service.GetMeta().GetMesh(),
	})
	if err != nil {
		return "", fmt.Errorf("pre evaluation of template with parameters failed with error=%q", err.Error())
	}
	return sb.String(), nil
}
