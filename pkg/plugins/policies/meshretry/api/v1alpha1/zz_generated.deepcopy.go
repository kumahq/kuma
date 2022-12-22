//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022 Kuma authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
