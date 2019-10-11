package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/tokens/builtin"
	"github.com/Kong/kuma/pkg/tokens/builtin/issuer"
	"io/ioutil"
	"net/http"

	config_core "github.com/Kong/kuma/pkg/config/core"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
)

func SetupServer(rt core_runtime.Runtime) error {
	if rt.Config().Environment != config_core.KubernetesEnvironment {
		generator, err := builtin.NewDataplaneTokenIssuer(rt)
		if err != nil {
			return err
		}
		srv := &DataplaneTokenServer{
			Port:   rt.Config().DataplaneTokenServer.Port,
			Issuer: generator,
		}
		if err := core_runtime.Add(rt, srv); err != nil {
			return err
		}
	}
	return nil
}

type DataplaneTokenRequest struct {
	Name string `json:"name"`
	Mesh string `json:"mesh"`
}

func (i DataplaneTokenRequest) ToProxyId() xds.ProxyId {
	return xds.ProxyId{
		Mesh:      i.Mesh,
		Namespace: "default",
		Name:      i.Name,
	}
}

var log = core.Log.WithName("dataplane-token-server")

type DataplaneTokenServer struct {
	Port   int
	Issuer issuer.DataplaneTokenIssuer
}

var _ core_runtime.Component = &DataplaneTokenServer{}

func (a *DataplaneTokenServer) Start(stop <-chan struct{}) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", a.handleIdentityRequest)

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", a.Port),
		Handler: mux,
	}

	errChan := make(chan error)

	go func() {
		defer close(errChan)
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Error(err, "terminated with an error")
				errChan <- err
				return
			}
		}
		log.Info("terminated normally")
	}()
	log.Info("starting", "port", a.Port)

	select {
	case <-stop:
		log.Info("stopping")
		return server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func (a *DataplaneTokenServer) handleIdentityRequest(resp http.ResponseWriter, req *http.Request) {
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err, "Could not read a request")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	idReq := DataplaneTokenRequest{}
	if err := json.Unmarshal(bytes, &idReq); err != nil {
		log.Error(err, "Could not parse a request")
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	if idReq.Name == "" {
		resp.WriteHeader(http.StatusBadRequest)
		if _, err := resp.Write([]byte("name cannot be empty")); err != nil {
			log.Error(err, "Could write a response")
		}
		return
	}
	if idReq.Mesh == "" {
		resp.WriteHeader(http.StatusBadRequest)
		if _, err := resp.Write([]byte("mesh cannot be empty")); err != nil {
			log.Error(err, "Could write a response")
		}
		return
	}

	token, err := a.Issuer.Generate(idReq.ToProxyId())
	if err != nil {
		log.Error(err, "Could not sign a token")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp.Header().Set("content-type", "text/plain")
	if _, err := resp.Write([]byte(token)); err != nil {
		log.Error(err, "Could write a response")
	}
}
