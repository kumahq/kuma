//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AbortConf) DeepCopyInto(out *AbortConf) {
	*out = *in
	out.Percentage = in.Percentage
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AbortConf.
func (in *AbortConf) DeepCopy() *AbortConf {
	if in == nil {
		return nil
	}
	out := new(AbortConf)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Conf) DeepCopyInto(out *Conf) {
	*out = *in
	if in.Http != nil {
		in, out := &in.Http, &out.Http
		*out = new([]FaultInjectionConf)
		if **in != nil {
			in, out := *in, *out
			*out = make([]FaultInjectionConf, len(*in))
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
func (in *DelayConf) DeepCopyInto(out *DelayConf) {
	*out = *in
	out.Value = in.Value
	out.Percentage = in.Percentage
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DelayConf.
func (in *DelayConf) DeepCopy() *DelayConf {
	if in == nil {
		return nil
	}
	out := new(DelayConf)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FaultInjectionConf) DeepCopyInto(out *FaultInjectionConf) {
	*out = *in
	if in.Abort != nil {
		in, out := &in.Abort, &out.Abort
		*out = new(AbortConf)
		**out = **in
	}
	if in.Delay != nil {
		in, out := &in.Delay, &out.Delay
		*out = new(DelayConf)
		**out = **in
	}
	if in.ResponseBandwidth != nil {
		in, out := &in.ResponseBandwidth, &out.ResponseBandwidth
		*out = new(ResponseBandwidthConf)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FaultInjectionConf.
func (in *FaultInjectionConf) DeepCopy() *FaultInjectionConf {
	if in == nil {
		return nil
	}
	out := new(FaultInjectionConf)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *From) DeepCopyInto(out *From) {
	*out = *in
	in.TargetRef.DeepCopyInto(&out.TargetRef)
	in.Default.DeepCopyInto(&out.Default)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new From.
func (in *From) DeepCopy() *From {
	if in == nil {
		return nil
	}
	out := new(From)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshFaultInjection) DeepCopyInto(out *MeshFaultInjection) {
	*out = *in
	in.TargetRef.DeepCopyInto(&out.TargetRef)
	if in.From != nil {
		in, out := &in.From, &out.From
		*out = make([]From, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshFaultInjection.
func (in *MeshFaultInjection) DeepCopy() *MeshFaultInjection {
	if in == nil {
		return nil
	}
	out := new(MeshFaultInjection)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResponseBandwidthConf) DeepCopyInto(out *ResponseBandwidthConf) {
	*out = *in
	out.Percentage = in.Percentage
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResponseBandwidthConf.
func (in *ResponseBandwidthConf) DeepCopy() *ResponseBandwidthConf {
	if in == nil {
		return nil
	}
	out := new(ResponseBandwidthConf)
	in.DeepCopyInto(out)
	return out
}
