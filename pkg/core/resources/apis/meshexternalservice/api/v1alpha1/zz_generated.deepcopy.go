//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshExternalService) DeepCopyInto(out *MeshExternalService) {
	*out = *in
	in.TargetRef.DeepCopyInto(&out.TargetRef)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshExternalService.
func (in *MeshExternalService) DeepCopy() *MeshExternalService {
	if in == nil {
		return nil
	}
	out := new(MeshExternalService)
	in.DeepCopyInto(out)
	return out
}
