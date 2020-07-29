package system

import (
	"net"

	"github.com/kumahq/kuma/pkg/core/validators"
)

func (c *ZoneResource) Validate() error {
	var err validators.ValidationError
	err.Add(c.validateIngress())
	return err.OrNil()
}

func (c *ZoneResource) validateIngress() validators.ValidationError {
	var verr validators.ValidationError
	host, port, err := net.SplitHostPort(c.Spec.GetIngress().GetAddress())
	if err != nil {
		verr.AddViolation("address", "invalid address")
	} else {
		if host == "" {
			verr.AddViolation("address", "host has to be explicitly specified")
		}
		if port == "" {
			verr.AddViolation("address", "port has to be explicitly specified")
		}
	}

	return verr
}
