package config_test

import (
	"fmt"
)

func requiredFlagNotSet(name string) string {
	return fmt.Sprintf(`required flag\(s\) .*"%s".* not set`, name)
}
