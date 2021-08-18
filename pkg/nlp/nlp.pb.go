// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.3
// source: pkg/nlp/nlp.proto

package nlp

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

// FrequencyMap is a set of known words and their frequencies.
type FrequencyMap struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Words map[string]uint32 `protobuf:"bytes,1,rep,name=words,proto3" json:"words,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
}

func (x *FrequencyMap) Reset() {
	*x = FrequencyMap{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_nlp_nlp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FrequencyMap) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FrequencyMap) ProtoMessage() {}

func (x *FrequencyMap) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_nlp_nlp_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FrequencyMap.ProtoReflect.Descriptor instead.
func (*FrequencyMap) Descriptor() ([]byte, []int) {
	return file_pkg_nlp_nlp_proto_rawDescGZIP(), []int{0}
}

func (x *FrequencyMap) GetWords() map[string]uint32 {
	if x != nil {
		return x.Words
	}
	return nil
}

// FrequencyMap is a set of known words and their frequencies.
type FrequencyTable struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Words []*WordCount `protobuf:"bytes,1,rep,name=words,proto3" json:"words,omitempty"`
}

func (x *FrequencyTable) Reset() {
	*x = FrequencyTable{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_nlp_nlp_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FrequencyTable) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FrequencyTable) ProtoMessage() {}

func (x *FrequencyTable) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_nlp_nlp_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FrequencyTable.ProtoReflect.Descriptor instead.
func (*FrequencyTable) Descriptor() ([]byte, []int) {
	return file_pkg_nlp_nlp_proto_rawDescGZIP(), []int{1}
}

func (x *FrequencyTable) GetWords() []*WordCount {
	if x != nil {
		return x.Words
	}
	return nil
}

type WordCount struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Word  string `protobuf:"bytes,1,opt,name=word,proto3" json:"word,omitempty"`
	Count uint32 `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *WordCount) Reset() {
	*x = WordCount{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_nlp_nlp_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WordCount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WordCount) ProtoMessage() {}

func (x *WordCount) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_nlp_nlp_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WordCount.ProtoReflect.Descriptor instead.
func (*WordCount) Descriptor() ([]byte, []int) {
	return file_pkg_nlp_nlp_proto_rawDescGZIP(), []int{2}
}

func (x *WordCount) GetWord() string {
	if x != nil {
		return x.Word
	}
	return ""
}

func (x *WordCount) GetCount() uint32 {
	if x != nil {
		return x.Count
	}
	return 0
}

// Dictionary is a set of known words.
type Dictionary struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Words is a list of recognized words, in the order they appear in a frequency table.
	Words []string `protobuf:"bytes,1,rep,name=words,proto3" json:"words,omitempty"`
}

func (x *Dictionary) Reset() {
	*x = Dictionary{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_nlp_nlp_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Dictionary) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Dictionary) ProtoMessage() {}

func (x *Dictionary) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_nlp_nlp_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Dictionary.ProtoReflect.Descriptor instead.
func (*Dictionary) Descriptor() ([]byte, []int) {
	return file_pkg_nlp_nlp_proto_rawDescGZIP(), []int{3}
}

func (x *Dictionary) GetWords() []string {
	if x != nil {
		return x.Words
	}
	return nil
}

var File_pkg_nlp_nlp_proto protoreflect.FileDescriptor

var file_pkg_nlp_nlp_proto_rawDesc = []byte{
	0x0a, 0x11, 0x70, 0x6b, 0x67, 0x2f, 0x6e, 0x6c, 0x70, 0x2f, 0x6e, 0x6c, 0x70, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x03, 0x6e, 0x6c, 0x70, 0x22, 0x7c, 0x0a, 0x0c, 0x46, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x6e, 0x63, 0x79, 0x4d, 0x61, 0x70, 0x12, 0x32, 0x0a, 0x05, 0x77, 0x6f, 0x72, 0x64,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6e, 0x6c, 0x70, 0x2e, 0x46, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x79, 0x4d, 0x61, 0x70, 0x2e, 0x57, 0x6f, 0x72, 0x64, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x77, 0x6f, 0x72, 0x64, 0x73, 0x1a, 0x38, 0x0a, 0x0a,
	0x57, 0x6f, 0x72, 0x64, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x36, 0x0a, 0x0e, 0x46, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x6e, 0x63, 0x79, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x24, 0x0a, 0x05, 0x77, 0x6f, 0x72, 0x64,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x6e, 0x6c, 0x70, 0x2e, 0x57, 0x6f,
	0x72, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x05, 0x77, 0x6f, 0x72, 0x64, 0x73, 0x22, 0x35,
	0x0a, 0x09, 0x57, 0x6f, 0x72, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x77,
	0x6f, 0x72, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x77, 0x6f, 0x72, 0x64, 0x12,
	0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x22, 0x0a, 0x0a, 0x44, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x61, 0x72, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x77, 0x6f, 0x72, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x05, 0x77, 0x6f, 0x72, 0x64, 0x73, 0x42, 0x09, 0x5a, 0x07, 0x70, 0x6b, 0x67,
	0x2f, 0x6e, 0x6c, 0x70, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_nlp_nlp_proto_rawDescOnce sync.Once
	file_pkg_nlp_nlp_proto_rawDescData = file_pkg_nlp_nlp_proto_rawDesc
)

func file_pkg_nlp_nlp_proto_rawDescGZIP() []byte {
	file_pkg_nlp_nlp_proto_rawDescOnce.Do(func() {
		file_pkg_nlp_nlp_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_nlp_nlp_proto_rawDescData)
	})
	return file_pkg_nlp_nlp_proto_rawDescData
}

var file_pkg_nlp_nlp_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_pkg_nlp_nlp_proto_goTypes = []interface{}{
	(*FrequencyMap)(nil),   // 0: nlp.FrequencyMap
	(*FrequencyTable)(nil), // 1: nlp.FrequencyTable
	(*WordCount)(nil),      // 2: nlp.WordCount
	(*Dictionary)(nil),     // 3: nlp.Dictionary
	nil,                    // 4: nlp.FrequencyMap.WordsEntry
}
var file_pkg_nlp_nlp_proto_depIdxs = []int32{
	4, // 0: nlp.FrequencyMap.words:type_name -> nlp.FrequencyMap.WordsEntry
	2, // 1: nlp.FrequencyTable.words:type_name -> nlp.WordCount
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_pkg_nlp_nlp_proto_init() }
func file_pkg_nlp_nlp_proto_init() {
	if File_pkg_nlp_nlp_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_nlp_nlp_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FrequencyMap); i {
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
		file_pkg_nlp_nlp_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FrequencyTable); i {
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
		file_pkg_nlp_nlp_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WordCount); i {
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
		file_pkg_nlp_nlp_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Dictionary); i {
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
			RawDescriptor: file_pkg_nlp_nlp_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_nlp_nlp_proto_goTypes,
		DependencyIndexes: file_pkg_nlp_nlp_proto_depIdxs,
		MessageInfos:      file_pkg_nlp_nlp_proto_msgTypes,
	}.Build()
	File_pkg_nlp_nlp_proto = out.File
	file_pkg_nlp_nlp_proto_rawDesc = nil
	file_pkg_nlp_nlp_proto_goTypes = nil
	file_pkg_nlp_nlp_proto_depIdxs = nil
}
