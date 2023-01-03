//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2023 Kuma authors.

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

import ()

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
		*out = make([]Backend, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Sampling.DeepCopyInto(&out.Sampling)
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make([]Tag, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
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
func (in *Sampling) DeepCopyInto(out *Sampling) {
	*out = *in
	if in.Overall != nil {
		in, out := &in.Overall, &out.Overall
		*out = new(uint32)
		**out = **in
	}
	if in.Client != nil {
		in, out := &in.Client, &out.Client
		*out = new(uint32)
		**out = **in
	}
	if in.Random != nil {
		in, out := &in.Random, &out.Random
		*out = new(uint32)
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
	if in.Header != nil {
		in, out := &in.Header, &out.Header
		*out = new(HeaderTag)
		**out = **in
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
