package bootstrap

type configParameters struct {
	Id        string
	Service   string
	AdminPort uint32
	XdsHost   string
	XdsPort   uint32
}

const configTemplate string = `
node:
  id: {{.Id}}
  cluster: {{.Service}}

admin:
  access_log_path: /dev/null
  address:
    socket_address:
      protocol: TCP
      address: 127.0.0.1
      port_value: {{ .AdminPort }}

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
                address: {{ .XdsHost }}
                port_value: {{ .XdsPort }}
`
