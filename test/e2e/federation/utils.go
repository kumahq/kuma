package federation

import (
	. "github.com/kumahq/kuma/test/framework"
)

func PrintLogs(clusters ...Cluster) {
	for _, cluster := range clusters {
		Logf("\n\n\n\n\nCP logs of: " + cluster.Name())
		logs, err := cluster.GetKumaCPLogs()
		if err != nil {
			Logf("could not retrieve cp logs")
		} else {
			Logf(logs)
		}
	}
}
