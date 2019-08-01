package http

import (
	nethttp "net/http"
	"net/url"
)

type Client interface {
	Do(req *nethttp.Request) (*nethttp.Response, error)
}

type ClientFunc func(req *nethttp.Request) (*nethttp.Response, error)

func (f ClientFunc) Do(req *nethttp.Request) (*nethttp.Response, error) {
	return f(req)
}

func ClientWithBaseURL(delegate Client, baseURL *url.URL) Client {
	return ClientFunc(func(req *nethttp.Request) (*nethttp.Response, error) {
		if req.URL != nil {
			req.URL.Scheme = baseURL.Scheme
			req.URL.Host = baseURL.Host
		}
		return delegate.Do(req)
	})
}
