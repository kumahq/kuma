package versions

import (
	"os"

	"github.com/Masterminds/semver/v3"
	"sigs.k8s.io/yaml"
)

const previewVersion = "preview"

func Supported(path string) ([]*semver.Version, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	plainVersions := []struct {
		Version string `json:"version"`
	}{}
	if err := yaml.Unmarshal(content, &plainVersions); err != nil {
		return nil, err
	}

	var versions []*semver.Version
	for _, v := range plainVersions {
		if v.Version == previewVersion {
			continue
		}

		ver, err := semver.NewVersion(v.Version)
		if err != nil {
			return nil, err
		}
		versions = append(versions, ver)
	}

	return versions, nil
}

func OldestUpgradableToLatest(version []*semver.Version) string {
	return version[len(version)-1-2].String()
}
