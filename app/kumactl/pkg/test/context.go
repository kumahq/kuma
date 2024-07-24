package test

import (
	"bytes"
	"time"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/api/openapi/types"
	kumactl_app "github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/resources"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	util_http "github.com/kumahq/kuma/pkg/util/http"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var defaultArgs = kumactl_cmd.RootArgs{
	ConfigType: kumactl_cmd.InMemory,
}

var DummyIndexResponse = types.IndexResponse{
	Hostname:   "localhost",
	Product:    kuma_version.Product,
	Version:    kuma_version.Build.Version,
	InstanceId: "test-instance",
	ClusterId:  "test-cluster",
}

var defaultNewAPIServerClient = func(client util_http.Client) resources.ApiServerClient {
	return resources.NewStaticApiServiceClient(DummyIndexResponse)
}

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

// DefaultTestingRootCmd returns the DefaultRootCmd with server API mocked to return
// current version. Useful for tests which don't actually require the server but need to
// avoid extraneous warnings.
func DefaultTestingRootCmd(args ...string) (*bytes.Buffer, *bytes.Buffer, *cobra.Command) {
	ctx := kumactl_cmd.DefaultRootContext()
	ctx.Runtime.NewAPIServerClient = defaultNewAPIServerClient

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd := kumactl_app.NewRootCmd(ctx)
	cmd.SetArgs(args)
	cmd.SetErr(stderr)
	cmd.SetOut(stdout)

	return stdout, stderr, cmd
}
