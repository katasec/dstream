// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: proto/ingester.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Config or startup options (e.g., JSON blob)
type StreamRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ConfigJson    string                 `protobuf:"bytes,1,opt,name=config_json,json=configJson,proto3" json:"config_json,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StreamRequest) Reset() {
	*x = StreamRequest{}
	mi := &file_proto_ingester_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StreamRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamRequest) ProtoMessage() {}

func (x *StreamRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_ingester_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamRequest.ProtoReflect.Descriptor instead.
func (*StreamRequest) Descriptor() ([]byte, []int) {
	return file_proto_ingester_proto_rawDescGZIP(), []int{0}
}

func (x *StreamRequest) GetConfigJson() string {
	if x != nil {
		return x.ConfigJson
	}
	return ""
}

// A single event emitted by the plugin (payload as JSON for now)
type Event struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	JsonPayload   string                 `protobuf:"bytes,1,opt,name=json_payload,json=jsonPayload,proto3" json:"json_payload,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Event) Reset() {
	*x = Event{}
	mi := &file_proto_ingester_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_proto_ingester_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_proto_ingester_proto_rawDescGZIP(), []int{1}
}

func (x *Event) GetJsonPayload() string {
	if x != nil {
		return x.JsonPayload
	}
	return ""
}

var File_proto_ingester_proto protoreflect.FileDescriptor

const file_proto_ingester_proto_rawDesc = "" +
	"\n" +
	"\x14proto/ingester.proto\x12\adstream\"0\n" +
	"\rStreamRequest\x12\x1f\n" +
	"\vconfig_json\x18\x01 \x01(\tR\n" +
	"configJson\"*\n" +
	"\x05Event\x12!\n" +
	"\fjson_payload\x18\x01 \x01(\tR\vjsonPayload2C\n" +
	"\x0eIngesterPlugin\x121\n" +
	"\x05Start\x12\x16.dstream.StreamRequest\x1a\x0e.dstream.Event0\x01B,Z*github.com/katasec/dstream/pkg/proto;protob\x06proto3"

var (
	file_proto_ingester_proto_rawDescOnce sync.Once
	file_proto_ingester_proto_rawDescData []byte
)

func file_proto_ingester_proto_rawDescGZIP() []byte {
	file_proto_ingester_proto_rawDescOnce.Do(func() {
		file_proto_ingester_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_ingester_proto_rawDesc), len(file_proto_ingester_proto_rawDesc)))
	})
	return file_proto_ingester_proto_rawDescData
}

var file_proto_ingester_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_ingester_proto_goTypes = []any{
	(*StreamRequest)(nil), // 0: dstream.StreamRequest
	(*Event)(nil),         // 1: dstream.Event
}
var file_proto_ingester_proto_depIdxs = []int32{
	0, // 0: dstream.IngesterPlugin.Start:input_type -> dstream.StreamRequest
	1, // 1: dstream.IngesterPlugin.Start:output_type -> dstream.Event
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_ingester_proto_init() }
func file_proto_ingester_proto_init() {
	if File_proto_ingester_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_ingester_proto_rawDesc), len(file_proto_ingester_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_ingester_proto_goTypes,
		DependencyIndexes: file_proto_ingester_proto_depIdxs,
		MessageInfos:      file_proto_ingester_proto_msgTypes,
	}.Build()
	File_proto_ingester_proto = out.File
	file_proto_ingester_proto_goTypes = nil
	file_proto_ingester_proto_depIdxs = nil
}
