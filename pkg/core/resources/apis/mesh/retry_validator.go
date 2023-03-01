package mesh

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

const (
	EmptyFieldViolation                     = "field cannot be empty"
	HasToBeDefinedViolation                 = "has to be defined"
	HasToBeGreaterThan0Violation            = "has to be greater than 0"
	WhenDefinedHasToBeGreaterThan0Violation = "has to be greater than 0" +
		" when defined"
	StatusCodesNotDefinedViolation = "retriableStatusCodes cannot" +
		" be empty when this option is specified"
)

func (r *RetryResource) Validate() error {
	var err validators.ValidationError

	err.Add(r.validateSources())
	err.Add(r.validateDestinations())
	err.Add(r.validateConf())

	return err.OrNil()
}

func (r *RetryResource) validateSources() validators.ValidationError {
	return ValidateSelectors(
		validators.RootedAt("sources"),
		r.Spec.Sources,
		ValidateSelectorsOpts{
			ValidateTagsOpts: ValidateTagsOpts{
				RequireAtLeastOneTag: true,
				RequireService:       true,
			},
			RequireAtLeastOneSelector: true,
		},
	)
}

func (r *RetryResource) validateDestinations() validators.ValidationError {
	return ValidateSelectors(
		validators.RootedAt("destinations"),
		r.Spec.Destinations,
		OnlyServiceTagAllowed,
	)
}

func getRepeatedRetryOnValues(
	values []mesh_proto.Retry_Conf_Grpc_RetryOn,
) map[string][]int {
	if len(values) == 0 {
		return nil
	}

	indexMap := map[string][]int{}
	repeatedMap := map[string][]int{}

	for i, item := range values {
		key := item.String()

		indexMap[key] = append(indexMap[key], i)
	}

	for retryOn, indexes := range indexMap {
		if len(indexes) > 1 {
			repeatedMap[retryOn] = indexes
		}
	}

	return repeatedMap
}

func getRepeatedRetryOnViolations(
	values []mesh_proto.Retry_Conf_Grpc_RetryOn,
) []string {
	if len(values) == 0 {
		return nil
	}

	var violations []string

	for value, indexes := range getRepeatedRetryOnValues(values) {
		fields := strings.Fields(fmt.Sprint(indexes))
		indexes := strings.Trim(strings.Join(fields, ", "), "[]")

		violations = append(violations, fmt.Sprintf(
			"repeated value %q at indexes [%s]",
			value,
			indexes,
		))
	}

	// As with our current tests design, it has to be sorted
	sort.Strings(violations)

	return violations
}

func validateDuration_GreaterThan0(
	path validators.PathBuilder,
	duration *durationpb.Duration,
) validators.ValidationError {
	var err validators.ValidationError
	if duration.Seconds == 0 && duration.Nanos == 0 {
		err.AddViolationAt(path, HasToBeGreaterThan0Violation)
	}

	return err
}

func validateDuration_GreaterThan0OrNil(
	path validators.PathBuilder,
	duration *durationpb.Duration,
) validators.ValidationError {
	var err validators.ValidationError
	if duration == nil {
		return err
	}

	if duration.Seconds == 0 && duration.Nanos == 0 {
		err.AddViolationAt(path, WhenDefinedHasToBeGreaterThan0Violation)
	}

	return err
}

func validateConfProtocolBackOff(
	path validators.PathBuilder,
	conf *mesh_proto.Retry_Conf_BackOff,
) validators.ValidationError {
	var err validators.ValidationError
	if conf == nil {
		return err
	}

	if conf.BaseInterval == nil {
		err.AddViolationAt(
			path.Field("baseInterval"),
			HasToBeDefinedViolation,
		)
	} else {
		if conf.BaseInterval.Seconds == 0 && conf.BaseInterval.Nanos == 0 {
			err.Add(validateDuration_GreaterThan0(
				path.Field("baseInterval"),
				conf.BaseInterval,
			))
		} else {
			err.Add(validateDuration_GreaterThan0OrNil(
				path.Field("maxInterval"),
				conf.MaxInterval,
			))
		}
	}

	return err
}

