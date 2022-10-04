// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.20.0
// source: pkg/plugins/policies/meshtrafficpermission/api/v1alpha1/meshtrafficpermission.proto

package v1alpha1

import (
	v1alpha1 "github.com/kumahq/kuma/api/common/v1alpha1"
	_ "github.com/kumahq/kuma/api/mesh"
	_ "github.com/kumahq/protoc-gen-kumadoc/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type MeshTrafficPermission_Conf_Action int32

const (
	MeshTrafficPermission_Conf_ALLOW                  MeshTrafficPermission_Conf_Action = 0
	MeshTrafficPermission_Conf_DENY                   MeshTrafficPermission_Conf_Action = 1
	MeshTrafficPermission_Conf_ALLOW_WITH_SHADOW_DENY MeshTrafficPermission_Conf_Action = 2
)

// Enum value maps for MeshTrafficPermission_Conf_Action.
var (
	MeshTrafficPermission_Conf_Action_name = map[int32]string{
		0: "ALLOW",
		1: "DENY",
		2: "ALLOW_WITH_SHADOW_DENY",
	}
	MeshTrafficPermission_Conf_Action_value = map[string]int32{
		"ALLOW":                  0,
		"DENY":                   1,
		"ALLOW_WITH_SHADOW_DENY": 2,
	}
)

func (x MeshTrafficPermission_Conf_Action) Enum() *MeshTrafficPermission_Conf_Action {
	p := new(MeshTrafficPermission_Conf_Action)
	*p = x
	return p
}

func (x MeshTrafficPermission_Conf_Action) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MeshTrafficPermission_Conf_Action) Descriptor() protoreflect.EnumDescriptor {
	return file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_enumTypes[0].Descriptor()
}

func (MeshTrafficPermission_Conf_Action) Type() protoreflect.EnumType {
	return &file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_enumTypes[0]
}

func (x MeshTrafficPermission_Conf_Action) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MeshTrafficPermission_Conf_Action.Descriptor instead.
func (MeshTrafficPermission_Conf_Action) EnumDescriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDescGZIP(), []int{0, 0, 0}
}

// MeshTrafficPermission defines permission for traffic between data planes
// proxies.
type MeshTrafficPermission struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef *v1alpha1.TargetRef `protobuf:"bytes,1,opt,name=targetRef,proto3" json:"targetRef,omitempty"`
	// From is a list of pairs – a group of clients and action applied for it
	From []*MeshTrafficPermission_From `protobuf:"bytes,3,rep,name=from,proto3" json:"from,omitempty"`
}

func (x *MeshTrafficPermission) Reset() {
	*x = MeshTrafficPermission{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshTrafficPermission) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshTrafficPermission) ProtoMessage() {}

func (x *MeshTrafficPermission) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshTrafficPermission.ProtoReflect.Descriptor instead.
func (*MeshTrafficPermission) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDescGZIP(), []int{0}
}

func (x *MeshTrafficPermission) GetTargetRef() *v1alpha1.TargetRef {
	if x != nil {
		return x.TargetRef
	}
	return nil
}

func (x *MeshTrafficPermission) GetFrom() []*MeshTrafficPermission_From {
	if x != nil {
		return x.From
	}
	return nil
}

type MeshTrafficPermission_Conf struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Action defines a behaviour for the specified group of clients:
	//  * ALLOW - lets the requests pass
	//  * DENY - blocks the requests
	//  * ALLOW_WITH_SHADOW_DENY - lets the requests pass but emits logs as if
	//  requests are denied
	//  * DENY_WITH_SHADOW_ALLOW - blocks the requests but emits logs as if
	//  requests are allowed
	// +kubebuilder:validation:Enum=ALLOW;DENY;ALLOW_WITH_SHADOW_DENY;DENY_WITH_SHADOW_ALLOW
	Action string `protobuf:"bytes,1,opt,name=action,proto3" json:"action,omitempty"`
}

