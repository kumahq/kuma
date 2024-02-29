package bootstrap

import (
	// force plugins to get initialized and registered
	_ "github.com/kumahq/kuma/pkg/plugins/authn/api-server/certs"
	_ "github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens"
	_ "github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s"
	_ "github.com/kumahq/kuma/pkg/plugins/bootstrap/universal"
	_ "github.com/kumahq/kuma/pkg/plugins/ca/builtin"
	_ "github.com/kumahq/kuma/pkg/plugins/ca/provided"
	_ "github.com/kumahq/kuma/pkg/plugins/config/k8s"
	_ "github.com/kumahq/kuma/pkg/plugins/config/universal"
	_ "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	_ "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	_ "github.com/kumahq/kuma/pkg/plugins/resources/postgres"
	_ "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	_ "github.com/kumahq/kuma/pkg/plugins/runtime/k8s"
	_ "github.com/kumahq/kuma/pkg/plugins/runtime/opentelemetry"
	_ "github.com/kumahq/kuma/pkg/plugins/runtime/universal"
	_ "github.com/kumahq/kuma/pkg/plugins/secrets/k8s"
	_ "github.com/kumahq/kuma/pkg/plugins/secrets/universal"
)
