// Code generated by protoc-gen-deepcopy. DO NOT EDIT.

package v1alpha1

import (
	proto "google.golang.org/protobuf/proto"
)

// DeepCopyInto supports using MeshTrace within kubernetes types, where deepcopy-gen is used.
func (in *MeshTrace) DeepCopyInto(out *MeshTrace) {
	p := proto.Clone(in).(*MeshTrace)
	*out = *p
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshTrace. Required by controller-gen.
func (in *MeshTrace) DeepCopy() *MeshTrace {
	if in == nil {
		return nil
	}
	out := new(MeshTrace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInterface is an autogenerated deepcopy function, copying the receiver, creating a new MeshTrace. Required by controller-gen.
func (in *MeshTrace) DeepCopyInterface() interface{} {
	return in.DeepCopy()
}
