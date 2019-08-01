/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8s

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
)

// NewKubeProxyHandler creates a proxy server which is similar to `kubectl proxy`
// but, unlike it, is not exposed outside of the process.
func NewKubeProxyHandler(cfg *rest.Config) (http.Handler, error) {
	host := cfg.Host
	if !strings.HasSuffix(host, "/") {
		host = host + "/"
	}
	target, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	transport, err := rest.TransportFor(cfg)
	if err != nil {
		return nil, err
	}
	upgradeTransport, err := makeUpgradeTransport(cfg)
	if err != nil {
		return nil, err
	}
	responder := &responder{}
	proxy := proxy.NewUpgradeAwareHandler(target, transport, false, false, responder)
	proxy.UpgradeTransport = upgradeTransport
	proxy.UseRequestLocation = true

	return proxy, nil
}

// makeUpgradeTransport creates a transport that explicitly bypasses HTTP2 support
// for proxy connections that must upgrade.
func makeUpgradeTransport(config *rest.Config) (proxy.UpgradeRequestRoundTripper, error) {
	transportConfig, err := config.TransportConfig()
	if err != nil {
		return nil, err
	}
	tlsConfig, err := transport.TLSConfigFor(transportConfig)
	if err != nil {
		return nil, err
	}
	rt := utilnet.SetOldTransportDefaults(&http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext: (&net.Dialer{
			Timeout: 30 * time.Second,
		}).DialContext,
	})

	upgrader, err := transport.HTTPWrappersForConfig(transportConfig, proxy.MirrorRequest)
	if err != nil {
		return nil, err
	}
	return proxy.NewUpgradeRequestRoundTripper(rt, upgrader), nil
}

type responder struct{}

func (r *responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
