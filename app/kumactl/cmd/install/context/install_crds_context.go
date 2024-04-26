package context

import (
	"strings"

	"github.com/kumahq/kuma/deployments"
	"github.com/kumahq/kuma/pkg/util/data"
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
			helmFiles, err := data.ReadFiles(deployments.KumaChartFS())
			if err != nil {
				return nil, err
			}

			return helmFiles.Filter(func(file data.File) bool {
				return strings.Contains(file.FullPath, "crds/")
			}), nil
		},
	}
}
