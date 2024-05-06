package webhooks

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

type PolicyNamespaceValidator struct {
	Decoder         *admission.Decoder
	SystemNamespace string
}

func (p *PolicyNamespaceValidator) InjectDecoder(decoder *admission.Decoder) {
	p.Decoder = decoder
}

func (p *PolicyNamespaceValidator) Handle(ctx context.Context, request admission.Request) admission.Response {
	if request.Namespace != p.SystemNamespace {
		return admission.Denied(fmt.Sprintf("policy can only be created in the system namespace:%s", p.SystemNamespace))
	}
	return admission.Allowed("")
}

func (p *PolicyNamespaceValidator) Supports(request admission.Request) bool {
	desc, err := registry.Global().DescriptorFor(core_model.ResourceType(request.Kind.Kind))
	if err != nil {
		return false
	}
	return desc.IsPluginOriginated && desc.IsPolicy
}
