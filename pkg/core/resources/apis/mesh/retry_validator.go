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
) (err validators.ValidationError) {
	if duration.Seconds == 0 && duration.Nanos == 0 {
		err.AddViolationAt(path, HasToBeGreaterThan0Violation)
	}

	return
}
func validateDuration_GreaterThan0OrNil(
	path validators.PathBuilder,
	duration *durationpb.Duration,
) (err validators.ValidationError) {
	if duration == nil {
		return
	}

	if duration.Seconds == 0 && duration.Nanos == 0 {
		err.AddViolationAt(path, WhenDefinedHasToBeGreaterThan0Violation)
	}

	return
}

func validateConfProtocolBackOff(
	path validators.PathBuilder,
	conf *mesh_proto.Retry_Conf_BackOff,
) (err validators.ValidationError) {
	if conf == nil {
		return
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

	return
}

func validateUint32_GreaterThan0OrNil(
	path validators.PathBuilder,
	value *wrapperspb.UInt32Value,
) (err validators.ValidationError) {
	if value == nil {
		return
	}

	if value.Value == 0 {
		err.AddViolationAt(path, WhenDefinedHasToBeGreaterThan0Violation)
	}

	return
}

func validateConfHttp(
	path validators.PathBuilder,
	conf *mesh_proto.Retry_Conf_Http,
) (err validators.ValidationError) {
	if conf == nil {
		return
	}

	numRetries, perTryTimeout, backOff, retriableStatusCodes, retriableMethods :=
		conf.NumRetries, conf.PerTryTimeout, conf.BackOff,
		conf.RetriableStatusCodes, conf.RetriableMethods

	if numRetries == nil && perTryTimeout == nil && backOff == nil &&
		retriableStatusCodes == nil && retriableMethods == nil {
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

	return
}

func validateConfGrpc(
	path validators.PathBuilder,
	conf *mesh_proto.Retry_Conf_Grpc,
) (err validators.ValidationError) {
	if conf == nil {
		return
	}

	numRetries, perTryTimeout, backOff, retryOn :=
		conf.NumRetries, conf.PerTryTimeout, conf.BackOff, conf.RetryOn

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

	return
}

func validateConfTcp(
	path validators.PathBuilder,
	conf *mesh_proto.Retry_Conf_Tcp,
) (err validators.ValidationError) {
	if conf == nil {
		return
	}

	if conf.MaxConnectAttempts == 0 {
		err.AddViolationAt(
			path.Field("maxConnectAttempts"),
			HasToBeGreaterThan0Violation,
		)
	}

	return
}

func (r *RetryResource) validateConf() (err validators.ValidationError) {
	path := validators.RootedAt("conf")
	conf := r.Spec.GetConf()

	if conf == nil {
		err.AddViolationAt(path, HasToBeDefinedViolation)
		return
	}

	if conf.Http == nil && conf.Grpc == nil && conf.Tcp == nil {
		err.AddViolationAt(
			path,
			"missing protocol [grpc|http|tcp] configuration",
		)
		return
	}

	err.Add(validateConfHttp(path.Field("http"), conf.GetHttp()))
	err.Add(validateConfGrpc(path.Field("grpc"), conf.GetGrpc()))
	err.Add(validateConfTcp(path.Field("tcp"), conf.GetTcp()))

	return
}
