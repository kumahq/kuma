package cmd_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestKumactlCmdPkg(t *testing.T) {
	test.RunSpecs(t, "Kumactl Cmd Package Suite")
}
