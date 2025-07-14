package names

import "fmt"

const SystemPrefix = "system_"

func SystemGetAdminResourceName() string {
	return fmt.Sprintf("%senvoy_admin", SystemPrefix)
}
