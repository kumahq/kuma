package dp_server

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/runtime"
	sds_server "github.com/kumahq/kuma/pkg/sds/server"
	"github.com/kumahq/kuma/pkg/xds/bootstrap"
	xds_server "github.com/kumahq/kuma/pkg/xds/server"
)

func SetupServer(rt runtime.Runtime) error {
	dpServer := NewDpServer(*rt.Config().DpServer, rt.Metrics())
	if err := sds_server.RegisterSDS(rt, dpServer.grpcServer); err != nil {
		return errors.Wrap(err, "could not register SDS")
	}
	if err := xds_server.RegisterXDS(rt, dpServer.grpcServer); err != nil {
		return errors.Wrap(err, "could not register XDS")
	}
	if err := bootstrap.RegisterBootstrap(rt, dpServer.httpMux); err != nil {
		return err
	}
	if err := rt.Add(dpServer); err != nil {
		return err
	}
	return nil
}