func validateUint32_GreaterThan0OrNil(
	path validators.PathBuilder,
	value *wrapperspb.UInt32Value,
) validators.ValidationError {
	var err validators.ValidationError
	if value == nil {
		return err
	}

	if value.Value == 0 {
		err.AddViolationAt(path, WhenDefinedHasToBeGreaterThan0Violation)
	}

	return err
}

func validateConfHttp(
	path validators.PathBuilder,
	conf *mesh_proto.Retry_Conf_Http,
) validators.ValidationError {
	var err validators.ValidationError
	if conf == nil {
		return err
	}

	numRetries, perTryTimeout, backOff, retriableStatusCodes, retriableMethods, retryOn := conf.NumRetries, conf.PerTryTimeout, conf.BackOff,
		conf.RetriableStatusCodes, conf.RetriableMethods, conf.RetryOn

	if numRetries == nil && perTryTimeout == nil && backOff == nil &&
		retriableStatusCodes == nil && retriableMethods == nil && retryOn == nil {
		err.AddViolationAt(path, EmptyFieldViolation)
	}

	err.Add(validateUint32_GreaterThan0OrNil(
		path.Field("numRetries"),
		numRetries,
	))

	err.Add(validateDuration_GreaterThan0OrNil(
		path.Field("perTryTimeout"),
		perTryTimeout,
	))

	err.Add(validateConfProtocolBackOff(path.Field("backOff"), backOff))

	for i, m := range retriableMethods {
		if m == mesh_proto.HttpMethod_NONE {
			err.AddViolationAt(path.Field("retriableMethods").Index(i), EmptyFieldViolation)
		}
	}

	for i, r := range retryOn {
		if r.String() == "retriable_status_codes" && retriableStatusCodes == nil {
			err.AddViolationAt(path.Field("retryOn").Index(i), StatusCodesNotDefinedViolation)
		}
	}

	return err
}

func validateConfGrpc(
	path validators.PathBuilder,
	conf *mesh_proto.Retry_Conf_Grpc,
) validators.ValidationError {
	var err validators.ValidationError
	if conf == nil {
		return err
	}

	numRetries, perTryTimeout, backOff, retryOn := conf.NumRetries, conf.PerTryTimeout, conf.BackOff, conf.RetryOn

	if numRetries == nil && perTryTimeout == nil && backOff == nil &&
		retryOn == nil {
		err.AddViolationAt(path, EmptyFieldViolation)
	}

	for _, violation := range getRepeatedRetryOnViolations(retryOn) {
		err.AddViolationAt(path.Field("retryOn"), violation)
	}

	err.Add(validateUint32_GreaterThan0OrNil(
		path.Field("numRetries"),
		numRetries,
	))

	err.Add(validateDuration_GreaterThan0OrNil(
		path.Field("perTryTimeout"),
		perTryTimeout,
	))

	err.Add(validateConfProtocolBackOff(path.Field("backOff"), backOff))

	return err
}

func validateConfTcp(
	path validators.PathBuilder,
	conf *mesh_proto.Retry_Conf_Tcp,
) validators.ValidationError {
	var err validators.ValidationError
	if conf == nil {
		return err
	}

	if conf.MaxConnectAttempts == 0 {
		err.AddViolationAt(
			path.Field("maxConnectAttempts"),
			HasToBeGreaterThan0Violation,
		)
	}

	return err
}

func (r *RetryResource) validateConf() validators.ValidationError {
	var err validators.ValidationError
	path := validators.RootedAt("conf")
	conf := r.Spec.GetConf()

	if conf == nil {
		err.AddViolationAt(path, HasToBeDefinedViolation)
		return err
	}

	if conf.Http == nil && conf.Grpc == nil && conf.Tcp == nil {
		err.AddViolationAt(
			path,
			"missing protocol [grpc|http|tcp] configuration",
		)
		return err
	}

	err.Add(validateConfHttp(path.Field("http"), conf.GetHttp()))
	err.Add(validateConfGrpc(path.Field("grpc"), conf.GetGrpc()))
	err.Add(validateConfTcp(path.Field("tcp"), conf.GetTcp()))

	return err
}
