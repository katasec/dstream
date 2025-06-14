// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: proto/plugin.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	structpb "google.golang.org/protobuf/types/known/structpb"
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

// ──────────────────────────────────────────────────────────────────────────────
//
//	Field-schema definition
//
// ──────────────────────────────────────────────────────────────────────────────
type FieldSchema struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"` // e.g. "lock", "provider"
	Type          string                 `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"` // "string", "list", "object", …
	Required      bool                   `protobuf:"varint,3,opt,name=required,proto3" json:"required,omitempty"`
	Description   string                 `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Fields        []*FieldSchema         `protobuf:"bytes,5,rep,name=fields,proto3" json:"fields,omitempty"` // populated when type == "object"
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FieldSchema) Reset() {
	*x = FieldSchema{}
	mi := &file_proto_plugin_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FieldSchema) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldSchema) ProtoMessage() {}

func (x *FieldSchema) ProtoReflect() protoreflect.Message {
	mi := &file_proto_plugin_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldSchema.ProtoReflect.Descriptor instead.
func (*FieldSchema) Descriptor() ([]byte, []int) {
	return file_proto_plugin_proto_rawDescGZIP(), []int{0}
}

func (x *FieldSchema) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FieldSchema) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *FieldSchema) GetRequired() bool {
	if x != nil {
		return x.Required
	}
	return false
}

func (x *FieldSchema) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *FieldSchema) GetFields() []*FieldSchema {
	if x != nil {
		return x.Fields
	}
	return nil
}

type GetSchemaResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Fields        []*FieldSchema         `protobuf:"bytes,1,rep,name=fields,proto3" json:"fields,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetSchemaResponse) Reset() {
	*x = GetSchemaResponse{}
	mi := &file_proto_plugin_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetSchemaResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetSchemaResponse) ProtoMessage() {}

func (x *GetSchemaResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_plugin_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetSchemaResponse.ProtoReflect.Descriptor instead.
func (*GetSchemaResponse) Descriptor() ([]byte, []int) {
	return file_proto_plugin_proto_rawDescGZIP(), []int{1}
}

func (x *GetSchemaResponse) GetFields() []*FieldSchema {
	if x != nil {
		return x.Fields
	}
	return nil
}

var File_proto_plugin_proto protoreflect.FileDescriptor

const file_proto_plugin_proto_rawDesc = "" +
	"\n" +
	"\x12proto/plugin.proto\x12\x06plugin\x1a\x1bgoogle/protobuf/empty.proto\x1a\x1cgoogle/protobuf/struct.proto\"\xa0\x01\n" +
	"\vFieldSchema\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x12\n" +
	"\x04type\x18\x02 \x01(\tR\x04type\x12\x1a\n" +
	"\brequired\x18\x03 \x01(\bR\brequired\x12 \n" +
	"\vdescription\x18\x04 \x01(\tR\vdescription\x12+\n" +
	"\x06fields\x18\x05 \x03(\v2\x13.plugin.FieldSchemaR\x06fields\"@\n" +
	"\x11GetSchemaResponse\x12+\n" +
	"\x06fields\x18\x01 \x03(\v2\x13.plugin.FieldSchemaR\x06fields2\x82\x01\n" +
	"\x06Plugin\x12>\n" +
	"\tGetSchema\x12\x16.google.protobuf.Empty\x1a\x19.plugin.GetSchemaResponse\x128\n" +
	"\x05Start\x12\x17.google.protobuf.Struct\x1a\x16.google.protobuf.EmptyB(Z&github.com/katasec/dstream/proto;protob\x06proto3"

var (
	file_proto_plugin_proto_rawDescOnce sync.Once
	file_proto_plugin_proto_rawDescData []byte
)

func file_proto_plugin_proto_rawDescGZIP() []byte {
	file_proto_plugin_proto_rawDescOnce.Do(func() {
		file_proto_plugin_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_plugin_proto_rawDesc), len(file_proto_plugin_proto_rawDesc)))
	})
	return file_proto_plugin_proto_rawDescData
}

var file_proto_plugin_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_plugin_proto_goTypes = []any{
	(*FieldSchema)(nil),       // 0: plugin.FieldSchema
	(*GetSchemaResponse)(nil), // 1: plugin.GetSchemaResponse
	(*emptypb.Empty)(nil),     // 2: google.protobuf.Empty
	(*structpb.Struct)(nil),   // 3: google.protobuf.Struct
}
var file_proto_plugin_proto_depIdxs = []int32{
	0, // 0: plugin.FieldSchema.fields:type_name -> plugin.FieldSchema
	0, // 1: plugin.GetSchemaResponse.fields:type_name -> plugin.FieldSchema
	2, // 2: plugin.Plugin.GetSchema:input_type -> google.protobuf.Empty
	3, // 3: plugin.Plugin.Start:input_type -> google.protobuf.Struct
	1, // 4: plugin.Plugin.GetSchema:output_type -> plugin.GetSchemaResponse
	2, // 5: plugin.Plugin.Start:output_type -> google.protobuf.Empty
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_proto_plugin_proto_init() }
func file_proto_plugin_proto_init() {
	if File_proto_plugin_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_plugin_proto_rawDesc), len(file_proto_plugin_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_plugin_proto_goTypes,
		DependencyIndexes: file_proto_plugin_proto_depIdxs,
		MessageInfos:      file_proto_plugin_proto_msgTypes,
	}.Build()
	File_proto_plugin_proto = out.File
	file_proto_plugin_proto_goTypes = nil
	file_proto_plugin_proto_depIdxs = nil
}
