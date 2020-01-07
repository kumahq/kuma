package server

import (
	"github.com/pkg/errors"
	"net/http"

	"github.com/Kong/kuma/app/kuma-injector/pkg/injector"
	kuma_injector_conf "github.com/Kong/kuma/pkg/config/app/kuma-injector"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"

	kube_manager "sigs.k8s.io/controller-runtime/pkg/manager"
	kube_webhook "sigs.k8s.io/controller-runtime/pkg/webhook"
)

func Setup(mgr kube_manager.Manager, cfg *kuma_injector_conf.Config) error {
	webhookServer := &kube_webhook.Server{
		Host:    cfg.WebHookServer.Address,
		Port:    int(cfg.WebHookServer.Port),
		CertDir: cfg.WebHookServer.CertDir,
	}
	if err := mesh_k8s.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrap(err, "could not add to scheme")
	}

	webhookServer.Register("/inject-sidecar", PodMutatingWebhook(injector.New(cfg.Injector, mgr.GetClient()).InjectKuma))
	webhookServer.WebhookMux.HandleFunc("/healthy", func(resp http.ResponseWriter, _ *http.Request) {
		resp.WriteHeader(http.StatusOK)
	})
	webhookServer.WebhookMux.HandleFunc("/ready", func(resp http.ResponseWriter, _ *http.Request) {
		resp.WriteHeader(http.StatusOK)
	})
	return mgr.Add(webhookServer)
}
