package kumactl

import (
	"time"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	util_test "github.com/kumahq/kuma/pkg/util/test"
)

func MakeRootContext(now time.Time, resourceStore store.ResourceStore, res ...model.ResourceTypeDescriptor) (*kumactl_cmd.RootContext, error) {
	reg := registry.NewTypeRegistry()
	for _, r := range res {
		if err := reg.RegisterType(r); err != nil {
			return nil, err
		}
	}
	return &kumactl_cmd.RootContext{
		Runtime: kumactl_cmd.RootRuntime{
			Registry: reg,
			Now:      func() time.Time { return now },
			NewResourceStore: func(server *config_proto.ControlPlaneCoordinates_ApiServer) (store.ResourceStore, error) {
				return resourceStore, nil
			},
			NewAPIServerClient: util_test.GetMockNewAPIServerClient(),
		},
	}, nil
}
