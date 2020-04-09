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
	CertBytes          string
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
    dataplane.admin.port: "{{ .AdminPort }}"
{{ end }}

{{if .AdminPort }}
admin:
  access_log_path: {{ .AdminAccessLogPath }}
  address:
    socket_address:
      protocol: TCP
      address: {{ .AdminAddress }}
      port_value: {{ .AdminPort }}
{{ end }}

stats_config:
  stats_tags:
  - tag_name: name
    regex: '^grpc\.((.+)\.)'
  - tag_name: status
    regex: '^grpc.*streams_closed(_([0-9]+))'
  - tag_name: worker
    regex: '(worker_([0-9]+)\.)'
  - tag_name: listener
    regex: '((.+?)\.)rbac\.'

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
{{ if .CertBytes }}
    transport_socket:
      name: envoy.transport_sockets.tls
      typed_config:
        "@type": type.googleapis.com/envoy.api.v2.auth.UpstreamTlsContext
        sni: {{ .XdsHost }}
        common_tls_context:
          tls_params:
            tls_minimum_protocol_version: TLSv1_2
          validation_context:
            trusted_ca:
              inline_bytes: "{{ .CertBytes }}"
{{ end }}
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
