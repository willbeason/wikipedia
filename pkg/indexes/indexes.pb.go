// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.3
// source: pkg/indexes/indexes.proto

package indexes

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Index allows locating a specific file containing an article with a given ID.
type Index struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Root is the root directory the Index is of.
	Root string `protobuf:"bytes,1,opt,name=root,proto3" json:"root,omitempty"`
	// Entries are the sorted in ascending order by the maximum ID contained in the files.
	Entries []*Entry `protobuf:"bytes,2,rep,name=entries,proto3" json:"entries,omitempty"`
}

func (x *Index) Reset() {
	*x = Index{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_indexes_indexes_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Index) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Index) ProtoMessage() {}

func (x *Index) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_indexes_indexes_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Index.ProtoReflect.Descriptor instead.
func (*Index) Descriptor() ([]byte, []int) {
	return file_pkg_indexes_indexes_proto_rawDescGZIP(), []int{0}
}

func (x *Index) GetRoot() string {
	if x != nil {
		return x.Root
	}
	return ""
}

func (x *Index) GetEntries() []*Entry {
	if x != nil {
		return x.Entries
	}
	return nil
}

type Entry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// File is the file which corresponds to this entry.
	// Relative to the index's root.
	File string `protobuf:"bytes,1,opt,name=file,proto3" json:"file,omitempty"`
	// Max is the maximum page ID contained in File.
	Max uint32 `protobuf:"varint,2,opt,name=max,proto3" json:"max,omitempty"`
}

func (x *Entry) Reset() {
	*x = Entry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_indexes_indexes_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Entry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Entry) ProtoMessage() {}

func (x *Entry) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_indexes_indexes_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Entry.ProtoReflect.Descriptor instead.
func (*Entry) Descriptor() ([]byte, []int) {
	return file_pkg_indexes_indexes_proto_rawDescGZIP(), []int{1}
}

func (x *Entry) GetFile() string {
	if x != nil {
		return x.File
	}
	return ""
}

func (x *Entry) GetMax() uint32 {
	if x != nil {
		return x.Max
	}
	return 0
}

var File_pkg_indexes_indexes_proto protoreflect.FileDescriptor

var file_pkg_indexes_indexes_proto_rawDesc = []byte{
	0x0a, 0x19, 0x70, 0x6b, 0x67, 0x2f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x73, 0x2f, 0x69, 0x6e,
	0x64, 0x65, 0x78, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3d, 0x0a, 0x05, 0x49,
	0x6e, 0x64, 0x65, 0x78, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x72, 0x6f, 0x6f, 0x74, 0x12, 0x20, 0x0a, 0x07, 0x65, 0x6e, 0x74, 0x72,
	0x69, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x06, 0x2e, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x07, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x22, 0x2d, 0x0a, 0x05, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x78, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x6d, 0x61, 0x78, 0x42, 0x0d, 0x5a, 0x0b, 0x70, 0x6b, 0x67,
	0x2f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_indexes_indexes_proto_rawDescOnce sync.Once
	file_pkg_indexes_indexes_proto_rawDescData = file_pkg_indexes_indexes_proto_rawDesc
)

func file_pkg_indexes_indexes_proto_rawDescGZIP() []byte {
	file_pkg_indexes_indexes_proto_rawDescOnce.Do(func() {
		file_pkg_indexes_indexes_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_indexes_indexes_proto_rawDescData)
	})
	return file_pkg_indexes_indexes_proto_rawDescData
}

var file_pkg_indexes_indexes_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pkg_indexes_indexes_proto_goTypes = []interface{}{
	(*Index)(nil), // 0: Index
	(*Entry)(nil), // 1: Entry
}
var file_pkg_indexes_indexes_proto_depIDxs = []int32{
	1, // 0: Index.entries:type_name -> Entry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_pkg_indexes_indexes_proto_init() }
func file_pkg_indexes_indexes_proto_init() {
	if File_pkg_indexes_indexes_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_indexes_indexes_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Index); i {
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
		file_pkg_indexes_indexes_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Entry); i {
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
			RawDescriptor: file_pkg_indexes_indexes_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_indexes_indexes_proto_goTypes,
		DependencyIndexes: file_pkg_indexes_indexes_proto_depIDxs,
		MessageInfos:      file_pkg_indexes_indexes_proto_msgTypes,
	}.Build()
	File_pkg_indexes_indexes_proto = out.File
	file_pkg_indexes_indexes_proto_rawDesc = nil
	file_pkg_indexes_indexes_proto_goTypes = nil
	file_pkg_indexes_indexes_proto_depIDxs = nil
}
