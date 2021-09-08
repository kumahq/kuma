package config

import "github.com/kumahq/kuma/pkg/transparentproxy/istio"

// TransparentProxyConfig defines the configuration for all transparent
// proxy configurations, not just the Istio one.
//
// We alias the types in reverse here to avoid an import loop (the Istio
// submodule depending on Kuma for the config type).
type TransparentProxyConfig = istio.TransparentProxyConfig
