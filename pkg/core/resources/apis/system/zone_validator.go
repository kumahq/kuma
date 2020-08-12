package system

import (
	"net"
	"net/url"

	"github.com/kumahq/kuma/pkg/core/validators"
)

func (c *ZoneResource) Validate() error {
	var err validators.ValidationError
	err.Add(c.validateIngress())
	return err.OrNil()
}

func (c *ZoneResource) validateIngress() validators.ValidationError {
	var verr validators.ValidationError
	if c.Spec.GetIngress().GetAddress() == "" {
		verr.AddViolation("address", "cannot be empty")
	} else {
		host, port, err := net.SplitHostPort(c.Spec.GetIngress().GetAddress())
		if err != nil {
			url, urlErr := url.Parse(c.Spec.GetIngress().GetAddress())
			if urlErr == nil && url.Scheme != "" {
				verr.AddViolation("address", "should not be URL. Expected format is hostname:port")
			} else {
				verr.AddViolation("address", "invalid address: "+err.Error())
			}
		} else {
			if host == "" {
				verr.AddViolation("address", "host has to be explicitly specified")
			}
			if port == "" {
				verr.AddViolation("address", "port has to be explicitly specified")
			}
		}
	}

	return verr
}
