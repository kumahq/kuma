package install

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/strvals"
	"k8s.io/client-go/discovery"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"

	install_context "github.com/kumahq/kuma/app/kumactl/cmd/install/context"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

type componentVersion struct {
	args    *install_context.InstallControlPlaneArgs
	version string
}

func (cv *componentVersion) String() string {
	return cv.version
}

func (cv *componentVersion) Set(v string) error {
	cv.version = v
	cv.args.ControlPlane_image_tag = v
	cv.args.DataPlane_image_tag = v
	cv.args.DataPlane_initImage_tag = v
	cv.args.Cni_image_tag = v
	return nil
}

func (cv *componentVersion) Type() string {
	return "string"
}

// getVersionSet retrieves a set of available k8s API versions
// This is inlined from https://github.com/helm/helm/blob/v3.8.1/pkg/action/action.go#L309
func getVersionSet(client discovery.ServerResourcesInterface) (chartutil.VersionSet, error) {
	groups, resources, err := client.ServerGroupsAndResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return chartutil.DefaultVersionSet, errors.Wrap(err, "could not get apiVersions from Kubernetes")
	}

	if len(groups) == 0 && len(resources) == 0 {
		return chartutil.DefaultVersionSet, nil
	}

	versionMap := make(map[string]interface{})
	versions := []string{}

	// Extract the groups
	for _, g := range groups {
		for _, gv := range g.Versions {
			versionMap[gv.GroupVersion] = struct{}{}
		}
	}

	// Extract the resources
	var id string
	var ok bool
	for _, r := range resources {
		for _, rl := range r.APIResources {
			// A Kind at a GroupVersion can show up more than once. We only want
			// it displayed once in the final output.
			id = path.Join(r.GroupVersion, rl.Kind)
			if _, ok = versionMap[id]; !ok {
				versionMap[id] = struct{}{}
			}
		}
	}

	// Convert to a form that NewVersionSet can use
	for k := range versionMap {
		versions = append(versions, k)
	}

	return chartutil.VersionSet(versions), nil
}

// getCapabilities builds a Capabilities from discovery information.
// This is inlined from https://github.com/helm/helm/blob/v3.8.1/pkg/action/action.go#L239
func getCapabilities(kubeConfig *rest.Config) (*chartutil.Capabilities, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	kubeVersion, err := dc.ServerVersion()
	if err != nil {
		return nil, errors.Wrap(err, "could not get server version from Kubernetes")
	}

	// Issue #6361:
	// Client-Go emits an error when an API service is registered but unimplemented.
	// We ignore that error here. But since the discovery client continues
	// building the API object, it is correctly populated with all valid APIs.
	// See https://github.com/kubernetes/kubernetes/issues/72051#issuecomment-521157642
	apiVersions, err := getVersionSet(dc)
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return nil, errors.Wrap(err, "could not get apiVersions from Kubernetes")
	}

	return &chartutil.Capabilities{
		APIVersions: apiVersions,
		KubeVersion: chartutil.KubeVersion{
			Version: kubeVersion.GitVersion,
			Major:   kubeVersion.Major,
			Minor:   kubeVersion.Minor,
		},
	}, nil
}

