//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	commonv1alpha1 "github.com/kumahq/kuma/api/common/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackOff) DeepCopyInto(out *BackOff) {
	*out = *in
	if in.BaseInterval != nil {
		in, out := &in.BaseInterval, &out.BaseInterval
		*out = new(v1.Duration)
		**out = **in
	}
	if in.MaxInterval != nil {
		in, out := &in.MaxInterval, &out.MaxInterval
		*out = new(v1.Duration)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackOff.
func (in *BackOff) DeepCopy() *BackOff {
	if in == nil {
		return nil
	}
	out := new(BackOff)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Conf) DeepCopyInto(out *Conf) {
	*out = *in
	if in.TCP != nil {
		in, out := &in.TCP, &out.TCP
		*out = new(TCP)
		(*in).DeepCopyInto(*out)
	}
	if in.HTTP != nil {
		in, out := &in.HTTP, &out.HTTP
		*out = new(HTTP)
		(*in).DeepCopyInto(*out)
	}
	if in.GRPC != nil {
		in, out := &in.GRPC, &out.GRPC
		*out = new(GRPC)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Conf.
func (in *Conf) DeepCopy() *Conf {
	if in == nil {
		return nil
	}
	out := new(Conf)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GRPC) DeepCopyInto(out *GRPC) {
	*out = *in
	if in.NumRetries != nil {
		in, out := &in.NumRetries, &out.NumRetries
		*out = new(uint32)
		**out = **in
	}
	if in.PerTryTimeout != nil {
		in, out := &in.PerTryTimeout, &out.PerTryTimeout
		*out = new(v1.Duration)
		**out = **in
	}
	if in.BackOff != nil {
		in, out := &in.BackOff, &out.BackOff
		*out = new(BackOff)
		(*in).DeepCopyInto(*out)
	}
	if in.RateLimitedBackOff != nil {
		in, out := &in.RateLimitedBackOff, &out.RateLimitedBackOff
		*out = new(RateLimitedBackOff)
		(*in).DeepCopyInto(*out)
	}
	if in.RetryOn != nil {
		in, out := &in.RetryOn, &out.RetryOn
		*out = new([]GRPCRetryOn)
		if **in != nil {
			in, out := *in, *out
			*out = make([]GRPCRetryOn, len(*in))
			copy(*out, *in)
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GRPC.
func (in *GRPC) DeepCopy() *GRPC {
	if in == nil {
		return nil
	}
	out := new(GRPC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTP) DeepCopyInto(out *HTTP) {
	*out = *in
	if in.NumRetries != nil {
		in, out := &in.NumRetries, &out.NumRetries
		*out = new(uint32)
		**out = **in
	}
	if in.PerTryTimeout != nil {
		in, out := &in.PerTryTimeout, &out.PerTryTimeout
		*out = new(v1.Duration)
		**out = **in
	}
	if in.BackOff != nil {
		in, out := &in.BackOff, &out.BackOff
		*out = new(BackOff)
		(*in).DeepCopyInto(*out)
	}
	if in.RateLimitedBackOff != nil {
		in, out := &in.RateLimitedBackOff, &out.RateLimitedBackOff
		*out = new(RateLimitedBackOff)
		(*in).DeepCopyInto(*out)
	}
	if in.RetryOn != nil {
		in, out := &in.RetryOn, &out.RetryOn
		*out = new([]HTTPRetryOn)
		if **in != nil {
			in, out := *in, *out
			*out = make([]HTTPRetryOn, len(*in))
			copy(*out, *in)
		}
	}
	if in.RetriableResponseHeaders != nil {
		in, out := &in.RetriableResponseHeaders, &out.RetriableResponseHeaders
		*out = new([]commonv1alpha1.HeaderMatcher)
		if **in != nil {
			in, out := *in, *out
			*out = make([]commonv1alpha1.HeaderMatcher, len(*in))
			copy(*out, *in)
		}
	}
	if in.RetriableRequestHeaders != nil {
		in, out := &in.RetriableRequestHeaders, &out.RetriableRequestHeaders
		*out = new([]commonv1alpha1.HeaderMatcher)
		if **in != nil {
			in, out := *in, *out
			*out = make([]commonv1alpha1.HeaderMatcher, len(*in))
			copy(*out, *in)
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTP.
func (in *HTTP) DeepCopy() *HTTP {
	if in == nil {
		return nil
	}
	out := new(HTTP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshRetry) DeepCopyInto(out *MeshRetry) {
	*out = *in
	in.TargetRef.DeepCopyInto(&out.TargetRef)
	if in.To != nil {
		in, out := &in.To, &out.To
		*out = make([]To, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshRetry.
func (in *MeshRetry) DeepCopy() *MeshRetry {
	if in == nil {
		return nil
	}
	out := new(MeshRetry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitedBackOff) DeepCopyInto(out *RateLimitedBackOff) {
	*out = *in
	if in.ResetHeaders != nil {
		in, out := &in.ResetHeaders, &out.ResetHeaders
		*out = make([]ResetHeader, len(*in))
		copy(*out, *in)
	}
	if in.MaxInterval != nil {
		in, out := &in.MaxInterval, &out.MaxInterval
		*out = new(v1.Duration)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitedBackOff.
func (in *RateLimitedBackOff) DeepCopy() *RateLimitedBackOff {
	if in == nil {
		return nil
	}
	out := new(RateLimitedBackOff)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResetHeader) DeepCopyInto(out *ResetHeader) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResetHeader.
func (in *ResetHeader) DeepCopy() *ResetHeader {
	if in == nil {
		return nil
	}
	out := new(ResetHeader)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TCP) DeepCopyInto(out *TCP) {
	*out = *in
	if in.MaxConnectAttempt != nil {
		in, out := &in.MaxConnectAttempt, &out.MaxConnectAttempt
		*out = new(uint32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TCP.
func (in *TCP) DeepCopy() *TCP {
	if in == nil {
		return nil
	}
	out := new(TCP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *To) DeepCopyInto(out *To) {
	*out = *in
	in.TargetRef.DeepCopyInto(&out.TargetRef)
	in.Default.DeepCopyInto(&out.Default)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new To.
func (in *To) DeepCopy() *To {
	if in == nil {
		return nil
	}
	out := new(To)
	in.DeepCopyInto(out)
	return out
}
