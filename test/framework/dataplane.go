package framework

import (
	"strings"
)

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
