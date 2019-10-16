package bootstrap

import "time"

type configParameters struct {
	Id                 string
	Service            string
	AdminAddress       string
	AdminPort          uint32
	AdminAccessLogPath string
	XdsHost            string
	XdsPort            uint32
	XdsConnectTimeout  time.Duration
	AccessLogPipe      string
	DataplaneTokenPath string
}

const configTemplate string = `
node:
  id: {{.Id}}
  cluster: {{.Service}}
  metadata:
{{if .DataplaneTokenPath}}
    dataplaneTokenPath: {{.DataplaneTokenPath}}
{{end}}    

{{if .AdminPort }}
admin:
  access_log_path: {{ .AdminAccessLogPath }}
  address:
    socket_address:
      protocol: TCP
      address: {{ .AdminAddress }}
      port_value: {{ .AdminPort }}
{{ end }}

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
    connect_timeout: {{ .XdsConnectTimeout }}
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
                address: {{ .XdsHost }}
                port_value: {{ .XdsPort }}
  - name: access_log_sink
    connect_timeout: {{ .XdsConnectTimeout }}
    type: STATIC
    lb_policy: ROUND_ROBIN
    http2_protocol_options: {}
    upstream_connection_options:
      # configure a TCP keep-alive to detect and reconnect to the admin
      # server in the event of a TCP socket half open connection
      tcp_keepalive: {}
    load_assignment:
      cluster_name: access_log_sink
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              pipe:
                path: {{ .AccessLogPipe }}
`