func newInstallControlPlaneCmd(ctx *install_context.InstallCpContext) *cobra.Command {
	args := ctx.Args
	cmd := &cobra.Command{
		Use:   "control-plane",
		Short: "Install Kuma Control Plane on Kubernetes",
		Long: `Install Kuma Control Plane on Kubernetes in its own namespace.
This command requires that the KUBECONFIG environment is set`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			mesh_k8s.RegisterK8sGatewayTypes()

			if args.ExperimentalGatewayAPI {
				mesh_k8s.RegisterK8sGatewayAPITypes()
			}

			templateFiles, err := ctx.InstallCpTemplateFiles(&args)
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}
			if args.DumpValues {
				fList := templateFiles.Filter(func(file data.File) bool {
					return file.FullPath == "values.yaml"
				})
				if len(fList) != 1 {
					return errors.New("More than one file 'values.yaml'")
				}
				_, err = cmd.OutOrStdout().Write(fList[0].Data)
				return err
			}

			// Inline parameters
			vals := generateOverrideValues(args, ctx.HELMValuesPrefix)

			// User specified a values files via -f/--values
			for _, filePath := range args.ValueFiles {
				currentMap := map[string]interface{}{}

				var bytes []byte
				if strings.TrimSpace(filePath) == "-" {
					bytes, err = io.ReadAll(cmd.InOrStdin())
				} else {
					bytes, err = os.ReadFile(filePath)
				}
				if err != nil {
					return err
				}

				if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
					return errors.Wrapf(err, "failed to parse %s", filePath)
				}
				// Merge with the previous map
				vals = mergeMaps(vals, currentMap)
			}

			// User specified a value via --set
			for _, value := range args.Values {
				if err := strvals.ParseInto(value, vals); err != nil {
					return errors.Wrap(err, "failed parsing --set data")
				}
			}
			if err != nil {
				return errors.Wrap(err, "Failed to evaluate helm values")
			}

			if args.UseNodePort && args.ControlPlane_mode == config_core.Global {
				v := "controlPlane.globalZoneSyncService.type=NodePort"
				if ctx.HELMValuesPrefix != "" {
					v = fmt.Sprintf("%s.%s", ctx.HELMValuesPrefix, v)
				}
				if err := strvals.ParseInto(v, vals); err != nil {
					return errors.Wrap(err, "Failed using NodePort")
				}
			}

			if args.IngressUseNodePort {
				v := "ingress.service.type=NodePort"
				if ctx.HELMValuesPrefix != "" {
					v = fmt.Sprintf("%s.%s", ctx.HELMValuesPrefix, v)
				}
				if err := strvals.ParseInto(v, vals); err != nil {
					return errors.Wrap(err, "Failed using NodePort for ingress")
				}
			}

			var kubeClientConfig *rest.Config
			if !args.WithoutKubernetesConnection {
				var err error
				kubeClientConfig, err = k8s.DefaultClientConfig("", "")
				if err != nil {
					return errors.Wrap(err, "could not detect Kubernetes configuration")
				}
			}

			capabilities := chartutil.DefaultCapabilities
			if !args.WithoutKubernetesConnection {
				capabilities, err = getCapabilities(kubeClientConfig)
				if err != nil {
					return errors.Wrap(err, "could not get Capabilities")
				}
			}
			for _, version := range args.APIVersions {
				capabilities.APIVersions = append(capabilities.APIVersions, version)
			}

			renderedFiles, err := renderHelmFiles(templateFiles, args.Namespace, vals, kubeClientConfig, *capabilities)
			if err != nil {
				return errors.Wrap(err, "Failed to render helm template files")
			}

			sortedResources, err := k8s.SortResourcesByKind(renderedFiles, args.SkipKinds...)
			if err != nil {
				return errors.Wrap(err, "Failed to sort resources by kind")
			}

			singleFile := data.JoinYAML(sortedResources)

			if _, err := cmd.OutOrStdout().Write(singleFile.Data); err != nil {
				return errors.Wrap(err, "Failed to output rendered resources")
			}

			return nil
		},
	}
	componentVersion := componentVersion{
		args: &args,
	}
	// flags
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install Kuma Control Plane to")

	cmd.Flags().Var(&componentVersion, "version", "version of Kuma Control Plane components")

	cmd.Flags().StringVar(&args.Image_registry, "registry", args.Image_registry, "registry for all images")
	cmd.Flags().StringVar(&args.ControlPlane_image_pullPolicy, "image-pull-policy", args.ControlPlane_image_pullPolicy, "image pull policy that applies to all components of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.ControlPlane_image_registry, "control-plane-registry", args.ControlPlane_image_registry, "registry for the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlane_image_repository, "control-plane-repository", args.ControlPlane_image_repository, "repository for the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlane_image_tag, "control-plane-version", args.ControlPlane_image_tag, "version of the image of the Kuma Control Plane component")
	cmd.Flags().StringVar(&args.ControlPlane_service_name, "control-plane-service-name", args.ControlPlane_service_name, "Service name of the Kuma Control Plane")
	cmd.Flags().StringVar(&args.ControlPlane_tls_general_secret, "tls-general-secret", args.ControlPlane_tls_general_secret, "Secret that contains tls.crt, tls.key [and ca.crt when no --tls-general-ca-secret specified] for protecting Kuma in-cluster communication")
	cmd.Flags().StringVar(&args.ControlPlane_tls_general_ca_secret, "tls-general-ca-secret", args.ControlPlane_tls_general_ca_secret, "Secret that contains ca.crt that was used to sign cert for protecting Kuma in-cluster communication (ca.crt present in this secret have precedence over the one provided in --tls-general-secret)")
	cmd.Flags().StringVar(&args.ControlPlane_tls_general_caBundle, "tls-general-ca-bundle", args.ControlPlane_tls_general_secret, "Base64 encoded CA certificate (the same as in controlPlane.tls.general.secret#ca.crt)")
	cmd.Flags().StringVar(&args.ControlPlane_tls_apiServer_secret, "tls-api-server-secret", args.ControlPlane_tls_apiServer_secret, "Secret that contains tls.crt, tls.key for protecting Kuma API on HTTPS")
	cmd.Flags().StringVar(&args.ControlPlane_tls_apiServer_clientCertsSecret, "tls-api-server-client-certs-secret", args.ControlPlane_tls_apiServer_clientCertsSecret, "Secret that contains list of .pem certificates that can access admin endpoints of Kuma API on HTTPS")
	cmd.Flags().StringVar(&args.ControlPlane_tls_kdsGlobalServer_secret, "tls-kds-global-server-secret", args.ControlPlane_tls_kdsGlobalServer_secret, "Secret that contains tls.crt, tls.key for protecting cross cluster communication")
	cmd.Flags().StringVar(&args.ControlPlane_tls_kdsZoneClient_secret, "tls-kds-zone-client-secret", args.ControlPlane_tls_kdsZoneClient_secret, "Secret that contains ca.crt which was used to sign KDS Global server. Used for CP verification")
	cmd.Flags().StringVar(&args.ControlPlane_injectorFailurePolicy, "injector-failure-policy", args.ControlPlane_injectorFailurePolicy, "failure policy of the mutating web hook implemented by the Kuma Injector component")
	cmd.Flags().StringToStringVar(&args.ControlPlane_nodeSelector, "control-plane-node-selector", args.ControlPlane_nodeSelector, "node selector for Kuma Control Plane")
	cmd.Flags().StringToStringVar(&args.ControlPlane_envVars, "env-var", args.ControlPlane_envVars, "environment variables that will be passed to the control plane")
	cmd.Flags().StringVar(&args.DataPlane_image_registry, "dataplane-registry", args.DataPlane_image_registry, "registry for the image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_image_repository, "dataplane-repository", args.DataPlane_image_repository, "repository for the image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_image_tag, "dataplane-version", args.DataPlane_image_tag, "version of the image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_initImage_registry, "dataplane-init-registry", args.DataPlane_initImage_registry, "registry for the init image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_initImage_repository, "dataplane-init-repository", args.DataPlane_initImage_repository, "repository for the init image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.DataPlane_initImage_tag, "dataplane-init-version", args.DataPlane_initImage_tag, "version of the init image of the Kuma DataPlane component")
	cmd.Flags().StringVar(&args.ControlPlane_kdsGlobalAddress, "kds-global-address", args.ControlPlane_kdsGlobalAddress, "URL of Global Kuma CP (example: grpcs://192.168.0.1:5685)")
	cmd.Flags().BoolVar(&args.Cni_enabled, "cni-enabled", args.Cni_enabled, "install Kuma with CNI instead of proxy init container")
	cmd.Flags().BoolVar(&args.Cni_chained, "cni-chained", args.Cni_chained, "enable chained CNI installation")
	cmd.Flags().StringVar(&args.Cni_net_dir, "cni-net-dir", args.Cni_net_dir, "set the CNI install directory")
	cmd.Flags().StringVar(&args.Cni_bin_dir, "cni-bin-dir", args.Cni_bin_dir, "set the CNI binary directory")
	cmd.Flags().StringVar(&args.Cni_conf_name, "cni-conf-name", args.Cni_conf_name, "set the CNI configuration name")
	cmd.Flags().StringToStringVar(&args.Cni_nodeSelector, "cni-node-selector", args.Cni_nodeSelector, "node selector for CNI deployment")
	cmd.Flags().StringVar(&args.ControlPlane_mode, "mode", args.ControlPlane_mode, kuma_cmd.UsageOptions("kuma cp modes", "standalone", "zone", "global"))
	cmd.Flags().StringVar(&args.ControlPlane_zone, "zone", args.ControlPlane_zone, "set the Kuma zone name")
	cmd.Flags().BoolVar(&args.UseNodePort, "use-node-port", false, "use NodePort instead of LoadBalancer")
	cmd.Flags().BoolVar(&args.Ingress_enabled, "ingress-enabled", args.Ingress_enabled, "install Kuma with an Ingress deployment, using the Data Plane image")
	cmd.Flags().StringVar(&args.Ingress_drainTime, "ingress-drain-time", args.Ingress_drainTime, "drain time for Envoy proxy")
	cmd.Flags().BoolVar(&args.IngressUseNodePort, "ingress-use-node-port", false, "use NodePort instead of LoadBalancer for the Ingress Service")
	cmd.Flags().StringToStringVar(&args.Ingress_nodeSelector, "ingress-node-selector", args.Ingress_nodeSelector, "node selector for Zone Ingress")
	cmd.Flags().BoolVar(&args.Egress_enabled, "egress-enabled", args.Egress_enabled, "install Kuma with an Egress deployment, using the Data Plane image")
	cmd.Flags().StringVar(&args.Egress_drainTime, "egress-drain-time", args.Egress_drainTime, "drain time for Envoy proxy")
	cmd.Flags().StringVar(&args.Egress_service_type, "egress-service-type", "ClusterIP", "the type for the Egress Service (ie. ClusterIP, NodePort, LoadBalancer)")
	cmd.Flags().StringToStringVar(&args.Egress_nodeSelector, "egress-node-selector", args.Egress_nodeSelector, "node selector for Zone Egress")
	cmd.Flags().StringToStringVar(&args.Hooks_nodeSelector, "hooks-node-selector", args.Hooks_nodeSelector, "node selector for Helm hooks")
	cmd.Flags().BoolVar(&args.WithoutKubernetesConnection, "without-kubernetes-connection", false, "install without connection to Kubernetes cluster. This can be used for initial Kuma installation, but not for upgrades")
	cmd.Flags().BoolVar(&args.ExperimentalGatewayAPI, "experimental-gatewayapi", false, "install experimental Gateway API support")
	cmd.Flags().StringSliceVarP(&args.ValueFiles, "values", "f", []string{}, "specify values in a YAML file or '-' for stdin. This is similar to `helm template <chart> -f ...`")
	cmd.Flags().StringArrayVar(&args.Values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2), This is similar to `helm template <chart> --set ...` to use set-file or set-string just use helm instead")
	cmd.Flags().StringArrayVar(&args.SkipKinds, "skip-kinds", []string{}, "INTERNAL: A list of kubernetes kinds to not generate (useful, to ignore CRDs for example)")
	if err := cmd.Flags().MarkHidden("skip-kinds"); err != nil {
		panic(err.Error())
	}

	// This is used for testing the install command without a cluster
	cmd.Flags().StringArrayVar(&args.APIVersions, "api-versions", []string{}, "INTERNAL: Kubernetes api versions used for Capabilities.APIVersions")
	if err := cmd.Flags().MarkHidden("api-versions"); err != nil {
		panic(err.Error())
	}

	cmd.Flags().BoolVar(&args.DumpValues, "dump-values", false, "output all possible values for the configuration. This is similar to `helm show values <chart>")
	return cmd
}

func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