func (x *MeshTrafficPermission_Conf) Reset() {
	*x = MeshTrafficPermission_Conf{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshTrafficPermission_Conf) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshTrafficPermission_Conf) ProtoMessage() {}

func (x *MeshTrafficPermission_Conf) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshTrafficPermission_Conf.ProtoReflect.Descriptor instead.
func (*MeshTrafficPermission_Conf) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDescGZIP(), []int{0, 0}
}

func (x *MeshTrafficPermission_Conf) GetAction() string {
	if x != nil {
		return x.Action
	}
	return ""
}

type MeshTrafficPermission_From struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TargetRef is a reference to the resource that represents a group of
	// clients.
	TargetRef *v1alpha1.TargetRef `protobuf:"bytes,1,opt,name=targetRef,proto3" json:"targetRef,omitempty"`
	// Default is a configuration specific to the group of clients referenced in
	// 'targetRef'
	Default *MeshTrafficPermission_Conf `protobuf:"bytes,2,opt,name=default,proto3" json:"default,omitempty"`
}

func (x *MeshTrafficPermission_From) Reset() {
	*x = MeshTrafficPermission_From{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshTrafficPermission_From) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshTrafficPermission_From) ProtoMessage() {}

func (x *MeshTrafficPermission_From) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshTrafficPermission_From.ProtoReflect.Descriptor instead.
func (*MeshTrafficPermission_From) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDescGZIP(), []int{0, 1}
}

func (x *MeshTrafficPermission_From) GetTargetRef() *v1alpha1.TargetRef {
	if x != nil {
		return x.TargetRef
	}
	return nil
}

func (x *MeshTrafficPermission_From) GetDefault() *MeshTrafficPermission_Conf {
	if x != nil {
		return x.Default
	}
	return nil
}

var File_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto protoreflect.FileDescriptor

var file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDesc = []byte{
	0x0a, 0x53, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x70, 0x6f,
	0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x68, 0x74, 0x72, 0x61, 0x66, 0x66,
	0x69, 0x63, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x6d, 0x65, 0x73, 0x68, 0x74, 0x72,
	0x61, 0x66, 0x66, 0x69, 0x63, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x34, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x73, 0x2e, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73,
	0x68, 0x74, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x12, 0x6d, 0x65, 0x73,
	0x68, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2f, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x72, 0x65, 0x66, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x15, 0x6b, 0x75, 0x6d, 0x61, 0x2d, 0x64, 0x6f, 0x63, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xdf, 0x03, 0x0a, 0x15, 0x4d, 0x65, 0x73, 0x68,
	0x54, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x3d, 0x0a, 0x09, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x52, 0x65, 0x66, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x52, 0x65, 0x66, 0x52, 0x09, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x52, 0x65, 0x66,
	0x12, 0x64, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x50,
	0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x70, 0x6f,
	0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x74, 0x72, 0x61, 0x66, 0x66,
	0x69, 0x63, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x54, 0x72, 0x61, 0x66, 0x66, 0x69,
	0x63, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x46, 0x72, 0x6f, 0x6d,
	0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x1a, 0x59, 0x0a, 0x04, 0x43, 0x6f, 0x6e, 0x66, 0x12, 0x16,
	0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x39, 0x0a, 0x06, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x09, 0x0a, 0x05, 0x41, 0x4c, 0x4c, 0x4f, 0x57, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x44,
	0x45, 0x4e, 0x59, 0x10, 0x01, 0x12, 0x1a, 0x0a, 0x16, 0x41, 0x4c, 0x4c, 0x4f, 0x57, 0x5f, 0x57,
	0x49, 0x54, 0x48, 0x5f, 0x53, 0x48, 0x41, 0x44, 0x4f, 0x57, 0x5f, 0x44, 0x45, 0x4e, 0x59, 0x10,
	0x02, 0x1a, 0xbd, 0x01, 0x0a, 0x04, 0x46, 0x72, 0x6f, 0x6d, 0x12, 0x43, 0x0a, 0x09, 0x74, 0x61,
	0x72, 0x67, 0x65, 0x74, 0x52, 0x65, 0x66, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e,
	0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x52, 0x65, 0x66, 0x42, 0x04,
	0x88, 0xb5, 0x18, 0x01, 0x52, 0x09, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x52, 0x65, 0x66, 0x12,
	0x70, 0x0a, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x50, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e,
	0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x74, 0x72, 0x61,
	0x66, 0x66, 0x69, 0x63, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x54, 0x72, 0x61, 0x66,
	0x66, 0x69, 0x63, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x43, 0x6f,
	0x6e, 0x66, 0x42, 0x04, 0x88, 0xb5, 0x18, 0x01, 0x52, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c,
	0x74, 0x3a, 0x06, 0xb2, 0x8c, 0x89, 0xa6, 0x01, 0x00, 0x42, 0x86, 0x01, 0x5a, 0x4e, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x68, 0x71, 0x2f,
	0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73,
	0x2f, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x68, 0x74, 0x72,
	0x61, 0x66, 0x66, 0x69, 0x63, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x8a, 0xb5, 0x18, 0x32,
	0x50, 0x01, 0xa2, 0x01, 0x15, 0x4d, 0x65, 0x73, 0x68, 0x54, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63,
	0x50, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0xf2, 0x01, 0x15, 0x6d, 0x65, 0x73,
	0x68, 0x74, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDescOnce sync.Once
	file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDescData = file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDesc
)

