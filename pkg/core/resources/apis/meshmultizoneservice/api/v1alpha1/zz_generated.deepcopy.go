//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	apiv1alpha1 "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	meshserviceapiv1alpha1 "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MatchedMeshService) DeepCopyInto(out *MatchedMeshService) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MatchedMeshService.
func (in *MatchedMeshService) DeepCopy() *MatchedMeshService {
	if in == nil {
		return nil
	}
	out := new(MatchedMeshService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshMultiZoneService) DeepCopyInto(out *MeshMultiZoneService) {
	*out = *in
	in.Selector.DeepCopyInto(&out.Selector)
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]Port, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshMultiZoneService.
func (in *MeshMultiZoneService) DeepCopy() *MeshMultiZoneService {
	if in == nil {
		return nil
	}
	out := new(MeshMultiZoneService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshMultiZoneServiceStatus) DeepCopyInto(out *MeshMultiZoneServiceStatus) {
	*out = *in
	if in.Addresses != nil {
		in, out := &in.Addresses, &out.Addresses
		*out = make([]apiv1alpha1.Address, len(*in))
		copy(*out, *in)
	}
	if in.VIPs != nil {
		in, out := &in.VIPs, &out.VIPs
		*out = make([]meshserviceapiv1alpha1.VIP, len(*in))
		copy(*out, *in)
	}
	if in.MeshServices != nil {
		in, out := &in.MeshServices, &out.MeshServices
		*out = make([]MatchedMeshService, len(*in))
		copy(*out, *in)
	}
	if in.HostnameGenerators != nil {
		in, out := &in.HostnameGenerators, &out.HostnameGenerators
		*out = make([]apiv1alpha1.HostnameGeneratorStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshMultiZoneServiceStatus.
func (in *MeshMultiZoneServiceStatus) DeepCopy() *MeshMultiZoneServiceStatus {
	if in == nil {
		return nil
	}
	out := new(MeshMultiZoneServiceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshServiceSelector) DeepCopyInto(out *MeshServiceSelector) {
	*out = *in
	if in.MatchLabels != nil {
		in, out := &in.MatchLabels, &out.MatchLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshServiceSelector.
func (in *MeshServiceSelector) DeepCopy() *MeshServiceSelector {
	if in == nil {
		return nil
	}
	out := new(MeshServiceSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Port) DeepCopyInto(out *Port) {
	*out = *in
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Port.
func (in *Port) DeepCopy() *Port {
	if in == nil {
		return nil
	}
	out := new(Port)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Selector) DeepCopyInto(out *Selector) {
	*out = *in
	in.MeshService.DeepCopyInto(&out.MeshService)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Selector.
func (in *Selector) DeepCopy() *Selector {
	if in == nil {
		return nil
	}
	out := new(Selector)
	in.DeepCopyInto(out)
	return out
}
