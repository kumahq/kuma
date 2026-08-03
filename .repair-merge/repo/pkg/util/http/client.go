package http

import (
	nethttp "net/http"
	"net/url"
	"path"
)

type Client interface {
	Do(req *nethttp.Request) (*nethttp.Response, error)
}

type ClientFunc func(req *nethttp.Request) (*nethttp.Response, error)

func (f ClientFunc) Do(req *nethttp.Request) (*nethttp.Response, error) {
	return f(req)
}

func ClientWithBaseURL(delegate Client, baseURL *url.URL, headers map[string]string) Client {
	return ClientFunc(func(req *nethttp.Request) (*nethttp.Response, error) {
		if req.URL != nil {
			req.URL.Scheme = baseURL.Scheme
			req.URL.Host = baseURL.Host
			req.URL.Path = path.Join(baseURL.Path, req.URL.Path)
			for k, v := range headers {
				req.Header.Add(k, v)
			}
		}
		return delegate.Do(req)
	})
}
