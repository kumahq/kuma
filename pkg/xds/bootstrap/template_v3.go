package bootstrap

const configTemplateV3 string = `
node:
  id: {{.Id}}
  cluster: {{.Service}}
  metadata:
{{if .DataplaneToken }}
    dataplane.token: "{{.DataplaneToken}}"
{{end}}
{{if .DataplaneResource}}
    dataplane.resource: '{{.DataplaneResource}}'
{{end}}
{{if .AdminPort }}
    dataplane.admin.port: "{{ .AdminPort }}"
{{ end }}
{{if .DNSPort }}
    dataplane.dns.port: "{{ .DNSPort }}"
{{ end }}
{{if .EmptyDNSPort }}
    dataplane.dns.empty.port: "{{ .EmptyDNSPort }}"
{{ end }}
{{if .ProxyType }}
    dataplane.proxyType: "{{ .ProxyType }}"
{{ end }}
    version:
      kumaDp:
        version: "{{ .KumaDpVersion }}"
        gitTag: "{{ .KumaDpGitTag }}"
        gitCommit: "{{ .KumaDpGitCommit }}"
        buildDate: "{{ .KumaDpBuildDate }}"
      envoy:
        version: "{{ .EnvoyVersion }}"
        build: "{{ .EnvoyBuild }}"
{{if .DynamicMetadata }}
    dynamicMetadata:
{{ range $key, $value := .DynamicMetadata }}
      {{ $key }}: "{{ $value }}"
{{ end }}
{{ end }}

{{if .AdminPort }}
admin:
  access_log_path: {{ .AdminAccessLogPath }}
  address:
    socket_address:
      protocol: TCP
      address: "{{ .AdminAddress }}"
      port_value: {{ .AdminPort }}
{{ end }}

layered_runtime:
  layers:
  - name: kuma
    static_layer:
      envoy.restart_features.use_apple_api_for_dns_lookups: false
      re2.max_program_size.error_level: 4294967295 # UINT32_MAX
      re2.max_program_size.warn_level: 1000

stats_config:
  stats_tags:
  - tag_name: name
    regex: '^grpc\.((.+)\.)'
  - tag_name: status
    regex: '^grpc.*streams_closed(_([0-9]+))'
  - tag_name: kafka_name
    regex: '^kafka(\.(\S*[0-9]))\.'
  - tag_name: kafka_type
    regex: '^kafka\..*\.(.*)'
  - tag_name: worker
    regex: '(worker_([0-9]+)\.)'
  - tag_name: listener
    regex: '((.+?)\.)rbac\.'

{{ if .HdsEnabled }}
hds_config:
  api_type: GRPC
  transport_api_version: V3
  set_node_on_first_message_only: true
  grpc_services:
    - envoy_grpc:
        cluster_name: ads_cluster
{{ if .DataplaneToken }}
      initialMetadata:
      - key: "authorization"
        value: "{{ .DataplaneToken }}"
{{ end }}
{{ end }}

dynamic_resources:
  lds_config:
    ads: {}
    resourceApiVersion: V3 
  cds_config:
    ads: {}
    resourceApiVersion: V3
  ads_config:
    api_type: GRPC
    transport_api_version: V3
    timeout: {{ .XdsConnectTimeout }}
    grpc_services:
    - envoy_grpc:
        cluster_name: ads_cluster
{{ if .DataplaneToken }}
      initialMetadata:
      - key: "authorization"
        value: "{{ .DataplaneToken }}"
{{ end }}
static_resources:
  clusters:
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
  - name: ads_cluster
    connect_timeout: {{ .XdsConnectTimeout }}
    type: {{ .XdsClusterType }}
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
    transport_socket:
      name: envoy.transport_sockets.tls
      typed_config:
        '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
        sni: {{ .XdsHost }}
        common_tls_context:
          tls_params:
            tls_minimum_protocol_version: TLSv1_2
          validation_context:
            match_subject_alt_names:
            - exact: {{ .XdsHost }}
{{ if .CertBytes }}
            trusted_ca:
              inline_bytes: "{{ .CertBytes }}"
{{ end }}
`
