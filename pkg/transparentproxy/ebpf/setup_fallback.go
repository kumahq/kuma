//go:build !linux

package ebpf

import (
	"fmt"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

func Setup(config.Config) (string, error) {
	return "", fmt.Errorf("ebpf is currently supported only on linux")
}
