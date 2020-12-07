// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.6.1
// source: mesh/v1alpha1/dataplane_insight.proto

package v1alpha1

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// DataplaneInsight defines the observed state of a Dataplane.
type DataplaneInsight struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// List of ADS subscriptions created by a given Dataplane.
	Subscriptions []*DiscoverySubscription `protobuf:"bytes,1,rep,name=subscriptions,proto3" json:"subscriptions,omitempty"`
	// Insights about mTLS for Dataplane.
	MTLS *DataplaneInsight_MTLS `protobuf:"bytes,2,opt,name=mTLS,proto3" json:"mTLS,omitempty"`
}

func (x *DataplaneInsight) Reset() {
	*x = DataplaneInsight{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataplaneInsight) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataplaneInsight) ProtoMessage() {}

func (x *DataplaneInsight) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataplaneInsight.ProtoReflect.Descriptor instead.
func (*DataplaneInsight) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_dataplane_insight_proto_rawDescGZIP(), []int{0}
}

func (x *DataplaneInsight) GetSubscriptions() []*DiscoverySubscription {
	if x != nil {
		return x.Subscriptions
	}
	return nil
}

func (x *DataplaneInsight) GetMTLS() *DataplaneInsight_MTLS {
	if x != nil {
		return x.MTLS
	}
	return nil
}

