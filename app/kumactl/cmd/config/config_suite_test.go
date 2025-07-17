package config_test

import (
	"fmt"
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestConfigCmd(t *testing.T) {
	test.RunSpecs(t, "Config Cmd Suite")
}

func requiredFlagNotSet(name string) string {
	return fmt.Sprintf(`required flag\(s\) .*%q.* not set`, name)
}
