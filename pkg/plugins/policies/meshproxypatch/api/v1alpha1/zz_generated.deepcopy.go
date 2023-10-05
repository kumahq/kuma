//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	commonv1alpha1 "github.com/kumahq/kuma/api/common/v1alpha1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterMatch) DeepCopyInto(out *ClusterMatch) {
	*out = *in
	if in.Origin != nil {
		in, out := &in.Origin, &out.Origin
		*out = new(string)
		**out = **in
	}
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterMatch.
func (in *ClusterMatch) DeepCopy() *ClusterMatch {
	if in == nil {
		return nil
	}
	out := new(ClusterMatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterMod) DeepCopyInto(out *ClusterMod) {
	*out = *in
	if in.Match != nil {
		in, out := &in.Match, &out.Match
		*out = new(ClusterMatch)
		(*in).DeepCopyInto(*out)
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
	if in.JsonPatches != nil {
		in, out := &in.JsonPatches, &out.JsonPatches
		*out = make([]commonv1alpha1.JsonPatchBlock, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterMod.
func (in *ClusterMod) DeepCopy() *ClusterMod {
	if in == nil {
		return nil
	}
	out := new(ClusterMod)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Conf) DeepCopyInto(out *Conf) {
	*out = *in
	if in.AppendModifications != nil {
		in, out := &in.AppendModifications, &out.AppendModifications
		*out = make([]Modification, len(*in))
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
func (in *HTTPFilterMatch) DeepCopyInto(out *HTTPFilterMatch) {
	*out = *in
	if in.Origin != nil {
		in, out := &in.Origin, &out.Origin
		*out = new(string)
		**out = **in
	}
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.ListenerName != nil {
		in, out := &in.ListenerName, &out.ListenerName
		*out = new(string)
		**out = **in
	}
	if in.ListenerTags != nil {
		in, out := &in.ListenerTags, &out.ListenerTags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPFilterMatch.
func (in *HTTPFilterMatch) DeepCopy() *HTTPFilterMatch {
	if in == nil {
		return nil
	}
	out := new(HTTPFilterMatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPFilterMod) DeepCopyInto(out *HTTPFilterMod) {
	*out = *in
	if in.Match != nil {
		in, out := &in.Match, &out.Match
		*out = new(HTTPFilterMatch)
		(*in).DeepCopyInto(*out)
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
	if in.JsonPatches != nil {
		in, out := &in.JsonPatches, &out.JsonPatches
		*out = make([]commonv1alpha1.JsonPatchBlock, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPFilterMod.
func (in *HTTPFilterMod) DeepCopy() *HTTPFilterMod {
	if in == nil {
		return nil
	}
	out := new(HTTPFilterMod)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ListenerMatch) DeepCopyInto(out *ListenerMatch) {
	*out = *in
	if in.Origin != nil {
		in, out := &in.Origin, &out.Origin
		*out = new(string)
		**out = **in
	}
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ListenerMatch.
func (in *ListenerMatch) DeepCopy() *ListenerMatch {
	if in == nil {
		return nil
	}
	out := new(ListenerMatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ListenerMod) DeepCopyInto(out *ListenerMod) {
	*out = *in
	if in.Match != nil {
		in, out := &in.Match, &out.Match
		*out = new(ListenerMatch)
		(*in).DeepCopyInto(*out)
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
	if in.JsonPatches != nil {
		in, out := &in.JsonPatches, &out.JsonPatches
		*out = make([]commonv1alpha1.JsonPatchBlock, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ListenerMod.
func (in *ListenerMod) DeepCopy() *ListenerMod {
	if in == nil {
		return nil
	}
	out := new(ListenerMod)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshProxyPatch) DeepCopyInto(out *MeshProxyPatch) {
	*out = *in
	in.TargetRef.DeepCopyInto(&out.TargetRef)
	in.Default.DeepCopyInto(&out.Default)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshProxyPatch.
func (in *MeshProxyPatch) DeepCopy() *MeshProxyPatch {
	if in == nil {
		return nil
	}
	out := new(MeshProxyPatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Modification) DeepCopyInto(out *Modification) {
	*out = *in
	if in.Cluster != nil {
		in, out := &in.Cluster, &out.Cluster
		*out = new(ClusterMod)
		(*in).DeepCopyInto(*out)
	}
	if in.Listener != nil {
		in, out := &in.Listener, &out.Listener
		*out = new(ListenerMod)
		(*in).DeepCopyInto(*out)
	}
	if in.NetworkFilter != nil {
		in, out := &in.NetworkFilter, &out.NetworkFilter
		*out = new(NetworkFilterMod)
		(*in).DeepCopyInto(*out)
	}
	if in.HTTPFilter != nil {
		in, out := &in.HTTPFilter, &out.HTTPFilter
		*out = new(HTTPFilterMod)
		(*in).DeepCopyInto(*out)
	}
	if in.VirtualHost != nil {
		in, out := &in.VirtualHost, &out.VirtualHost
		*out = new(VirtualHostMod)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Modification.
func (in *Modification) DeepCopy() *Modification {
	if in == nil {
		return nil
	}
	out := new(Modification)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkFilterMatch) DeepCopyInto(out *NetworkFilterMatch) {
	*out = *in
	if in.Origin != nil {
		in, out := &in.Origin, &out.Origin
		*out = new(string)
		**out = **in
	}
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.ListenerName != nil {
		in, out := &in.ListenerName, &out.ListenerName
		*out = new(string)
		**out = **in
	}
	if in.ListenerTags != nil {
		in, out := &in.ListenerTags, &out.ListenerTags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkFilterMatch.
func (in *NetworkFilterMatch) DeepCopy() *NetworkFilterMatch {
	if in == nil {
		return nil
	}
	out := new(NetworkFilterMatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkFilterMod) DeepCopyInto(out *NetworkFilterMod) {
	*out = *in
	if in.Match != nil {
		in, out := &in.Match, &out.Match
		*out = new(NetworkFilterMatch)
		(*in).DeepCopyInto(*out)
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
	if in.JsonPatches != nil {
		in, out := &in.JsonPatches, &out.JsonPatches
		*out = make([]commonv1alpha1.JsonPatchBlock, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkFilterMod.
func (in *NetworkFilterMod) DeepCopy() *NetworkFilterMod {
	if in == nil {
		return nil
	}
	out := new(NetworkFilterMod)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualHostMatch) DeepCopyInto(out *VirtualHostMatch) {
	*out = *in
	if in.Origin != nil {
		in, out := &in.Origin, &out.Origin
		*out = new(string)
		**out = **in
	}
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.RouteConfigurationName != nil {
		in, out := &in.RouteConfigurationName, &out.RouteConfigurationName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualHostMatch.
func (in *VirtualHostMatch) DeepCopy() *VirtualHostMatch {
	if in == nil {
		return nil
	}
	out := new(VirtualHostMatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualHostMod) DeepCopyInto(out *VirtualHostMod) {
	*out = *in
	if in.Match != nil {
		in, out := &in.Match, &out.Match
		*out = new(VirtualHostMatch)
		(*in).DeepCopyInto(*out)
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
	if in.JsonPatches != nil {
		in, out := &in.JsonPatches, &out.JsonPatches
		*out = make([]commonv1alpha1.JsonPatchBlock, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualHostMod.
func (in *VirtualHostMod) DeepCopy() *VirtualHostMod {
	if in == nil {
		return nil
	}
	out := new(VirtualHostMod)
	in.DeepCopyInto(out)
	return out
}
