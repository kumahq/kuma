//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/util/intstr"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Backend) DeepCopyInto(out *Backend) {
	*out = *in
	if in.Zipkin != nil {
		in, out := &in.Zipkin, &out.Zipkin
		*out = new(ZipkinBackend)
		(*in).DeepCopyInto(*out)
	}
	if in.Datadog != nil {
		in, out := &in.Datadog, &out.Datadog
		*out = new(DatadogBackend)
		(*in).DeepCopyInto(*out)
	}
	if in.OpenTelemetry != nil {
		in, out := &in.OpenTelemetry, &out.OpenTelemetry
		*out = new(OpenTelemetryBackend)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Backend.
func (in *Backend) DeepCopy() *Backend {
	if in == nil {
		return nil
	}
	out := new(Backend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Conf) DeepCopyInto(out *Conf) {
	*out = *in
	if in.Backends != nil {
		in, out := &in.Backends, &out.Backends
		*out = new([]Backend)
		if **in != nil {
			in, out := *in, *out
			*out = make([]Backend, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
	if in.Sampling != nil {
		in, out := &in.Sampling, &out.Sampling
		*out = new(Sampling)
		(*in).DeepCopyInto(*out)
	}
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = new([]Tag)
		if **in != nil {
			in, out := *in, *out
			*out = make([]Tag, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
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
func (in *DatadogBackend) DeepCopyInto(out *DatadogBackend) {
	*out = *in
	if in.SplitService != nil {
		in, out := &in.SplitService, &out.SplitService
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatadogBackend.
func (in *DatadogBackend) DeepCopy() *DatadogBackend {
	if in == nil {
		return nil
	}
	out := new(DatadogBackend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HeaderTag) DeepCopyInto(out *HeaderTag) {
	*out = *in
	if in.Default != nil {
		in, out := &in.Default, &out.Default
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HeaderTag.
func (in *HeaderTag) DeepCopy() *HeaderTag {
	if in == nil {
		return nil
	}
	out := new(HeaderTag)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshTrace) DeepCopyInto(out *MeshTrace) {
	*out = *in
	in.TargetRef.DeepCopyInto(&out.TargetRef)
	in.Default.DeepCopyInto(&out.Default)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshTrace.
func (in *MeshTrace) DeepCopy() *MeshTrace {
	if in == nil {
		return nil
	}
	out := new(MeshTrace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenTelemetryBackend) DeepCopyInto(out *OpenTelemetryBackend) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenTelemetryBackend.
func (in *OpenTelemetryBackend) DeepCopy() *OpenTelemetryBackend {
	if in == nil {
		return nil
	}
	out := new(OpenTelemetryBackend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Sampling) DeepCopyInto(out *Sampling) {
	*out = *in
	if in.Overall != nil {
		in, out := &in.Overall, &out.Overall
		*out = new(intstr.IntOrString)
		**out = **in
	}
	if in.Client != nil {
		in, out := &in.Client, &out.Client
		*out = new(intstr.IntOrString)
		**out = **in
	}
	if in.Random != nil {
		in, out := &in.Random, &out.Random
		*out = new(intstr.IntOrString)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Sampling.
func (in *Sampling) DeepCopy() *Sampling {
	if in == nil {
		return nil
	}
	out := new(Sampling)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Tag) DeepCopyInto(out *Tag) {
	*out = *in
	if in.Literal != nil {
		in, out := &in.Literal, &out.Literal
		*out = new(string)
		**out = **in
	}
	if in.Header != nil {
		in, out := &in.Header, &out.Header
		*out = new(HeaderTag)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Tag.
func (in *Tag) DeepCopy() *Tag {
	if in == nil {
		return nil
	}
	out := new(Tag)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZipkinBackend) DeepCopyInto(out *ZipkinBackend) {
	*out = *in
	if in.TraceId128Bit != nil {
		in, out := &in.TraceId128Bit, &out.TraceId128Bit
		*out = new(bool)
		**out = **in
	}
	if in.ApiVersion != nil {
		in, out := &in.ApiVersion, &out.ApiVersion
		*out = new(string)
		**out = **in
	}
	if in.SharedSpanContext != nil {
		in, out := &in.SharedSpanContext, &out.SharedSpanContext
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZipkinBackend.
func (in *ZipkinBackend) DeepCopy() *ZipkinBackend {
	if in == nil {
		return nil
	}
	out := new(ZipkinBackend)
	in.DeepCopyInto(out)
	return out
}
