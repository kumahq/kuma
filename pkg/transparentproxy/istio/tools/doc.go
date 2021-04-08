// A copy of https://github.com/istio/istio/tree/master/tools
// We need this to fix the use-case where the host has both IPv4 and IPv6
// The toolos in Istio do not provide this option today.
// See `hasLocalIPv6` and its usage for details of the changes.

package tools
