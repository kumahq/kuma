package main

import (
	"fmt"

	"sigs.k8s.io/yaml"

	kuma_cp "github.com/kumahq/kuma/v2/pkg/config/app/kuma-cp"
)

func main() {
	cfg := kuma_cp.DefaultConfig()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
