package envoy

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"

	envoy_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	"github.com/gogo/protobuf/proto"

	konvoy_dp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-dp"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
)

const minimalConfigTmpl string = `
node:
  id: {{.Dataplane.Id}}
  cluster: {{.Dataplane.Service}}

admin:
  access_log_path: /dev/null
  address:
    socket_address:
      protocol: TCP
      address: 127.0.0.1
      port_value: {{ .Dataplane.AdminPort }}

dynamic_resources:
  lds_config: {ads: {}}
  cds_config: {ads: {}}
  ads_config:
    api_type: GRPC
    grpc_services:
    - envoy_grpc:
        cluster_name: ads_cluster

static_resources:
  clusters:
  - name: ads_cluster
    connect_timeout: 0.25s
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    http2_protocol_options: {}
    upstream_connection_options:
      # configure a TCP keep-alive to detect and reconnect to the admin
      # server in the event of a TCP socket half open connection
      tcp_keepalive: {}
    load_assignment:
      cluster_name: ads_cluster
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: {{ .ControlPlane.XdsServer.Address }}
                port_value: {{ .ControlPlane.XdsServer.Port }}
`

func MinimalBootstrapConfig(cfg konvoy_dp.Config) (proto.Message, error) {
	tmpl, err := template.New("bootstrap").Parse(minimalConfigTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse config template")
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return nil, errors.Wrap(err, "failed to render config template")
	}
	config := &envoy_bootstrap.Bootstrap{}
	if err := util_proto.FromYAML(buf.Bytes(), config); err != nil {
		return nil, errors.Wrap(err, "failed to parse bootstrap config")
	}
	return config, nil
}
