package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/sds/auth"
	"io/ioutil"
	"net/http"

	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
)

type IdentityRequest struct {
	Name string `json:"name"`
	Mesh string `json:"mesh"`
}

func (i IdentityRequest) ToProxyId() xds.ProxyId {
	return xds.ProxyId{
		Mesh:      i.Mesh,
		Namespace: "default",
		Name:      i.Name,
	}
}

var log = core.Log.WithName("initial-token-server")

type InitialTokenServer struct {
	LocalHttpPort       int
	CredentialGenerator auth.CredentialGenerator
}

var _ core_runtime.Component = &InitialTokenServer{}

func (a *InitialTokenServer) Start(stop <-chan struct{}) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", a.handleIdentityRequest)

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", a.LocalHttpPort),
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
	log.Info("starting", "port", a.LocalHttpPort)

	select {
	case <-stop:
		log.Info("stopping")
		return server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func (a *InitialTokenServer) handleIdentityRequest(resp http.ResponseWriter, req *http.Request) {
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err, "Could not read a request")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	idReq := IdentityRequest{}
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

	token, err := a.CredentialGenerator.Generate(idReq.ToProxyId())
	if err != nil {
		log.Error(err, "Could not sign a token")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := resp.Write([]byte(token)); err != nil {
		log.Error(err, "Could write a response")
	}
}
