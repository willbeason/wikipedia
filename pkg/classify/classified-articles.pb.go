// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.3
// source: pkg/classify/classified-articles.proto

package classify

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

type ClassifiedArticles struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Articles []*ClassifiedArticle `protobuf:"bytes,1,rep,name=articles,proto3" json:"articles,omitempty"`
}

func (x *ClassifiedArticles) Reset() {
	*x = ClassifiedArticles{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_classify_classified_articles_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClassifiedArticles) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClassifiedArticles) ProtoMessage() {}

func (x *ClassifiedArticles) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_classify_classified_articles_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClassifiedArticles.ProtoReflect.Descriptor instead.
func (*ClassifiedArticles) Descriptor() ([]byte, []int) {
	return file_pkg_classify_classified_articles_proto_rawDescGZIP(), []int{0}
}

func (x *ClassifiedArticles) GetArticles() []*ClassifiedArticle {
	if x != nil {
		return x.Articles
	}
	return nil
}

type ClassifiedArticle struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id             int32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Classification int32 `protobuf:"varint,2,opt,name=classification,proto3" json:"classification,omitempty"`
}

func (x *ClassifiedArticle) Reset() {
	*x = ClassifiedArticle{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_classify_classified_articles_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClassifiedArticle) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClassifiedArticle) ProtoMessage() {}

func (x *ClassifiedArticle) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_classify_classified_articles_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClassifiedArticle.ProtoReflect.Descriptor instead.
func (*ClassifiedArticle) Descriptor() ([]byte, []int) {
	return file_pkg_classify_classified_articles_proto_rawDescGZIP(), []int{1}
}

func (x *ClassifiedArticle) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *ClassifiedArticle) GetClassification() int32 {
	if x != nil {
		return x.Classification
	}
	return 0
}

var File_pkg_classify_classified_articles_proto protoreflect.FileDescriptor

var file_pkg_classify_classified_articles_proto_rawDesc = []byte{
	0x0a, 0x26, 0x70, 0x6b, 0x67, 0x2f, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x79, 0x2f, 0x63,
	0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69, 0x65, 0x64, 0x2d, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c,
	0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x44, 0x0a, 0x12, 0x43, 0x6c, 0x61, 0x73,
	0x73, 0x69, 0x66, 0x69, 0x65, 0x64, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x12, 0x2e,
	0x0a, 0x08, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x12, 0x2e, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69, 0x65, 0x64, 0x41, 0x72, 0x74,
	0x69, 0x63, 0x6c, 0x65, 0x52, 0x08, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x22, 0x4b,
	0x0a, 0x11, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69, 0x65, 0x64, 0x41, 0x72, 0x74, 0x69,
	0x63, 0x6c, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x26, 0x0a, 0x0e, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x63, 0x6c, 0x61,
	0x73, 0x73, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x2e, 0x5a, 0x2c, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x69, 0x6c, 0x6c, 0x62, 0x65,
	0x61, 0x73, 0x6f, 0x6e, 0x2f, 0x77, 0x69, 0x6b, 0x69, 0x70, 0x65, 0x64, 0x69, 0x61, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x79, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_pkg_classify_classified_articles_proto_rawDescOnce sync.Once
	file_pkg_classify_classified_articles_proto_rawDescData = file_pkg_classify_classified_articles_proto_rawDesc
)

func file_pkg_classify_classified_articles_proto_rawDescGZIP() []byte {
	file_pkg_classify_classified_articles_proto_rawDescOnce.Do(func() {
		file_pkg_classify_classified_articles_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_classify_classified_articles_proto_rawDescData)
	})
	return file_pkg_classify_classified_articles_proto_rawDescData
}

var file_pkg_classify_classified_articles_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pkg_classify_classified_articles_proto_goTypes = []interface{}{
	(*ClassifiedArticles)(nil), // 0: ClassifiedArticles
	(*ClassifiedArticle)(nil),  // 1: ClassifiedArticle
}
var file_pkg_classify_classified_articles_proto_depIdxs = []int32{
	1, // 0: ClassifiedArticles.articles:type_name -> ClassifiedArticle
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_pkg_classify_classified_articles_proto_init() }
func file_pkg_classify_classified_articles_proto_init() {
	if File_pkg_classify_classified_articles_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_classify_classified_articles_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClassifiedArticles); i {
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
		file_pkg_classify_classified_articles_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClassifiedArticle); i {
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
			RawDescriptor: file_pkg_classify_classified_articles_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_classify_classified_articles_proto_goTypes,
		DependencyIndexes: file_pkg_classify_classified_articles_proto_depIdxs,
		MessageInfos:      file_pkg_classify_classified_articles_proto_msgTypes,
	}.Build()
	File_pkg_classify_classified_articles_proto = out.File
	file_pkg_classify_classified_articles_proto_rawDesc = nil
	file_pkg_classify_classified_articles_proto_goTypes = nil
	file_pkg_classify_classified_articles_proto_depIdxs = nil
}
