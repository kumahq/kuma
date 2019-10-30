package k8s

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	kube_util_net "k8s.io/apimachinery/pkg/util/net"
	kube_rest "k8s.io/client-go/rest"

	"net/http/httptest"
)

func NewServiceProxyTransport(kubeApiServer http.RoundTripper, namespace, name string) http.RoundTripper {
	return &resourceProxyTransport{
		resource:      "services",
		namespace:     namespace,
		name:          name,
		kubeApiServer: kubeApiServer,
	}
}

var _ http.RoundTripper = &resourceProxyTransport{}

type resourceProxyTransport struct {
	resource      string
	namespace     string
	name          string
	kubeApiServer http.RoundTripper
}

func (t *resourceProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = kube_util_net.CloneRequest(req)

	if strings.HasPrefix(req.URL.Path, "/") {
		if 1 < len(req.URL.Path) {
			req.URL.Path = req.URL.Path[1:]
		} else {
			req.URL.Path = ""
		}
	}

	req.URL.Path = fmt.Sprintf("/api/v1/namespaces/%s/%s/%s/proxy/%s", t.namespace, t.resource, t.name, req.URL.Path)

	return t.kubeApiServer.RoundTrip(req)
}

func NewKubeApiProxyTransport(cfg *kube_rest.Config) (http.RoundTripper, error) {
	proxy, err := NewKubeProxyHandler(cfg)
	if err != nil {
		return nil, err
	}
	return NewInprocessKubeProxyTransport(proxy), nil
}

func NewInprocessKubeProxyTransport(handler http.Handler) http.RoundTripper {
	return &inprocessKubeProxyTransport{handler: handler}
}

var _ http.RoundTripper = &inprocessKubeProxyTransport{}

type inprocessKubeProxyTransport struct {
	handler http.Handler
}

func (t *inprocessKubeProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var buf bytes.Buffer
	if err := req.Write(&buf); err != nil {
		return nil, err
	}
	req, err := http.ReadRequest(bufio.NewReader(&buf))
	if err != nil {
		return nil, err
	}

	type result struct {
		resp *http.Response
		err  error
	}
	ch := make(chan result)
	go func() {
		defer close(ch)
		defer func() {
			if p := recover(); p != nil {
				ch <- result{resp: nil, err: fmt.Errorf("kube proxy paniced: %v\n%s", p, debug.Stack())}
			}
		}()
		w := httptest.NewRecorder()
		t.handler.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()
		ch <- result{resp: resp, err: nil}
	}()
	res := <-ch
	return res.resp, res.err
}
