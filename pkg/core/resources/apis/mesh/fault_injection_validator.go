package mesh

import (
	"net/http"
	"regexp"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (f *FaultInjectionResource) Validate() error {
	var err validators.ValidationError
	err.Add(f.validateSources())
	err.Add(f.validateDestinations())
	err.Add(f.validateConf())
	return err.OrNil()
}

func (f *FaultInjectionResource) HasFaultDelay() bool {
	faultDelay := f.Spec.Conf.GetDelay()
	return faultDelay != nil && !proto.Equal(faultDelay, &v1alpha1.FaultInjection_Conf_Delay{})
}

func (f *FaultInjectionResource) HasFaultAbort() bool {
	faultAbort := f.Spec.Conf.GetAbort()
	return faultAbort != nil && !proto.Equal(faultAbort, &v1alpha1.FaultInjection_Conf_Abort{})
}

func (f *FaultInjectionResource) HasFaultResponseBandwidth() bool {
	faultResponseBandwidth := f.Spec.Conf.GetResponseBandwidth()
	return faultResponseBandwidth != nil && !proto.Equal(faultResponseBandwidth, &v1alpha1.FaultInjection_Conf_ResponseBandwidth{})
}

func (f *FaultInjectionResource) validateSources() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("sources"), f.Spec.GetSources(), ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
		ValidateTagsOpts: ValidateTagsOpts{
			RequireAtLeastOneTag: true,
		},
	})
}

func (f *FaultInjectionResource) validateDestinations() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("destinations"), f.Spec.GetDestinations(), ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
		ValidateTagsOpts: ValidateTagsOpts{
			RequireAtLeastOneTag: true,
			ExtraTagsValidators: []TagsValidatorFunc{
				ProtocolValidator(ProtocolHTTP, ProtocolHTTP2, ProtocolGRPC),
			},
		},
	})
}

func (f *FaultInjectionResource) validateConf() (err validators.ValidationError) {
	root := validators.RootedAt("conf")
	if !f.HasFaultDelay() && !f.HasFaultAbort() && !f.HasFaultResponseBandwidth() {
		err.AddViolationAt(root, "must have at least one of the faults configured")
	}
	if f.HasFaultDelay() {
		err.Add(validateDelay(root.Field("delay"), f.Spec.GetConf().GetDelay()))
	}
	if f.HasFaultAbort() {
		err.Add(validateAbort(root.Field("abort"), f.Spec.GetConf().GetAbort()))
	}
	if f.HasFaultResponseBandwidth() {
		err.Add(validateResponseBandwidth(root.Field("responseBandwidth"), f.Spec.GetConf().GetResponseBandwidth()))
	}
	return
}

func validateDelay(path validators.PathBuilder, delay *v1alpha1.FaultInjection_Conf_Delay) (err validators.ValidationError) {
	err.Add(validatePercentage(path, delay.GetPercentage()))
	if delay.GetValue() == nil {
		err.AddViolationAt(path.Field("value"), "cannot be empty")
	}
	return
}

func validateAbort(path validators.PathBuilder, abort *v1alpha1.FaultInjection_Conf_Abort) (err validators.ValidationError) {
	err.Add(validatePercentage(path, abort.GetPercentage()))
	err.Add(validateHttpStatus(path, abort.GetHttpStatus()))
	return
}

func validateResponseBandwidth(path validators.PathBuilder, bandwidth *v1alpha1.FaultInjection_Conf_ResponseBandwidth) (err validators.ValidationError) {
	err.Add(validatePercentage(path, bandwidth.GetPercentage()))
	err.Add(validateLimit(path, bandwidth.GetLimit()))
	return
}

func validatePercentage(path validators.PathBuilder, percentage *wrapperspb.DoubleValue) (err validators.ValidationError) {
	if percentage == nil {
		err.AddViolationAt(path.Field("percentage"), "cannot be empty")
		return
	}

	if percentage.GetValue() < 0.0 || percentage.GetValue() > 100.0 {
		err.AddViolationAt(path.Field("percentage"), "has to be in [0.0 - 100.0] range")
	}
	return
}

func validateLimit(path validators.PathBuilder, limit *wrapperspb.StringValue) (err validators.ValidationError) {
	if limit == nil {
		err.AddViolationAt(path.Field("limit"), "cannot be empty")
		return
	}

	if matched, _ := regexp.MatchString(`\d*\s?[gmk]bps`, limit.GetValue()); !matched {
		err.AddViolationAt(path.Field("limit"), "has to be in kbps/mbps/gbps units")
	}
	return
}

func validateHttpStatus(path validators.PathBuilder, httpStatus *wrapperspb.UInt32Value) (err validators.ValidationError) {
	if httpStatus == nil {
		err.AddViolationAt(path.Field("httpStatus"), "cannot be empty")
		return
	}

	if status := http.StatusText(int(httpStatus.GetValue())); len(status) == 0 {
		err.AddViolationAt(path.Field("httpStatus"), "http status code is incorrect")
	}
	return
}
