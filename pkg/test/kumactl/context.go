package kumactl

import (
	"time"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	util_http "github.com/kumahq/kuma/pkg/util/http"
	util_test "github.com/kumahq/kuma/pkg/util/test"
)

var defaultArgs = kumactl_cmd.RootArgs{
	ConfigType: kumactl_cmd.InMemory,
}

var defaultNewAPIServerClient = util_test.GetMockNewAPIServerClient()

var defaultNewBaseAPIServerClient = func(server *config_proto.ControlPlaneCoordinates_ApiServer, _ time.Duration) (util_http.Client, error) {
	return nil, nil
}

var rootTime, _ = time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")

func MakeMinimalRootContext() *kumactl_cmd.RootContext {
	return &kumactl_cmd.RootContext{
		Args: defaultArgs,
		Runtime: kumactl_cmd.RootRuntime{
			Now: func() time.Time {
				return rootTime
			},
			Registry:               registry.NewTypeRegistry(),
			NewAPIServerClient:     defaultNewAPIServerClient,
			NewBaseAPIServerClient: defaultNewBaseAPIServerClient,
		},
	}
}

func MakeRootContext(now time.Time, resourceStore store.ResourceStore, res ...model.ResourceTypeDescriptor) (*kumactl_cmd.RootContext, error) {
	reg := registry.NewTypeRegistry()
	for _, r := range res {
		if err := reg.RegisterType(r); err != nil {
			return nil, err
		}
	}
	return &kumactl_cmd.RootContext{
		Args: defaultArgs,
		// should I add the in-memory here as well? or should I make other tests use this function?
		// why aren't other tests using this?
		Runtime: kumactl_cmd.RootRuntime{
			Registry:               reg,
			Now:                    func() time.Time { return now },
			NewBaseAPIServerClient: defaultNewBaseAPIServerClient,
			NewResourceStore: func(util_http.Client) store.ResourceStore {
				return resourceStore
			},
			NewAPIServerClient: defaultNewAPIServerClient,
		},
	}, nil
}
