package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	token_server "github.com/Kong/kuma/pkg/config/token-server"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/tokens/builtin"
	"github.com/Kong/kuma/pkg/tokens/builtin/issuer"
	"github.com/Kong/kuma/pkg/tokens/builtin/server/types"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"io/ioutil"
	"net/http"

	config_core "github.com/Kong/kuma/pkg/config/core"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
)

func SetupServer(rt core_runtime.Runtime) error {
	switch env := rt.Config().Environment; env {
	case config_core.KubernetesEnvironment:
		return nil
	case config_core.UniversalEnvironment:
		generator, err := builtin.NewDataplaneTokenIssuer(rt)
		if err != nil {
			return err
		}
		srv := &DataplaneTokenServer{
			Config: rt.Config().DataplaneTokenServer,
			Issuer: generator,
		}
		if err := core_runtime.Add(rt, srv); err != nil {
			return err
		}
	default:
		return errors.Errorf("unknown environment type %s", env)
	}
	return nil
}

var log = core.Log.WithName("dataplane-token-server")

type DataplaneTokenServer struct {
	Config *token_server.DataplaneTokenServerConfig
	Issuer issuer.DataplaneTokenIssuer
}

var _ core_runtime.Component = &DataplaneTokenServer{}

func (a *DataplaneTokenServer) Start(stop <-chan struct{}) error {
	httpServer, httpErrChan := a.startHttpServer()

	var httpsServer *http.Server
	var httpsErrChan chan error
	if a.Config.TlsEnabled() {
		httpsServer, httpsErrChan = a.startHttpsServer()
	} else {
		httpsErrChan = make(chan error)
	}

	select {
	case <-stop:
		log.Info("stopping")
		var multiErr error
		if err := httpServer.Shutdown(context.Background()); err != nil {
			multiErr = multierr.Combine(err)
		}
		if httpsServer != nil {
			if err := httpsServer.Shutdown(context.Background()); err != nil {
				multiErr = multierr.Combine(err)
			}
		}
		return multiErr
	case err := <-httpErrChan:
		return err
	case err := <-httpsErrChan:
		return err
	}
}

func (a *DataplaneTokenServer) startHttpServer() (*http.Server, chan error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tokens", a.handleIdentityRequest)

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", a.Config.Local.Port),
		Handler: mux,
	}

	errChan := make(chan error)

	go func() {
		defer close(errChan)
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Error(err, "http server terminated with an error")
				errChan <- err
				return
			}
		}
		log.Info("http server terminated normally")
	}()
	log.Info("starting server", "port", a.Config.Local.Port)
	return server, errChan
}

func (a *DataplaneTokenServer) startHttpsServer() (*http.Server, chan error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tokens", a.handleIdentityRequest)

	errChan := make(chan error)

	tlsConfig, err := requireClientCerts(a.Config.Public.ClientCertFiles)
	if err != nil {
		errChan <- err
	}

	server := &http.Server{
		Addr:      fmt.Sprintf("%s:%d", a.Config.Public.Interface, a.Config.Public.Port),
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	go func() {
		defer close(errChan)
		if err := server.ListenAndServeTLS(a.Config.Public.TlsCertFile, a.Config.Public.TlsKeyFile); err != nil {
			if err != http.ErrServerClosed {
				log.Error(err, "https server terminated with an error")
				errChan <- err
				return
			}
		}
		log.Info("https server terminated normally")
	}()
	log.Info("starting server", "interface", a.Config.Public.Interface, "port", a.Config.Public.Port, "tls", true)
	return server, errChan
}

func requireClientCerts(certFiles []string) (*tls.Config, error) {
	clientCertPool := x509.NewCertPool()
	for _, cert := range certFiles {
		caCert, err := ioutil.ReadFile(cert)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read certificate %s", cert)
		}
		clientCertPool.AppendCertsFromPEM(caCert)
	}
	tlsConfig := &tls.Config{
		ClientCAs:  clientCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

func (a *DataplaneTokenServer) handleIdentityRequest(resp http.ResponseWriter, req *http.Request) {
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err, "Could not read a request")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	idReq := types.DataplaneTokenRequest{}
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
