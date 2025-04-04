package export

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/api/system/v1alpha1"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	kuma_yaml "github.com/kumahq/kuma/app/kumactl/pkg/output/yaml"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/ca/provided/config"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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

			// Mesh resources
			var meshResource []model.Resource
			// resources created after Mesh
			var meshDependedResources []model.Resource
			// resources added last
			var lastResources []model.Resource
			// resource types are sorted
			for _, resDesc := range resTypes {
				switch resDesc.Scope {
				case model.ScopeGlobal:
					list := resDesc.NewList()
					if err := rs.List(cmd.Context(), list); err != nil {
						return errors.Wrapf(err, "could not list %q", resDesc.Name)
					}
					for _, res := range list.GetItems() {
						switch resDesc.Name {
						case core_mesh.MeshType:
							mesh := res.(*core_mesh.MeshResource)
							mesh.Spec.SkipCreatingInitialPolicies = []string{"*"}
							err := changeBuiltinBackendsToProvided(mesh)
							if err != nil {
								return nil
							}
							meshResource = append(meshResource, res)
							continue
						case core_system.GlobalSecretType:
							// filter out envoy-admin-ca and inter-cp-ca otherwise it will cause TLS handshake errors
							if res.GetMeta().GetName() == core_system.EnvoyAdminCA || res.GetMeta().GetName() == core_system.InterCpCA {
								continue
							}
							// put user token signing keys as last, because once we apply this, we cannot apply anything else without reconfiguring kumactl with a new auth data
							isUserTokenSigningKey := strings.HasPrefix(res.GetMeta().GetName(), core_system.UserTokenSigningKeyPrefix)
							if isUserTokenSigningKey {
								lastResources = append(lastResources, res)
								continue
							}
						}
						meshDependedResources = append(meshDependedResources, res)
					}
				case model.ScopeMesh:
					for _, mesh := range meshes.Items {
						list := resDesc.NewList()
						if err := rs.List(cmd.Context(), list, store.ListByMesh(mesh.GetMeta().GetName())); err != nil {
							return errors.Wrapf(err, "could not list %q", resDesc.Name)
						}
						meshDependedResources = append(meshDependedResources, list.GetItems()...)
					}
				}
			}

			allResources := append(meshResource, meshDependedResources...)
			allResources = append(allResources, lastResources...)

			switch ctx.args.format {
			case formatUniversal:
				for _, res := range allResources {
					// print mesh first since you cannot create other resources if there is no mesh
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
				yamlPrinter := kuma_yaml.NewPrinter()
				for _, res := range allResources {
					obj, err := k8sResources.Get(cmd.Context(), res.Descriptor(), res.GetMeta().GetName(), res.GetMeta().GetMesh())
					if err != nil {
						return err
					}
					if shouldSkipKubeObject(obj, ctx) {
						continue
					}

					cleanKubeObject(obj)
					switch res.Descriptor().Name {
					// only for the mesh we edit object by changing mtls backend from builtin to provided and adding skip initial resources
					case core_mesh.MeshType:
						result, err := model.ToMap(res.GetSpec())
						if err != nil {
							return err
						}
						obj["spec"] = result
					}
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

func changeBuiltinBackendsToProvided(res *core_mesh.MeshResource) error {
	if res.Spec.Mtls != nil {
		for _, backend := range res.Spec.Mtls.GetBackends() {
			switch backend.Type {
			case "builtin":
				cfg := &config.ProvidedCertificateAuthorityConfig{}
				cfg.Cert = &v1alpha1.DataSource{
					Type: &v1alpha1.DataSource_Secret{
						Secret: core_system.BuiltinCertSecretName(res.Meta.GetName(), backend.Name),
					},
				}
				cfg.Key = &v1alpha1.DataSource{
					Type: &v1alpha1.DataSource_Secret{
						Secret: core_system.BuiltinKeySecretName(res.Meta.GetName(), backend.Name),
					},
				}
				conf, err := util_proto.ToStruct(cfg)
				if err != nil {
					return err
				}
				backend.Type = "provided"
				backend.Conf = conf
				// we want to create secrets at any time
				res.Spec.Mtls.SkipValidation = true
			}
		}
	}
	return nil
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

	order := map[model.ResourceType]int{
		// we need secret for the mesh first
		core_system.SecretType: 0,
		// after secret is availabel we can add mesh with mTLS
		core_mesh.MeshType: 1,
		// we don't want user token to be added first
		core_system.GlobalSecretType: 99,
	}
	sort.SliceStable(resDescList, func(i, j int) bool {
		priorityI := 50
		priorityJ := 50
		if priority, exist := order[resDescList[i].Name]; exist {
			priorityI = priority
		}
		if priority, exist := order[resDescList[j].Name]; exist {
			priorityJ = priority
		}
		return priorityI < priorityJ
	})
	if len(incompatibleTypes) > 0 {
		msg := fmt.Sprintf("The following types won't be exported because they are unknown to kumactl: %s", strings.Join(incompatibleTypes, ","))
		cmd.Printf("# %s\n", msg)
		cmd.PrintErrf("WARNING: %s. Are you using a compatible version of kumactl?\n", msg)
	}
	return resDescList, nil
}
