//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Conf) DeepCopyInto(out *Conf) {
	*out = *in
	if in.Interval != nil {
		in, out := &in.Interval, &out.Interval
		*out = new(v1.Duration)
		**out = **in
	}
	if in.Timeout != nil {
		in, out := &in.Timeout, &out.Timeout
		*out = new(v1.Duration)
		**out = **in
	}
	if in.UnhealthyThreshold != nil {
		in, out := &in.UnhealthyThreshold, &out.UnhealthyThreshold
		*out = new(int32)
		**out = **in
	}
	if in.HealthyThreshold != nil {
		in, out := &in.HealthyThreshold, &out.HealthyThreshold
		*out = new(int32)
		**out = **in
	}
	if in.InitialJitter != nil {
		in, out := &in.InitialJitter, &out.InitialJitter
		*out = new(v1.Duration)
		**out = **in
	}
	if in.IntervalJitter != nil {
		in, out := &in.IntervalJitter, &out.IntervalJitter
		*out = new(v1.Duration)
		**out = **in
	}
	if in.IntervalJitterPercent != nil {
		in, out := &in.IntervalJitterPercent, &out.IntervalJitterPercent
		*out = new(int32)
		**out = **in
	}
	if in.HealthyPanicThreshold != nil {
		in, out := &in.HealthyPanicThreshold, &out.HealthyPanicThreshold
		*out = new(int32)
		**out = **in
	}
	if in.FailTrafficOnPanic != nil {
		in, out := &in.FailTrafficOnPanic, &out.FailTrafficOnPanic
		*out = new(bool)
		**out = **in
	}
	if in.EventLogPath != nil {
		in, out := &in.EventLogPath, &out.EventLogPath
		*out = new(string)
		**out = **in
	}
	if in.AlwaysLogHealthCheckFailures != nil {
		in, out := &in.AlwaysLogHealthCheckFailures, &out.AlwaysLogHealthCheckFailures
		*out = new(bool)
		**out = **in
	}
	if in.NoTrafficInterval != nil {
		in, out := &in.NoTrafficInterval, &out.NoTrafficInterval
		*out = new(v1.Duration)
		**out = **in
	}
	if in.Tcp != nil {
		in, out := &in.Tcp, &out.Tcp
		*out = new(TcpHealthCheck)
		(*in).DeepCopyInto(*out)
	}
	if in.Http != nil {
		in, out := &in.Http, &out.Http
		*out = new(HttpHealthCheck)
		(*in).DeepCopyInto(*out)
	}
	if in.Grpc != nil {
		in, out := &in.Grpc, &out.Grpc
		*out = new(GrpcHealthCheck)
		(*in).DeepCopyInto(*out)
	}
	if in.ReuseConnection != nil {
		in, out := &in.ReuseConnection, &out.ReuseConnection
		*out = new(bool)
		**out = **in
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
func (in *GrpcHealthCheck) DeepCopyInto(out *GrpcHealthCheck) {
	*out = *in
	if in.Disabled != nil {
		in, out := &in.Disabled, &out.Disabled
		*out = new(bool)
		**out = **in
	}
	if in.ServiceName != nil {
		in, out := &in.ServiceName, &out.ServiceName
		*out = new(string)
		**out = **in
	}
	if in.Authority != nil {
		in, out := &in.Authority, &out.Authority
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GrpcHealthCheck.
func (in *GrpcHealthCheck) DeepCopy() *GrpcHealthCheck {
	if in == nil {
		return nil
	}
	out := new(GrpcHealthCheck)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HeaderValue) DeepCopyInto(out *HeaderValue) {
	*out = *in
	if in.Append != nil {
		in, out := &in.Append, &out.Append
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HeaderValue.
func (in *HeaderValue) DeepCopy() *HeaderValue {
	if in == nil {
		return nil
	}
	out := new(HeaderValue)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HttpHealthCheck) DeepCopyInto(out *HttpHealthCheck) {
	*out = *in
	if in.Disabled != nil {
		in, out := &in.Disabled, &out.Disabled
		*out = new(bool)
		**out = **in
	}
	if in.Path != nil {
		in, out := &in.Path, &out.Path
		*out = new(string)
		**out = **in
	}
	if in.RequestHeadersToAdd != nil {
		in, out := &in.RequestHeadersToAdd, &out.RequestHeadersToAdd
		*out = new([]HeaderValue)
		if **in != nil {
			in, out := *in, *out
			*out = make([]HeaderValue, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
	if in.ExpectedStatuses != nil {
		in, out := &in.ExpectedStatuses, &out.ExpectedStatuses
		*out = new([]int32)
		if **in != nil {
			in, out := *in, *out
			*out = make([]int32, len(*in))
			copy(*out, *in)
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HttpHealthCheck.
func (in *HttpHealthCheck) DeepCopy() *HttpHealthCheck {
	if in == nil {
		return nil
	}
	out := new(HttpHealthCheck)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshHealthCheck) DeepCopyInto(out *MeshHealthCheck) {
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

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshHealthCheck.
func (in *MeshHealthCheck) DeepCopy() *MeshHealthCheck {
	if in == nil {
		return nil
	}
	out := new(MeshHealthCheck)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TcpHealthCheck) DeepCopyInto(out *TcpHealthCheck) {
	*out = *in
	if in.Disabled != nil {
		in, out := &in.Disabled, &out.Disabled
		*out = new(bool)
		**out = **in
	}
	if in.Send != nil {
		in, out := &in.Send, &out.Send
		*out = new(string)
		**out = **in
	}
	if in.Receive != nil {
		in, out := &in.Receive, &out.Receive
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TcpHealthCheck.
func (in *TcpHealthCheck) DeepCopy() *TcpHealthCheck {
	if in == nil {
		return nil
	}
	out := new(TcpHealthCheck)
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
