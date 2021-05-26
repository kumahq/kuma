package http

import (
	nethttp "net/http"
	"net/url"
	"path"

	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
)

type Client interface {
	Do(req *nethttp.Request) (*nethttp.Response, error)
}

type ClientFunc func(req *nethttp.Request) (*nethttp.Response, error)

func (f ClientFunc) Do(req *nethttp.Request) (*nethttp.Response, error) {
	return f(req)
}

func ClientWithBaseURL(delegate Client, baseURL *url.URL, apiServer *config_proto.ControlPlaneCoordinates_ApiServer) Client {
	return ClientFunc(func(req *nethttp.Request) (*nethttp.Response, error) {
		if req.URL != nil {
			req.URL.Scheme = baseURL.Scheme
			req.URL.Host = baseURL.Host
			req.URL.Path = path.Join(baseURL.Path, req.URL.Path)
			for _, header := range apiServer.AddHeaders {
				req.Header.Add(header.Key, header.Value)
			}
		}
		return delegate.Do(req)
	})
}
