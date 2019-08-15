package bootstrap

import (
	// force plugins to get initialized and registered
	_ "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/bootstrap/k8s"
	_ "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/discovery/k8s"
	_ "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s"

	_ "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/bootstrap/universal"
	_ "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/discovery/universal"
	_ "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	_ "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/postgres"
)
