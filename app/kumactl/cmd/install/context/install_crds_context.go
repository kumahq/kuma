package context

import (
	"strings"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	controlplane "github.com/kumahq/kuma/app/kumactl/pkg/install/k8s/control-plane"
)

type InstallCrdsArgs struct {
	OnlyMissing bool
}

type InstallCrdsContext struct {
	Args                    InstallCrdsArgs
	InstallCrdTemplateFiles func(InstallCrdsArgs) (data.FileList, error)
	FilterCrdNamesToInstall func([]string) []string
}

func DefaultInstallCrdsContext() InstallCrdsContext {
	return InstallCrdsContext{
		Args: InstallCrdsArgs{
			OnlyMissing: false,
		},
		InstallCrdTemplateFiles: func(args InstallCrdsArgs) (data.FileList, error) {
			helmFiles, err := data.ReadFiles(controlplane.HelmTemplates)
			if err != nil {
				return nil, err
			}

			crdFiles := helmFiles.Filter(func(file data.File) bool {
				return strings.HasPrefix(file.FullPath, "/crds")
			})

			return crdFiles, nil
		},
		FilterCrdNamesToInstall: func(names []string) []string {
			var result []string

			for _, name := range names {
				if strings.HasSuffix(name, "kuma.io") {
					result = append(result, name)
				}
			}

			return result
		},
	}
}
