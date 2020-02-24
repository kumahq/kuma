/*
Package accesslog replicates access log format supported by Envoy.

In order to allow users of Kuma to reuse the same access log format strings
both in file logs and TCP logs, we need to have native support for
Envoy access log command syntax.

Use ParseFormat() function to parse a format string.

Use HttpLogEntryFormatter interface to format an HTTP log entry.

Use TcpLogEntryFormatter interface to format a TCP log entry.

Use HttpLogConfigurer interface to configure `envoy.http_grpc_access_log` filter.

Use TcpLogConfigurer interface to configure `envoy.tcp_grpc_access_log` filter.

The initial implementation is missing the following features:
1. `%START_TIME%` commands ignore the user-defined format string
2. `%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%` commands return a stub value
3. `%FILTER_STATE(KEY):Z%` commands return a stub value
*/
package accesslog
