package context

import (
	"strings"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/deployments"
)

type InstallCrdsArgs struct {
	OnlyMissing             bool
	ExperimentalMeshGateway bool
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
			helmFiles, err := data.ReadFiles(deployments.KumaChartFS())
			if err != nil {
				return nil, err
			}

			crdFiles := helmFiles.Filter(func(file data.File) bool {
				return strings.Contains(file.FullPath, "crds/")
			})

			if !args.ExperimentalMeshGateway {
				crdFiles = crdFiles.Filter(ExcludeGatewayCRDs)
			}

			return crdFiles, nil
		},
	}
}
