package mesh

import (
	"fmt"
	"net"
	"strconv"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (es *ExternalServiceResource) Validate() error {
	var err validators.ValidationError
	err.Add(validateExternalServiceNetworking(es.Spec.GetNetworking()))

	err.Add(validateTags(es.Spec.Tags))
	if _, exist := es.Spec.Tags[mesh_proto.ServiceTag]; !exist {
		err.AddViolationAt(validators.RootedAt("tags").Key(mesh_proto.ServiceTag), `tag has to exist`)
	}
	if value, exist := es.Spec.Tags[mesh_proto.ProtocolTag]; exist {
		if ParseProtocol(value) == ProtocolUnknown {
			err.AddViolationAt(validators.RootedAt("tags").Key(mesh_proto.ProtocolTag), fmt.Sprintf("tag %q has an invalid value %q. %s", mesh_proto.ProtocolTag, value, AllowedValuesHint(SupportedProtocols.Strings()...)))
		}
	}

	return err.OrNil()
}

func validateExternalServiceNetworking(networking *mesh_proto.ExternalService_Networking) validators.ValidationError {
	var err validators.ValidationError
	path := validators.RootedAt("networking")
	err.Add(validateExtarnalServiceAddress(path, networking.Address))
	return err
}

func validateExtarnalServiceAddress(path validators.PathBuilder, address string) validators.ValidationError {
	var err validators.ValidationError
	if address == "" {
		err.AddViolationAt(path.Field("address"), "address can't be empty")
		return err
	}

	host, port, e := net.SplitHostPort(address)
	if e != nil {
		err.AddViolationAt(path.Field("address"), "unable to parse address")
	}
	if !DNSRegex.MatchString(host) {
		err.AddViolationAt(path.Field("address"), "address has to be valid IP address or domain name")
	}

	iport, e := strconv.Atoi(port)
	if e != nil {
		err.AddViolationAt(path.Field("address"), "unable to parse port in address")
	}
	if iport < 1 || iport > 65535 {
		err.AddViolationAt(path.Field("address"), "port has to be in range of [1, 65535]")
	}
	return err
}
