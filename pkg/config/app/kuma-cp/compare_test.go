package kuma_cp_test

import (
	"fmt"
	"testing"

	"github.com/kr/pretty"

	"github.com/kumahq/kuma/v2/pkg/config"
	kuma_cp "github.com/kumahq/kuma/v2/pkg/config/app/kuma-cp"
)

func TestCompare(t *testing.T) {
	// Load from YAML
	cfgFromYaml := kuma_cp.Config{}
	err := config.Load("kuma-cp.defaults.yaml", &cfgFromYaml)
	if err != nil {
		t.Fatal(err)
	}

	// Get default
	cfgDefault := kuma_cp.DefaultConfig()

	// Compare
	diffs := pretty.Diff(cfgFromYaml, cfgDefault)
	if len(diffs) > 0 {
		fmt.Println("Differences found:")
		for i, diff := range diffs {
			if i >= 50 {
				fmt.Printf("... and %d more differences\n", len(diffs)-50)
				break
			}
			fmt.Println(diff)
		}
		t.Fail()
	} else {
		fmt.Println("No differences!")
	}
}
