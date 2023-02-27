package framework

import (
	"strings"
)

// IsDataplaneOnline returns online, found, error
func IsDataplaneOnline(cluster Cluster, mesh, name string) (bool, bool, error) {
	out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes", "--mesh", mesh)
	if err != nil {
		return false, false, err
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, name) {
			return strings.Contains(line, "Online"), true, nil
		}
	}
	return false, false, nil
}

func DataplaneReceivedConfig(cluster Cluster, mesh, name string) (bool, error) {
	out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes", "--mesh", mesh, "-o", "yaml", name)
	if err != nil {
		return false, err
	}
	return strings.Contains(out, `responsesAcknowledged`), nil
}
