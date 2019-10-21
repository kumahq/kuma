package http

import (
	"context"
	nethttp "net/http"
	"net/url"
	"path"
	"time"
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
			req.URL.Path = path.Join(baseURL.Path, req.URL.Path)
		}
		return delegate.Do(req)
	})
}

func ClientWithTimeout(delegate Client, timeout time.Duration) Client {
	return ClientFunc(func(req *nethttp.Request) (*nethttp.Response, error) {
		ctx, cancel := context.WithTimeout(req.Context(), timeout)
		defer cancel()
		return delegate.Do(req.WithContext(ctx))
	})
}
