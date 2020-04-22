package bootstrap

import (
	// force plugins to get initialized and registered

	_ "github.com/Kong/kuma/pkg/plugins/bootstrap/k8s"
	_ "github.com/Kong/kuma/pkg/plugins/bootstrap/universal"

	_ "github.com/Kong/kuma/pkg/plugins/resources/k8s"
	_ "github.com/Kong/kuma/pkg/plugins/resources/memory"
	_ "github.com/Kong/kuma/pkg/plugins/resources/postgres"

	_ "github.com/Kong/kuma/pkg/plugins/secrets/k8s"
	_ "github.com/Kong/kuma/pkg/plugins/secrets/universal"

	_ "github.com/Kong/kuma/pkg/plugins/discovery/k8s"

	_ "github.com/Kong/kuma/pkg/plugins/runtime/k8s"

	_ "github.com/Kong/kuma/pkg/plugins/ca/builtin"
	_ "github.com/Kong/kuma/pkg/plugins/ca/provided"
)
