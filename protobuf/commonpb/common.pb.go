// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.21.12
// source: commonpb/common.proto

package commonpb

import (
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

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_commonpb_common_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_commonpb_common_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_commonpb_common_proto_rawDescGZIP(), []int{0}
}

type Addr struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ip   string `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	Port uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
}

func (x *Addr) Reset() {
	*x = Addr{}
	if protoimpl.UnsafeEnabled {
		mi := &file_commonpb_common_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Addr) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Addr) ProtoMessage() {}

func (x *Addr) ProtoReflect() protoreflect.Message {
	mi := &file_commonpb_common_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Addr.ProtoReflect.Descriptor instead.
func (*Addr) Descriptor() ([]byte, []int) {
	return file_commonpb_common_proto_rawDescGZIP(), []int{1}
}

func (x *Addr) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *Addr) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

type AddrPack struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Local  *Addr `protobuf:"bytes,1,opt,name=local,proto3" json:"local,omitempty"`
	Remote *Addr `protobuf:"bytes,2,opt,name=remote,proto3" json:"remote,omitempty"`
}

func (x *AddrPack) Reset() {
	*x = AddrPack{}
	if protoimpl.UnsafeEnabled {
		mi := &file_commonpb_common_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddrPack) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddrPack) ProtoMessage() {}

func (x *AddrPack) ProtoReflect() protoreflect.Message {
	mi := &file_commonpb_common_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddrPack.ProtoReflect.Descriptor instead.
func (*AddrPack) Descriptor() ([]byte, []int) {
	return file_commonpb_common_proto_rawDescGZIP(), []int{2}
}

func (x *AddrPack) GetLocal() *Addr {
	if x != nil {
		return x.Local
	}
	return nil
}

func (x *AddrPack) GetRemote() *Addr {
	if x != nil {
		return x.Remote
	}
	return nil
}

var File_commonpb_common_proto protoreflect.FileDescriptor

var file_commonpb_common_proto_rawDesc = []byte{
	0x0a, 0x15, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x70, 0x62, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x70,
	0x62, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x2a, 0x0a, 0x04, 0x41, 0x64,
	0x64, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x58, 0x0a, 0x08, 0x41, 0x64, 0x64, 0x72, 0x50, 0x61,
	0x63, 0x6b, 0x12, 0x24, 0x0a, 0x05, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0e, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x70, 0x62, 0x2e, 0x41, 0x64, 0x64,
	0x72, 0x52, 0x05, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x12, 0x26, 0x0a, 0x06, 0x72, 0x65, 0x6d, 0x6f,
	0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x70, 0x62, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x52, 0x06, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65,
	0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x56,
	0x65, 0x69, 0x64, 0x73, 0x2f, 0x67, 0x72, 0x64, 0x70, 0x32, 0x74, 0x63, 0x70, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_commonpb_common_proto_rawDescOnce sync.Once
	file_commonpb_common_proto_rawDescData = file_commonpb_common_proto_rawDesc
)

func file_commonpb_common_proto_rawDescGZIP() []byte {
	file_commonpb_common_proto_rawDescOnce.Do(func() {
		file_commonpb_common_proto_rawDescData = protoimpl.X.CompressGZIP(file_commonpb_common_proto_rawDescData)
	})
	return file_commonpb_common_proto_rawDescData
}

var file_commonpb_common_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_commonpb_common_proto_goTypes = []interface{}{
	(*Empty)(nil),    // 0: commonpb.Empty
	(*Addr)(nil),     // 1: commonpb.Addr
	(*AddrPack)(nil), // 2: commonpb.AddrPack
}
var file_commonpb_common_proto_depIdxs = []int32{
	1, // 0: commonpb.AddrPack.local:type_name -> commonpb.Addr
	1, // 1: commonpb.AddrPack.remote:type_name -> commonpb.Addr
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_commonpb_common_proto_init() }
func file_commonpb_common_proto_init() {
	if File_commonpb_common_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_commonpb_common_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
		file_commonpb_common_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Addr); i {
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
		file_commonpb_common_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddrPack); i {
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
			RawDescriptor: file_commonpb_common_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_commonpb_common_proto_goTypes,
		DependencyIndexes: file_commonpb_common_proto_depIdxs,
		MessageInfos:      file_commonpb_common_proto_msgTypes,
	}.Build()
	File_commonpb_common_proto = out.File
	file_commonpb_common_proto_rawDesc = nil
	file_commonpb_common_proto_goTypes = nil
	file_commonpb_common_proto_depIdxs = nil
}
