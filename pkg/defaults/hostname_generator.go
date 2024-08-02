package defaults

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/core"
	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/util/k8s"
)

func EnsureHostnameGeneratorExists(ctx context.Context, resManager core_manager.ResourceManager, logger logr.Logger, cfg kuma_cp.Config) error {
	if cfg.Defaults.SkipHostnameGenerators {
		log.V(1).Info("skip ensuring default hostname generators because SkipHostnameGenerators is set to true")
		return nil
	}
	namespace := ""
	if cfg.Environment == core.KubernetesEnvironment {
		namespace = cfg.Store.Kubernetes.SystemNamespace
	}
	if cfg.Mode == core.Global {
		spec := hostnamegenerator_api.HostnameGenerator{
			Selector: hostnamegenerator_api.Selector{
				MeshMultiZoneService: &hostnamegenerator_api.LabelSelector{
					MatchLabels: map[string]string{
						mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
					},
				},
			},
			Template: "{{ .DisplayName }}.mzsvc.mesh.local",
		}
		if err := ensureHostnameGeneratorExist(ctx, resManager, log, "synced-mesh-multi-zone-service", namespace, spec); err != nil {
			return err
		}

		spec = hostnamegenerator_api.HostnameGenerator{
			Selector: hostnamegenerator_api.Selector{
				MeshService: &hostnamegenerator_api.LabelSelector{
					MatchLabels: map[string]string{
						mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
						metadata.HeadlessService:       "false",
						mesh_proto.ManagedByLabel:      "k8s-controller", // todo(jakubdyszkiewicz): consider introduing kuma.io/env instead
					},
				},
			},
			Template: `{{ .DisplayName }}.{{ .Namespace }}.svc.{{ .Zone }}.mesh.local`,
		}
		if err := ensureHostnameGeneratorExist(ctx, resManager, log, "synced-kube-mesh-service", namespace, spec); err != nil {
			return err
		}

		spec = hostnamegenerator_api.HostnameGenerator{
			Selector: hostnamegenerator_api.Selector{
				MeshService: &hostnamegenerator_api.LabelSelector{
					MatchLabels: map[string]string{
						mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
						metadata.HeadlessService:       "true",
						mesh_proto.ManagedByLabel:      "k8s-controller", // todo(jakubdyszkiewicz): consider introduing kuma.io/env instead
					},
				},
			},
			Template: `{{ label "statefulset.kubernetes.io/pod-name" }}.{{ label "k8s.kuma.io/service-name" }}.{{ .Namespace }}.svc.{{ .Zone }}.mesh.local`,
		}
		if err := ensureHostnameGeneratorExist(ctx, resManager, log, "synced-headless-kube-mesh-service", namespace, spec); err != nil {
			return err
		}

		spec = hostnamegenerator_api.HostnameGenerator{
			Selector: hostnamegenerator_api.Selector{
				MeshService: &hostnamegenerator_api.LabelSelector{
					MatchLabels: map[string]string{
						mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
						mesh_proto.ManagedByLabel:      "meshservice-generator", // todo(jakubdyszkiewicz): consider introduing kuma.io/env instead
					},
				},
			},
			Template: `{{ .DisplayName }}.svc.{{ .Zone }}.mesh.local`,
		}
		if err := ensureHostnameGeneratorExist(ctx, resManager, log, "synced-universal-mesh-service", namespace, spec); err != nil {
			return err
		}

		spec = hostnamegenerator_api.HostnameGenerator{
			Selector: hostnamegenerator_api.Selector{
				MeshExternalService: &hostnamegenerator_api.LabelSelector{
					MatchLabels: map[string]string{
						mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
					},
				},
			},
			Template: `{{ .DisplayName }}.extsvc.mesh.local`,
		}
		if err := ensureHostnameGeneratorExist(ctx, resManager, log, "synced-mesh-external-service", namespace, spec); err != nil {
			return err
		}
	}

	if cfg.Mode == core.Zone {
		spec := hostnamegenerator_api.HostnameGenerator{
			Selector: hostnamegenerator_api.Selector{
				MeshExternalService: &hostnamegenerator_api.LabelSelector{
					MatchLabels: map[string]string{
						mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
					},
				},
			},
			Template: `{{ .DisplayName }}.extsvc.mesh.local`,
		}
		if err := ensureHostnameGeneratorExist(ctx, resManager, log, "local-mesh-external-service", namespace, spec); err != nil {
			return err
		}

		if cfg.Environment == core.UniversalEnvironment {
			spec = hostnamegenerator_api.HostnameGenerator{
				Selector: hostnamegenerator_api.Selector{
					MeshService: &hostnamegenerator_api.LabelSelector{
						MatchLabels: map[string]string{
							mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
						},
					},
				},
				Template: `{{ .DisplayName }}.svc.mesh.local`,
			}
			if err := ensureHostnameGeneratorExist(ctx, resManager, log, "local-universal-mesh-service", namespace, spec); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureHostnameGeneratorExist(
	ctx context.Context,
	resManager core_manager.ResourceManager,
	logger logr.Logger,
	name string,
	namespace string,
	spec hostnamegenerator_api.HostnameGenerator,
) error {
	if namespace != "" {
		name = k8s.K8sNamespacedNameToCoreName(name, namespace)
	}
	hostnameGen := hostnamegenerator_api.NewHostnameGeneratorResource()
	err := resManager.Get(ctx, hostnameGen, core_store.GetByKey(name, core_model.NoMesh))
	switch {
	case err == nil:
		logger.V(1).Info("hostname generator already exist", "name", name)
		return nil
	case core_store.IsResourceNotFound(err):
		hostnameGen.Spec = &spec
		if err := resManager.Create(ctx, hostnameGen, core_store.CreateByKey(name, core_model.NoMesh)); err != nil {
			return errors.Wrapf(err, "could not create a hostname generator %q", name)
		}
		logger.Info("hostname generator created", "name", name)
		return nil
	default:
		return errors.Wrapf(err, "could not get hostname generator %q to verify if it exist", name)
	}
}
