// A copy of https://github.com/istio/istio/tree/master/tools
// Last updated from branch release-1.10, commit 79d8df74cc04b1f8e6ee6b407f16f4ae512fb13a
// We need this to fix the use-case where the host has both IPv4 and IPv6
// The toolos in Istio do not provide this option today.
// See `hasLocalIPv6` and its usage for details of the changes.

package tools
