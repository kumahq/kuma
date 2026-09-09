package perf

import (
	"os"
	"testing"

	"github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_apis "github.com/kumahq/kuma/v3/pkg/core/resources/apis"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies"
)

func TestMain(m *testing.M) {
	plugins.InitAll(core_apis.NameToModule)
	plugins.InitAll(policies.NameToModule)
	os.Exit(m.Run())
}
