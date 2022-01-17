package mesh

import (
	"fmt"
	"net"
	"strconv"

	"github.com/asaskevich/govalidator"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (es *ExternalServiceResource) Validate() error {
	var err validators.ValidationError
	err.Add(validateExternalServiceNetworking(es.Spec.GetNetworking()))

	validateProtocol := func(path validators.PathBuilder, selector map[string]string) validators.ValidationError {
		var result validators.ValidationError
		if value, exist := selector[mesh_proto.ProtocolTag]; exist {
			if ParseProtocol(value) == ProtocolUnknown {
				err.AddViolationAt(path.Key(mesh_proto.ProtocolTag), fmt.Sprintf("tag %q has an invalid value %q. %s", mesh_proto.ProtocolTag, value, AllowedValuesHint(SupportedProtocols.Strings()...)))
			}
		}
		return result
	}
	err.Add(ValidateTags(validators.RootedAt("tags"), es.Spec.Tags, ValidateTagsOpts{
		RequireService:      true,
		ExtraTagsValidators: []TagsValidatorFunc{validateProtocol},
	}))

	return err.OrNil()
}

func validateExternalServiceNetworking(networking *mesh_proto.ExternalService_Networking) validators.ValidationError {
	var err validators.ValidationError
	path := validators.RootedAt("networking")
	if networking == nil {
		err.AddViolation("networking", "should have networking")
	} else {
		err.Add(validateExternalServiceAddress(path, networking.Address))
	}

	if networking.GetTls().GetServerName() != nil && networking.GetTls().GetServerName().GetValue() == "" {
		err.AddViolationAt(path.Field("tls").Field("serverName"), "cannot be empty")
	}
	return err
}

func validateExternalServiceAddress(path validators.PathBuilder, address string) validators.ValidationError {
	var err validators.ValidationError
	if address == "" {
		err.AddViolationAt(path.Field("address"), "address can't be empty")
		return err
	}

	host, port, e := net.SplitHostPort(address)
	if e != nil {
		err.AddViolationAt(path.Field("address"), "unable to parse address")
	}
	if !govalidator.IsIP(host) && !govalidator.IsDNSName(host) {
		err.AddViolationAt(path.Field("address"), "address has to be a valid IP address or a domain name")
	}

	iport, e := strconv.ParseUint(port, 10, 32)
	if e != nil {
		err.AddViolationAt(path.Field("address"), "unable to parse port in address")
	}

	err.Add(ValidatePort(path.Field("address"), uint32(iport)))

	return err
}
