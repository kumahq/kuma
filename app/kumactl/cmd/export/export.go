package export

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/yaml"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type exportContext struct {
	*kumactl_cmd.RootContext

	args struct {
		profile         string
		format          string
		systemNamespace string
		includeAdmin    bool
	}
}

const (
	profileFederation             = "federation"
	profileFederationWithPolicies = "federation-with-policies"
	profileAll                    = "all"
	profileNoDataplanes           = "no-dataplanes"

	formatUniversal  = "universal"
	formatKubernetes = "kubernetes"
)

var allProfiles = []string{
	profileAll, profileFederation, profileFederationWithPolicies, profileNoDataplanes,
}

func IsMigrationProfile(profile string) bool {
	return slices.Contains([]string{profileFederation, profileFederationWithPolicies}, profile)
}

func NewExportCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &exportContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export Kuma resources",
		Long:  `Export Kuma resources.`,
		Example: `
Export Kuma resources
$ kumactl export --profile federation --format universal > policies.yaml
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			version := kumactl_cmd.CheckCompatibility(pctx.FetchServerVersion, cmd.ErrOrStderr())
			if version != nil {
				cmd.Printf("# Product: %s, Version: %s, Hostname: %s, ClusterId: %s, InstanceId: %s\n",
					version.Product, version.Version, version.Hostname, version.ClusterId, version.InstanceId)
			}

			if !slices.Contains(allProfiles, ctx.args.profile) {
				return fmt.Errorf("invalid profile: %q", ctx.args.profile)
			}

			if !slices.Contains([]string{formatKubernetes, formatUniversal}, ctx.args.format) {
				return fmt.Errorf("invalid format: %q", ctx.args.format)
			}

			resTypes, err := resourcesTypesToDump(cmd, ctx)
			if err != nil {
				return err
			}

			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			meshes := &core_mesh.MeshResourceList{}
			if err := rs.List(cmd.Context(), meshes); err != nil {
				return errors.Wrap(err, "could not list meshes")
			}

			var meshSecrets []model.Resource
			var otherResources []model.Resource
			var meshesResources []model.Resource
			for _, resDesc := range resTypes {
				if resDesc.Scope == model.ScopeGlobal {
					list := resDesc.NewList()
					if err := rs.List(cmd.Context(), list); err != nil {
						return errors.Wrapf(err, "could not list %q", resDesc.Name)
					}
					for _, res := range list.GetItems() {
						if res.Descriptor().Name == core_mesh.MeshType {
							mesh := res.(*core_mesh.MeshResource)
							mesh.Spec.SkipCreatingInitialPolicies = []string{"*"}
							meshesResources = append(meshesResources, res)
						} else {
							otherResources = append(otherResources, res)
						}
					}
				} else {
					for _, mesh := range meshes.Items {
						list := resDesc.NewList()
						if err := rs.List(cmd.Context(), list, store.ListByMesh(mesh.GetMeta().GetName())); err != nil {
							return errors.Wrapf(err, "could not list %q", resDesc.Name)
						}
						for _, res := range list.GetItems() {
							if res.Descriptor().Name == core_system.SecretType {
								meshSecrets = append(meshSecrets, res)
							} else {
								otherResources = append(otherResources, res)
							}
						}
					}
				}
			}

			allResources := append(meshSecrets, otherResources...)
			var resources []model.Resource
			var userTokenSigningKeys []model.Resource
			// filter out envoy-admin-ca and inter-cp-ca otherwise it will cause TLS handshake errors
			for _, res := range allResources {
				isUserTokenSigningKey := strings.HasPrefix(res.GetMeta().GetName(), core_system.UserTokenSigningKeyPrefix)
				if res.GetMeta().GetName() != core_system.EnvoyAdminCA &&
					res.GetMeta().GetName() != core_system.InterCpCA &&
					!isUserTokenSigningKey {
					resources = append(resources, res)
				}
				if isUserTokenSigningKey {
					userTokenSigningKeys = append(userTokenSigningKeys, res)
				}
			}
			// put user token signing keys as last, because once we apply this, we cannot apply anything else without reconfiguring kumactl with a new auth data
			resources = append(resources, userTokenSigningKeys...)

			switch ctx.args.format {
			case formatUniversal:
				for _, res := range meshesResources {
					// print mesh first since you cannot create other resources if there is no mesh
					if _, err := cmd.OutOrStdout().Write([]byte("---\n")); err != nil {
						return err
					}
					if err := printers.GenericPrint(output.YAMLFormat, res, table.Table{}, cmd.OutOrStdout()); err != nil {
						return err
					}
				}
				for _, res := range resources {
					if _, err := cmd.OutOrStdout().Write([]byte("---\n")); err != nil {
						return err
					}
					if err := printers.GenericPrint(output.YAMLFormat, res, table.Table{}, cmd.OutOrStdout()); err != nil {
						return err
					}
				}
			case formatKubernetes:
				k8sResources, err := pctx.CurrentKubernetesResourcesClient()
				if err != nil {
					return err
				}
				yamlPrinter := yaml.NewPrinter()
				for _, res := range append(meshesResources, resources...) {
					obj, err := k8sResources.Get(cmd.Context(), res.Descriptor(), res.GetMeta().GetName(), res.GetMeta().GetMesh())
					if err != nil {
						return err
					}
					if shouldSkipKubeObject(obj, ctx) {
						continue
					}
					cleanKubeObject(obj)
					if err := yamlPrinter.Print(obj, cmd.OutOrStdout()); err != nil {
						return err
					}
				}
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&ctx.args.profile, "profile", "p", profileFederation, fmt.Sprintf(`Profile. Available values: %s`, strings.Join(allProfiles, ",")))
	cmd.Flags().StringVarP(&ctx.args.format, "format", "f", formatUniversal, fmt.Sprintf(`Policy format output. Available values: %q, %q`, formatUniversal, formatKubernetes))
	cmd.Flags().StringVarP(&ctx.args.systemNamespace, "system-namespace", "n", ctx.InstallCpContext.Args.Namespace, "Define namespace in which control-plane was installed")
	cmd.Flags().BoolVarP(&ctx.args.includeAdmin, "include-admin", "a", false, "Include admin resource types (like secrets), this flag is ignored on migration profiles like federation as these entities are required")
	return cmd
}

func shouldSkipKubeObject(obj map[string]interface{}, ectx *exportContext) bool {
	if ectx.args.profile == profileAll {
		return false
	}
	metadata, ok := obj["metadata"]
	if !ok {
		return false
	}
	meta := metadata.(map[string]interface{})
	if ns, found := meta["namespace"]; found {
		// we can't apply non system namespace resource on global
		return ns != ectx.args.systemNamespace
	}
	return false
}

// cleans kubernetes object, so it can be applied on any other cluster
func cleanKubeObject(obj map[string]interface{}) {
	metadata, ok := obj["metadata"]
	if !ok {
		return
	}
	meta := metadata.(map[string]interface{})
	delete(meta, "resourceVersion")
	delete(meta, "ownerReferences")
	delete(meta, "uid")
	delete(meta, "generation")
	delete(meta, "managedFields")
}

func resourcesTypesToDump(cmd *cobra.Command, ectx *exportContext) ([]model.ResourceTypeDescriptor, error) {
	client, err := ectx.CurrentResourcesListClient()
	if err != nil {
		return nil, err
	}
	list, err := client.List(cmd.Context())
	if err != nil {
		return nil, err
	}
	var resDescList []model.ResourceTypeDescriptor
	var incompatibleTypes []string
	for _, res := range list.Resources {
		resDesc, err := ectx.Runtime.Registry.DescriptorFor(model.ResourceType(res.Name))
		if err != nil {
			incompatibleTypes = append(incompatibleTypes, res.Name)
			continue
		}
		if resDesc.AdminOnly && !IsMigrationProfile(ectx.args.profile) && !ectx.args.includeAdmin {
			continue
		}
		// For each profile remove types we don't want
		switch ectx.args.profile {
		case profileFederation:
			if !res.IncludeInFederation { // base decision on `IncludeInFederation` field
				continue
			}
			if res.Policy != nil && res.Policy.IsTargetRef { // do not include new policies
				continue
			}
			if res.Name == string(core_mesh.MeshGatewayType) { // do not include MeshGateways
				continue
			}
		case profileFederationWithPolicies:
			if !res.IncludeInFederation {
				continue
			}
		case profileNoDataplanes:
			if resDesc.Name == core_mesh.DataplaneType || resDesc.Name == core_mesh.DataplaneInsightType {
				continue
			}
		}
		resDescList = append(resDescList, resDesc)
	}
	if len(incompatibleTypes) > 0 {
		msg := fmt.Sprintf("The following types won't be exported because they are unknown to kumactl: %s", strings.Join(incompatibleTypes, ","))
		cmd.Printf("# %s\n", msg)
		cmd.PrintErrf("WARNING: %s. Are you using a compatible version of kumactl?\n", msg)
	}
	return resDescList, nil
}
