package system

import (
	"net"
	"net/url"

	"github.com/kumahq/kuma/pkg/core/validators"
)

func (c *ZoneResource) Validate() error {
	var err validators.ValidationError
	err.Add(c.validateRemoteControlPlane())
	err.Add(c.validateIngress())
	return err.OrNil()
}

func (c *ZoneResource) validateRemoteControlPlane() validators.ValidationError {
	var verr validators.ValidationError
	uri, err := url.ParseRequestURI(c.Spec.GetRemoteControlPlane().GetAddress())
	if err != nil {
		verr.AddViolation("address", "invalid URL")
	} else if uri.Port() == "" {
		verr.AddViolation("address", "port has to be explicitly specified")
	}

	return verr
}

func (c *ZoneResource) validateIngress() validators.ValidationError {
	var verr validators.ValidationError
	host, port, err := net.SplitHostPort(c.Spec.GetIngress().GetAddress())
	if err != nil {
		verr.AddViolation("address", "invalid address")
	} else {
		if host == "" {
			verr.AddViolation("address", "host has to be explicitly specified")
		} else if port == "" {
			verr.AddViolation("address", "port has to be explicitly specified")
		}
	}

	return verr
}
