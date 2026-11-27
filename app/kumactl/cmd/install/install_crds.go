package install

import (
	"context"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8s_apixv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	install_context "github.com/kumahq/kuma/v3/app/kumactl/cmd/install/context"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/install/k8s"
	"github.com/kumahq/kuma/v3/pkg/util/data"
)

func newInstallCrdsCmd(ctx *install_context.InstallCrdsContext) *cobra.Command {
	args := ctx.Args

	cmd := &cobra.Command{
		Use:   "crds",
		Short: "Install Kuma Custom Resource Definitions on Kubernetes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			wantCrdFiles, err := ctx.InstallCrdTemplateFiles(args)
			if err != nil {
				return errors.Wrap(err, "Failed to read CRD files")
			}

			if !args.OnlyMissing {
				return writeOutput(cmd.OutOrStdout(), args.OutputFile, data.JoinYAML(wantCrdFiles).Data)
			}

			crdsToInstallMap, err := mapCrdNamesToFiles(wantCrdFiles)
			if err != nil {
				return errors.Wrap(err, "Failed mapping CRD files with CRD names")
			}

			kubeClientConfig, err := k8s.DefaultClientConfig("", "")
			if err != nil {
				return errors.Wrap(err, "Could not detect Kubernetes configuration")
			}

			k8sClient, err := k8s_apixv1client.NewForConfig(kubeClientConfig)
			if err != nil {
				return errors.Wrap(err, "Failed obtaining Kubernetes client")
			}

			crds, err := k8sClient.CustomResourceDefinitions().List(context.Background(), v1.ListOptions{})
			if err != nil {
				return errors.Wrap(err, "Failed obtaining CRDs from Kubernetes cluster")
			}

			for _, installedCrdName := range getCrdNamesFromList(crds) {
				delete(crdsToInstallMap, installedCrdName)
			}

			var crdsToInstall []data.File
			for _, crd := range crdsToInstallMap {
				crdsToInstall = append(crdsToInstall, crd)
			}

			// don't create an empty output file, `kubectl apply` fails on it with "no objects passed to apply"
			if len(crdsToInstallMap) == 0 {
				return nil
			}

			return writeOutput(cmd.OutOrStdout(), args.OutputFile, data.JoinYAML(crdsToInstall).Data)
		},
	}

	cmd.Flags().BoolVar(&args.OnlyMissing, "only-missing", false, "install only resources which are not already present in a cluster")
	cmd.Flags().StringVar(&args.OutputFile, "output-file", "", "write resources to the given file instead of stdout")

	return cmd
}

// writeOutput writes to stdout, or to outputFile when it's set.
// The file is created only here, so it isn't left behind when there is nothing to write.
func writeOutput(stdout io.Writer, outputFile string, content []byte) (err error) {
	out := stdout
	if outputFile != "" {
		f, createErr := os.Create(outputFile)
		if createErr != nil {
			return errors.Wrap(createErr, "Failed to create output file")
		}
		defer func() {
			if cerr := f.Close(); cerr != nil && err == nil {
				err = errors.Wrap(cerr, "Failed to close output file")
			}
		}()
		out = f
	}

	if _, err = out.Write(content); err != nil {
		return errors.Wrap(err, "Failed to output rendered resources")
	}

	return nil
}

func getCrdNamesFromList(crds *apiextensionsv1.CustomResourceDefinitionList) []string {
	var names []string

	for _, crd := range crds.Items {
		names = append(names, crd.Name)
	}

	return names
}

func mapCrdNamesToFiles(files []data.File) (map[string]data.File, error) {
	result := map[string]data.File{}

	for _, file := range files {
		var crd apiextensionsv1.CustomResourceDefinition

		if err := yaml.Unmarshal(file.Data, &crd); err != nil {
			return nil, errors.Wrap(err, "Failed parsing file as CRD")
		}

		result[crd.Name] = file
	}

	return result, nil
}
