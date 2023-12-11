package webhooks

import (
	"context"
	"errors"
	"net/http"

	admission "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kube_webhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

func NewGatewayAPIMultizoneValidator(gapiSupported bool, scheme *runtime.Scheme) *kube_admission.Webhook {
	return &kube_admission.Webhook{
		Handler: &GatewayAPIMultizoneValidator{
			GAPISupported: gapiSupported,
			Decoder:       kube_admission.NewDecoder(scheme),
		},
	}
}

var GatewayAPINotSupportedErr = errors.New("GatewayAPI of Kuma is only supported in a single-zone deployments")

type GatewayAPIMultizoneValidator struct {
	GAPISupported bool
	Decoder       *kube_admission.Decoder
}

var _ kube_webhook.AdmissionHandler = &GatewayAPIMultizoneValidator{}

func (g *GatewayAPIMultizoneValidator) Handle(_ context.Context, req kube_admission.Request) kube_admission.Response {
	if req.Operation == admission.Create {
		var controllerName string

		gatewayClass := &gatewayapi.GatewayClass{}
		if err := g.Decoder.Decode(req, gatewayClass); err != nil {
			return kube_admission.Errored(http.StatusBadRequest, err)
		} else {
			controllerName = string(gatewayClass.Spec.ControllerName)
		}

		if !g.GAPISupported && gatewayapi.GatewayController(controllerName) == common.ControllerName {
			return kube_admission.Errored(http.StatusBadRequest, GatewayAPINotSupportedErr)
		}
	}
	return kube_admission.Allowed("")
}
