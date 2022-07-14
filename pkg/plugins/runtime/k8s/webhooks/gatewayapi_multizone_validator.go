package webhooks

import (
	"context"
	"errors"
	"net/http"

	admission "k8s.io/api/admission/v1"
	kube_webhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	gatewayapi_alpha "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

func NewGatewayAPIMultizoneValidator(cpMode config_core.CpMode) *kube_admission.Webhook {
	return &kube_admission.Webhook{
		Handler: &GatewayAPIMultizoneValidator{
			CpMode: cpMode,
		},
	}
}

var GatewayAPINotSupportedErr = errors.New("GatewayAPI of Kuma is only supported in Standalone deployments")

type GatewayAPIMultizoneValidator struct {
	CpMode  config_core.CpMode
	Decoder *kube_admission.Decoder
}

var _ kube_webhook.AdmissionHandler = &GatewayAPIMultizoneValidator{}

func (g *GatewayAPIMultizoneValidator) InjectDecoder(d *kube_admission.Decoder) error {
	g.Decoder = d
	return nil
}

func (g *GatewayAPIMultizoneValidator) Handle(_ context.Context, req kube_admission.Request) kube_admission.Response {
	if req.Operation == admission.Create {
		var controllerName string

		gatewayClass := &gatewayapi.GatewayClass{}
		if err := g.Decoder.Decode(req, gatewayClass); err != nil {
			gatewayClassAlpha := &gatewayapi_alpha.GatewayClass{}
			if err := g.Decoder.Decode(req, gatewayClassAlpha); err != nil {
				return kube_admission.Errored(http.StatusBadRequest, err)
			}

			controllerName = string(gatewayClassAlpha.Spec.ControllerName)
		} else {
			controllerName = string(gatewayClass.Spec.ControllerName)
		}

		if g.CpMode != config_core.Standalone && gatewayapi.GatewayController(controllerName) == common.ControllerName {
			return kube_admission.Errored(http.StatusBadRequest, GatewayAPINotSupportedErr)
		}
	}
	return kube_admission.Allowed("")
}