// DiscoverySubscription describes a single ADS subscription
// created by a Dataplane to the Control Plane.
// Ideally, there should be only one such subscription per Dataplane lifecycle.
// Presence of multiple subscriptions might indicate one of the following
// events:
// - transient loss of network connection between Dataplane and Control Plane
// - Dataplane restart (i.e. hot restart or crash)
// - Control Plane restart (i.e. rolling update or crash)
// - etc
type DiscoverySubscription struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique id per ADS subscription.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Control Plane instance that handled given subscription.
	ControlPlaneInstanceId string `protobuf:"bytes,2,opt,name=control_plane_instance_id,json=controlPlaneInstanceId,proto3" json:"control_plane_instance_id,omitempty"`
	// Time when a given Dataplane connected to the Control Plane.
	ConnectTime *timestamp.Timestamp `protobuf:"bytes,3,opt,name=connect_time,json=connectTime,proto3" json:"connect_time,omitempty"`
	// Time when a given Dataplane disconnected from the Control Plane.
	DisconnectTime *timestamp.Timestamp `protobuf:"bytes,4,opt,name=disconnect_time,json=disconnectTime,proto3" json:"disconnect_time,omitempty"`
	// Status of the ADS subscription.
	Status *DiscoverySubscriptionStatus `protobuf:"bytes,5,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *DiscoverySubscription) Reset() {
	*x = DiscoverySubscription{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DiscoverySubscription) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DiscoverySubscription) ProtoMessage() {}

func (x *DiscoverySubscription) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DiscoverySubscription.ProtoReflect.Descriptor instead.
func (*DiscoverySubscription) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_dataplane_insight_proto_rawDescGZIP(), []int{1}
}

func (x *DiscoverySubscription) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *DiscoverySubscription) GetControlPlaneInstanceId() string {
	if x != nil {
		return x.ControlPlaneInstanceId
	}
	return ""
}

func (x *DiscoverySubscription) GetConnectTime() *timestamp.Timestamp {
	if x != nil {
		return x.ConnectTime
	}
	return nil
}

func (x *DiscoverySubscription) GetDisconnectTime() *timestamp.Timestamp {
	if x != nil {
		return x.DisconnectTime
	}
	return nil
}

func (x *DiscoverySubscription) GetStatus() *DiscoverySubscriptionStatus {
	if x != nil {
		return x.Status
	}
	return nil
}

// DiscoverySubscriptionStatus defines status of an ADS subscription.
type DiscoverySubscriptionStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Time when status of a given ADS subscription was most recently updated.
	LastUpdateTime *timestamp.Timestamp `protobuf:"bytes,1,opt,name=last_update_time,json=lastUpdateTime,proto3" json:"last_update_time,omitempty"`
	// Total defines an aggregate over individual xDS stats.
	Total *DiscoveryServiceStats `protobuf:"bytes,2,opt,name=total,proto3" json:"total,omitempty"`
	// CDS defines all CDS stats.
	Cds *DiscoveryServiceStats `protobuf:"bytes,3,opt,name=cds,proto3" json:"cds,omitempty"`
	// EDS defines all EDS stats.
	Eds *DiscoveryServiceStats `protobuf:"bytes,4,opt,name=eds,proto3" json:"eds,omitempty"`
	// LDS defines all LDS stats.
	Lds *DiscoveryServiceStats `protobuf:"bytes,5,opt,name=lds,proto3" json:"lds,omitempty"`
	// RDS defines all RDS stats.
	Rds *DiscoveryServiceStats `protobuf:"bytes,6,opt,name=rds,proto3" json:"rds,omitempty"`
}

func (x *DiscoverySubscriptionStatus) Reset() {
	*x = DiscoverySubscriptionStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DiscoverySubscriptionStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DiscoverySubscriptionStatus) ProtoMessage() {}

func (x *DiscoverySubscriptionStatus) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DiscoverySubscriptionStatus.ProtoReflect.Descriptor instead.
func (*DiscoverySubscriptionStatus) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_dataplane_insight_proto_rawDescGZIP(), []int{2}
}

func (x *DiscoverySubscriptionStatus) GetLastUpdateTime() *timestamp.Timestamp {
	if x != nil {
		return x.LastUpdateTime
	}
	return nil
}

func (x *DiscoverySubscriptionStatus) GetTotal() *DiscoveryServiceStats {
	if x != nil {
		return x.Total
	}
	return nil
}

func (x *DiscoverySubscriptionStatus) GetCds() *DiscoveryServiceStats {
	if x != nil {
		return x.Cds
	}
	return nil
}

func (x *DiscoverySubscriptionStatus) GetEds() *DiscoveryServiceStats {
	if x != nil {
		return x.Eds
	}
	return nil
}

func (x *DiscoverySubscriptionStatus) GetLds() *DiscoveryServiceStats {
	if x != nil {
		return x.Lds
	}
	return nil
}

func (x *DiscoverySubscriptionStatus) GetRds() *DiscoveryServiceStats {
	if x != nil {
		return x.Rds
	}
	return nil
}

// DiscoveryServiceStats defines all stats over a single xDS service.
type DiscoveryServiceStats struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Number of xDS responses sent to the Dataplane.
	ResponsesSent uint64 `protobuf:"varint,1,opt,name=responses_sent,json=responsesSent,proto3" json:"responses_sent,omitempty"`
	// Number of xDS responses ACKed by the Dataplane.
	ResponsesAcknowledged uint64 `protobuf:"varint,2,opt,name=responses_acknowledged,json=responsesAcknowledged,proto3" json:"responses_acknowledged,omitempty"`
	// Number of xDS responses NACKed by the Dataplane.
	ResponsesRejected uint64 `protobuf:"varint,3,opt,name=responses_rejected,json=responsesRejected,proto3" json:"responses_rejected,omitempty"`
}

func (x *DiscoveryServiceStats) Reset() {
	*x = DiscoveryServiceStats{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DiscoveryServiceStats) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DiscoveryServiceStats) ProtoMessage() {}

func (x *DiscoveryServiceStats) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DiscoveryServiceStats.ProtoReflect.Descriptor instead.
func (*DiscoveryServiceStats) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_dataplane_insight_proto_rawDescGZIP(), []int{3}
}

func (x *DiscoveryServiceStats) GetResponsesSent() uint64 {
	if x != nil {
		return x.ResponsesSent
	}
	return 0
}

func (x *DiscoveryServiceStats) GetResponsesAcknowledged() uint64 {
	if x != nil {
		return x.ResponsesAcknowledged
	}
	return 0
}

func (x *DiscoveryServiceStats) GetResponsesRejected() uint64 {
	if x != nil {
		return x.ResponsesRejected
	}
	return 0
}

// MTLS defines insights for mTLS
type DataplaneInsight_MTLS struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Expiration time of the last certificate that was generated for a
	// Dataplane.
	CertificateExpirationTime *timestamp.Timestamp `protobuf:"bytes,1,opt,name=certificate_expiration_time,json=certificateExpirationTime,proto3" json:"certificate_expiration_time,omitempty"`
	// Time on which the last certificate was generated.
	LastCertificateRegeneration *timestamp.Timestamp `protobuf:"bytes,2,opt,name=last_certificate_regeneration,json=lastCertificateRegeneration,proto3" json:"last_certificate_regeneration,omitempty"`
	// Number of certificate regenerations for a Dataplane.
	CertificateRegenerations uint32 `protobuf:"varint,3,opt,name=certificate_regenerations,json=certificateRegenerations,proto3" json:"certificate_regenerations,omitempty"`
}

func (x *DataplaneInsight_MTLS) Reset() {
	*x = DataplaneInsight_MTLS{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataplaneInsight_MTLS) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataplaneInsight_MTLS) ProtoMessage() {}

func (x *DataplaneInsight_MTLS) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataplaneInsight_MTLS.ProtoReflect.Descriptor instead.
func (*DataplaneInsight_MTLS) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_dataplane_insight_proto_rawDescGZIP(), []int{0, 0}
}

func (x *DataplaneInsight_MTLS) GetCertificateExpirationTime() *timestamp.Timestamp {
	if x != nil {
		return x.CertificateExpirationTime
	}
	return nil
}

func (x *DataplaneInsight_MTLS) GetLastCertificateRegeneration() *timestamp.Timestamp {
	if x != nil {
		return x.LastCertificateRegeneration
	}
	return nil
}

func (x *DataplaneInsight_MTLS) GetCertificateRegenerations() uint32 {
	if x != nil {
		return x.CertificateRegenerations
	}
	return 0
}

var File_mesh_v1alpha1_dataplane_insight_proto protoreflect.FileDescriptor

var file_mesh_v1alpha1_dataplane_insight_proto_rawDesc = []byte{
	0x0a, 0x25, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x64, 0x61, 0x74, 0x61, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x5f, 0x69, 0x6e, 0x73, 0x69, 0x67, 0x68,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65,
	0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61,
	0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa4, 0x03, 0x0a, 0x10, 0x44, 0x61, 0x74, 0x61, 0x70, 0x6c,
	0x61, 0x6e, 0x65, 0x49, 0x6e, 0x73, 0x69, 0x67, 0x68, 0x74, 0x12, 0x4f, 0x0a, 0x0d, 0x73, 0x75,
	0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x29, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79,
	0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0d, 0x73, 0x75,
	0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x3d, 0x0a, 0x04, 0x6d,
	0x54, 0x4c, 0x53, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6b, 0x75, 0x6d, 0x61,
	0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44,
	0x61, 0x74, 0x61, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x49, 0x6e, 0x73, 0x69, 0x67, 0x68, 0x74, 0x2e,
	0x4d, 0x54, 0x4c, 0x53, 0x52, 0x04, 0x6d, 0x54, 0x4c, 0x53, 0x1a, 0xff, 0x01, 0x0a, 0x04, 0x4d,
	0x54, 0x4c, 0x53, 0x12, 0x5a, 0x0a, 0x1b, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x65, 0x5f, 0x65, 0x78, 0x70, 0x69, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x19, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74,
	0x65, 0x45, 0x78, 0x70, 0x69, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x12,
	0x5e, 0x0a, 0x1d, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x5f, 0x72, 0x65, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x1b, 0x6c, 0x61, 0x73, 0x74, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x3b, 0x0a, 0x19, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x72,
	0x65, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x18, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0xd5, 0x02, 0x0a,
	0x15, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x17, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x42, 0x0a, 0x19, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x5f, 0x70, 0x6c, 0x61, 0x6e, 0x65,
	0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x16, 0x63, 0x6f, 0x6e,
	0x74, 0x72, 0x6f, 0x6c, 0x50, 0x6c, 0x61, 0x6e, 0x65, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x49, 0x64, 0x12, 0x47, 0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x08, 0xfa, 0x42, 0x05, 0xb2, 0x01, 0x02, 0x08, 0x01, 0x52,
	0x0b, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x43, 0x0a, 0x0f,
	0x64, 0x69, 0x73, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x52, 0x0e, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x54, 0x69, 0x6d,
	0x65, 0x12, 0x51, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x2f, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79,
	0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x8a, 0x01, 0x02, 0x10, 0x01, 0x52, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x22, 0x98, 0x03, 0x0a, 0x1b, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65,
	0x72, 0x79, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x44, 0x0a, 0x10, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x75, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0e, 0x6c, 0x61, 0x73, 0x74,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x3f, 0x0a, 0x05, 0x74, 0x6f,
	0x74, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6b, 0x75, 0x6d, 0x61,
	0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44,
	0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53,
	0x74, 0x61, 0x74, 0x73, 0x52, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x12, 0x3b, 0x0a, 0x03, 0x63,
	0x64, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e,
	0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x69,
	0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x74,
	0x61, 0x74, 0x73, 0x52, 0x03, 0x63, 0x64, 0x73, 0x12, 0x3b, 0x0a, 0x03, 0x65, 0x64, 0x73, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73,
	0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x69, 0x73, 0x63, 0x6f,
	0x76, 0x65, 0x72, 0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x73,
	0x52, 0x03, 0x65, 0x64, 0x73, 0x12, 0x3b, 0x0a, 0x03, 0x6c, 0x64, 0x73, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72,
	0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52, 0x03, 0x6c,
	0x64, 0x73, 0x12, 0x3b, 0x0a, 0x03, 0x72, 0x64, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x29, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52, 0x03, 0x72, 0x64, 0x73, 0x22,
	0xa4, 0x01, 0x0a, 0x15, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x73, 0x12, 0x25, 0x0a, 0x0e, 0x72, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x5f, 0x73, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x0d, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x53, 0x65, 0x6e, 0x74,
	0x12, 0x35, 0x0a, 0x16, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x5f, 0x61, 0x63,
	0x6b, 0x6e, 0x6f, 0x77, 0x6c, 0x65, 0x64, 0x67, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x15, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x41, 0x63, 0x6b, 0x6e, 0x6f,
	0x77, 0x6c, 0x65, 0x64, 0x67, 0x65, 0x64, 0x12, 0x2d, 0x0a, 0x12, 0x72, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x73, 0x5f, 0x72, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x65, 0x64, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x11, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x52, 0x65,
	0x6a, 0x65, 0x63, 0x74, 0x65, 0x64, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mesh_v1alpha1_dataplane_insight_proto_rawDescOnce sync.Once
	file_mesh_v1alpha1_dataplane_insight_proto_rawDescData = file_mesh_v1alpha1_dataplane_insight_proto_rawDesc
)

func file_mesh_v1alpha1_dataplane_insight_proto_rawDescGZIP() []byte {
	file_mesh_v1alpha1_dataplane_insight_proto_rawDescOnce.Do(func() {
		file_mesh_v1alpha1_dataplane_insight_proto_rawDescData = protoimpl.X.CompressGZIP(file_mesh_v1alpha1_dataplane_insight_proto_rawDescData)
	})
	return file_mesh_v1alpha1_dataplane_insight_proto_rawDescData
}

var file_mesh_v1alpha1_dataplane_insight_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_mesh_v1alpha1_dataplane_insight_proto_goTypes = []interface{}{
	(*DataplaneInsight)(nil),            // 0: kuma.mesh.v1alpha1.DataplaneInsight
	(*DiscoverySubscription)(nil),       // 1: kuma.mesh.v1alpha1.DiscoverySubscription
	(*DiscoverySubscriptionStatus)(nil), // 2: kuma.mesh.v1alpha1.DiscoverySubscriptionStatus
	(*DiscoveryServiceStats)(nil),       // 3: kuma.mesh.v1alpha1.DiscoveryServiceStats
	(*DataplaneInsight_MTLS)(nil),       // 4: kuma.mesh.v1alpha1.DataplaneInsight.MTLS
	(*timestamp.Timestamp)(nil),         // 5: google.protobuf.Timestamp
}
var file_mesh_v1alpha1_dataplane_insight_proto_depIdxs = []int32{
	1,  // 0: kuma.mesh.v1alpha1.DataplaneInsight.subscriptions:type_name -> kuma.mesh.v1alpha1.DiscoverySubscription
	4,  // 1: kuma.mesh.v1alpha1.DataplaneInsight.mTLS:type_name -> kuma.mesh.v1alpha1.DataplaneInsight.MTLS
	5,  // 2: kuma.mesh.v1alpha1.DiscoverySubscription.connect_time:type_name -> google.protobuf.Timestamp
	5,  // 3: kuma.mesh.v1alpha1.DiscoverySubscription.disconnect_time:type_name -> google.protobuf.Timestamp
	2,  // 4: kuma.mesh.v1alpha1.DiscoverySubscription.status:type_name -> kuma.mesh.v1alpha1.DiscoverySubscriptionStatus
	5,  // 5: kuma.mesh.v1alpha1.DiscoverySubscriptionStatus.last_update_time:type_name -> google.protobuf.Timestamp
	3,  // 6: kuma.mesh.v1alpha1.DiscoverySubscriptionStatus.total:type_name -> kuma.mesh.v1alpha1.DiscoveryServiceStats
	3,  // 7: kuma.mesh.v1alpha1.DiscoverySubscriptionStatus.cds:type_name -> kuma.mesh.v1alpha1.DiscoveryServiceStats
	3,  // 8: kuma.mesh.v1alpha1.DiscoverySubscriptionStatus.eds:type_name -> kuma.mesh.v1alpha1.DiscoveryServiceStats
	3,  // 9: kuma.mesh.v1alpha1.DiscoverySubscriptionStatus.lds:type_name -> kuma.mesh.v1alpha1.DiscoveryServiceStats
	3,  // 10: kuma.mesh.v1alpha1.DiscoverySubscriptionStatus.rds:type_name -> kuma.mesh.v1alpha1.DiscoveryServiceStats
	5,  // 11: kuma.mesh.v1alpha1.DataplaneInsight.MTLS.certificate_expiration_time:type_name -> google.protobuf.Timestamp
	5,  // 12: kuma.mesh.v1alpha1.DataplaneInsight.MTLS.last_certificate_regeneration:type_name -> google.protobuf.Timestamp
	13, // [13:13] is the sub-list for method output_type
	13, // [13:13] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_mesh_v1alpha1_dataplane_insight_proto_init() }
func file_mesh_v1alpha1_dataplane_insight_proto_init() {
	if File_mesh_v1alpha1_dataplane_insight_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataplaneInsight); i {
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
		file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DiscoverySubscription); i {
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
		file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DiscoverySubscriptionStatus); i {
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
		file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DiscoveryServiceStats); i {
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
		file_mesh_v1alpha1_dataplane_insight_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataplaneInsight_MTLS); i {
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
			RawDescriptor: file_mesh_v1alpha1_dataplane_insight_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mesh_v1alpha1_dataplane_insight_proto_goTypes,
		DependencyIndexes: file_mesh_v1alpha1_dataplane_insight_proto_depIdxs,
		MessageInfos:      file_mesh_v1alpha1_dataplane_insight_proto_msgTypes,
	}.Build()
	File_mesh_v1alpha1_dataplane_insight_proto = out.File
	file_mesh_v1alpha1_dataplane_insight_proto_rawDesc = nil
	file_mesh_v1alpha1_dataplane_insight_proto_goTypes = nil
	file_mesh_v1alpha1_dataplane_insight_proto_depIdxs = nil
}
