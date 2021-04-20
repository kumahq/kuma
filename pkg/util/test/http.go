package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

const healthCheckPath = "/---/ready"

type CheckedHttpServer interface {
	Server() *httptest.Server
	Ready() error
}

type healthCheckHandler struct {
	http.Handler
}

func (h *healthCheckHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.RequestURI == fmt.Sprintf("/%s", healthCheckPath) {
		writer.WriteHeader(200)
		return
	}
	h.Handler.ServeHTTP(writer, request)
}

type healthCheckServer struct {
	server *httptest.Server
}

func (s *healthCheckServer) Server() *httptest.Server {
	return s.server
}

func (s *healthCheckServer) Ready() error {
	res, err := http.DefaultClient.Get(fmt.Sprintf("%s/%s", s.server.URL, healthCheckPath))
	if err == nil {
		_ = res.Body.Close()
	}

	return err
}

func NewHttpServer(handler http.Handler) CheckedHttpServer {
	return &healthCheckServer{
		server: httptest.NewServer(&healthCheckHandler{
			handler,
		}),
	}
}
