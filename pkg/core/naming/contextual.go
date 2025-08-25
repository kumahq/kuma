package naming

import (
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
)

type sectionName interface {
	~string | constraints.Unsigned
}

// MustContextualInboundName is a helper for code paths where you are certain the resource's
// type exists in the global registry. Use it only when the resource and its type are guaranteed
// to be valid.
func MustContextualInboundName[T sectionName](r core_model.Resource, sectionName T) string {
	name, err := ContextualInboundName(r, sectionName)
	if err != nil {
		panic(err)
	}
	return name
}

func ContextualTransparentProxyName(direction string, ipVersion int) string {
	return fmt.Sprintf("self_transparentproxy_passthrough_%s_ipv%d", direction, ipVersion)
}

func ContextualInboundName[T sectionName](r core_model.Resource, sectionName T) (string, error) {
	if r == nil {
		return "", errors.New("cannot build contextual inbound name: resource is nil")
	}

	desc, err := core_registry.Global().DescriptorFor(r.Descriptor().Name)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"cannot build contextual inbound name for type %q and section %v: type not found in global registry",
			r.Descriptor().Name,
			sectionName,
		)
	}

	return fmt.Sprintf("self_inbound_%s_%v", desc.ShortName, sectionName), nil
}