func file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDescGZIP() []byte {
	file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDescOnce.Do(func() {
		file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDescData)
	})
	return file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDescData
}

var file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_goTypes = []interface{}{
	(MeshTrafficPermission_Conf_Action)(0), // 0: kuma.plugins.policies.meshtrafficpermission.v1alpha1.MeshTrafficPermission.Conf.Action
	(*MeshTrafficPermission)(nil),          // 1: kuma.plugins.policies.meshtrafficpermission.v1alpha1.MeshTrafficPermission
	(*MeshTrafficPermission_Conf)(nil),     // 2: kuma.plugins.policies.meshtrafficpermission.v1alpha1.MeshTrafficPermission.Conf
	(*MeshTrafficPermission_From)(nil),     // 3: kuma.plugins.policies.meshtrafficpermission.v1alpha1.MeshTrafficPermission.From
	(*v1alpha1.TargetRef)(nil),             // 4: kuma.common.v1alpha1.TargetRef
}
var file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_depIdxs = []int32{
	4, // 0: kuma.plugins.policies.meshtrafficpermission.v1alpha1.MeshTrafficPermission.targetRef:type_name -> kuma.common.v1alpha1.TargetRef
	3, // 1: kuma.plugins.policies.meshtrafficpermission.v1alpha1.MeshTrafficPermission.from:type_name -> kuma.plugins.policies.meshtrafficpermission.v1alpha1.MeshTrafficPermission.From
	4, // 2: kuma.plugins.policies.meshtrafficpermission.v1alpha1.MeshTrafficPermission.From.targetRef:type_name -> kuma.common.v1alpha1.TargetRef
	2, // 3: kuma.plugins.policies.meshtrafficpermission.v1alpha1.MeshTrafficPermission.From.default:type_name -> kuma.plugins.policies.meshtrafficpermission.v1alpha1.MeshTrafficPermission.Conf
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() {
	file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_init()
}
func file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_init() {
	if File_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshTrafficPermission); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshTrafficPermission_Conf); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshTrafficPermission_From); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_goTypes,
		DependencyIndexes: file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_depIdxs,
		EnumInfos:         file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_enumTypes,
		MessageInfos:      file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_msgTypes,
	}.Build()
	File_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto = out.File
	file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_rawDesc = nil
	file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_goTypes = nil
	file_pkg_plugins_policies_meshtrafficpermission_api_v1alpha1_meshtrafficpermission_proto_depIdxs = nil
}
