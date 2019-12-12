package webhooks

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type AdmissionValidator interface {
	admission.DecoderInjector
	webhook.AdmissionHandler
	Supports(admission.Request) bool
}

type CompositeValidator struct {
	Validators []AdmissionValidator
}

func (c *CompositeValidator) AddValidator(validator AdmissionValidator) {
	c.Validators = append(c.Validators, validator)
}

func (c *CompositeValidator) InjectDecoder(d *admission.Decoder) error {
	for _, validator := range c.Validators {
		if err := validator.InjectDecoder(d); err != nil {
			return err
		}
	}
	return nil
}

func (c *CompositeValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	for _, validator := range c.Validators {
		if validator.Supports(req) {
			resp := validator.Handle(ctx, req)
			if !resp.Allowed {
				return resp
			}
		}
	}
	return admission.Allowed("")
}

func (c *CompositeValidator) WebHook() *admission.Webhook {
	return &admission.Webhook{
		Handler: c,
	}
}
