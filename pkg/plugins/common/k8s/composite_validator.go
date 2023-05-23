package k8s

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type AdmissionValidator interface {
	webhook.AdmissionHandler
	InjectDecoder(d *admission.Decoder)
	Supports(admission.Request) bool
}

type CompositeValidator struct {
	Validators []AdmissionValidator
}

func (c *CompositeValidator) AddValidator(validator AdmissionValidator) {
	c.Validators = append(c.Validators, validator)
}

func (c *CompositeValidator) IntoWebhook(scheme *runtime.Scheme) *admission.Webhook {
	decoder := admission.NewDecoder(scheme)
	for _, validator := range c.Validators {
		validator.InjectDecoder(decoder)
	}

	return &admission.Webhook{
		Handler: admission.HandlerFunc(func(ctx context.Context, req admission.Request) admission.Response {
			for _, validator := range c.Validators {
				if validator.Supports(req) {
					resp := validator.Handle(ctx, req)
					if !resp.Allowed {
						return resp
					}
				}
			}
			return admission.Allowed("")
		}),
	}
}
