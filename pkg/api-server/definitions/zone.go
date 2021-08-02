package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
)

var ZoneWsDefinition = ResourceWsDefinition{
	Type: system.ZoneType,
	Path: "zones",
}
