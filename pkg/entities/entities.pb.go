// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: entities/entities.proto

package entities

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

type Entity struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// id is the unique Wikipedia article identifier. Persists across article title changes.
	// Not identical to the wikidata ID.
	Id         uint32               `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	WikidataId string               `protobuf:"bytes,2,opt,name=wikidata_id,json=wikidataId,proto3" json:"wikidata_id,omitempty"`
	Sitelinks  map[string]*SiteLink `protobuf:"bytes,3,rep,name=sitelinks,proto3" json:"sitelinks,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Claims     map[string]*Claims   `protobuf:"bytes,4,rep,name=claims,proto3" json:"claims,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Entity) Reset() {
	*x = Entity{}
	if protoimpl.UnsafeEnabled {
		mi := &file_entities_entities_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Entity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Entity) ProtoMessage() {}

func (x *Entity) ProtoReflect() protoreflect.Message {
	mi := &file_entities_entities_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Entity.ProtoReflect.Descriptor instead.
func (*Entity) Descriptor() ([]byte, []int) {
	return file_entities_entities_proto_rawDescGZIP(), []int{0}
}

func (x *Entity) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Entity) GetWikidataId() string {
	if x != nil {
		return x.WikidataId
	}
	return ""
}

func (x *Entity) GetSitelinks() map[string]*SiteLink {
	if x != nil {
		return x.Sitelinks
	}
	return nil
}

func (x *Entity) GetClaims() map[string]*Claims {
	if x != nil {
		return x.Claims
	}
	return nil
}

type SiteLink struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Site  string `protobuf:"bytes,1,opt,name=site,proto3" json:"site,omitempty"`
	Title string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Url   string `protobuf:"bytes,3,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *SiteLink) Reset() {
	*x = SiteLink{}
	if protoimpl.UnsafeEnabled {
		mi := &file_entities_entities_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SiteLink) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SiteLink) ProtoMessage() {}

func (x *SiteLink) ProtoReflect() protoreflect.Message {
	mi := &file_entities_entities_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SiteLink.ProtoReflect.Descriptor instead.
func (*SiteLink) Descriptor() ([]byte, []int) {
	return file_entities_entities_proto_rawDescGZIP(), []int{1}
}

func (x *SiteLink) GetSite() string {
	if x != nil {
		return x.Site
	}
	return ""
}

func (x *SiteLink) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *SiteLink) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

type Claims struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Claim []*Claim `protobuf:"bytes,1,rep,name=claim,proto3" json:"claim,omitempty"`
}

func (x *Claims) Reset() {
	*x = Claims{}
	if protoimpl.UnsafeEnabled {
		mi := &file_entities_entities_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Claims) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Claims) ProtoMessage() {}

func (x *Claims) ProtoReflect() protoreflect.Message {
	mi := &file_entities_entities_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Claims.ProtoReflect.Descriptor instead.
func (*Claims) Descriptor() ([]byte, []int) {
	return file_entities_entities_proto_rawDescGZIP(), []int{2}
}

func (x *Claims) GetClaim() []*Claim {
	if x != nil {
		return x.Claim
	}
	return nil
}

type Claim struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Property string `protobuf:"bytes,1,opt,name=property,proto3" json:"property,omitempty"`
	Value    string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Rank     string `protobuf:"bytes,3,opt,name=rank,proto3" json:"rank,omitempty"`
}

func (x *Claim) Reset() {
	*x = Claim{}
	if protoimpl.UnsafeEnabled {
		mi := &file_entities_entities_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Claim) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Claim) ProtoMessage() {}

func (x *Claim) ProtoReflect() protoreflect.Message {
	mi := &file_entities_entities_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Claim.ProtoReflect.Descriptor instead.
func (*Claim) Descriptor() ([]byte, []int) {
	return file_entities_entities_proto_rawDescGZIP(), []int{3}
}

func (x *Claim) GetProperty() string {
	if x != nil {
		return x.Property
	}
	return ""
}

func (x *Claim) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Claim) GetRank() string {
	if x != nil {
		return x.Rank
	}
	return ""
}

var File_entities_entities_proto protoreflect.FileDescriptor

var file_entities_entities_proto_rawDesc = []byte{
	0x0a, 0x17, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x2f, 0x65, 0x6e, 0x74, 0x69, 0x74,
	0x69, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa9, 0x02, 0x0a, 0x06, 0x45, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x77, 0x69, 0x6b, 0x69, 0x64, 0x61, 0x74, 0x61,
	0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x77, 0x69, 0x6b, 0x69, 0x64,
	0x61, 0x74, 0x61, 0x49, 0x64, 0x12, 0x34, 0x0a, 0x09, 0x73, 0x69, 0x74, 0x65, 0x6c, 0x69, 0x6e,
	0x6b, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x2e, 0x53, 0x69, 0x74, 0x65, 0x6c, 0x69, 0x6e, 0x6b, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x09, 0x73, 0x69, 0x74, 0x65, 0x6c, 0x69, 0x6e, 0x6b, 0x73, 0x12, 0x2b, 0x0a, 0x06, 0x63,
	0x6c, 0x61, 0x69, 0x6d, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x45, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x2e, 0x43, 0x6c, 0x61, 0x69, 0x6d, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x06, 0x63, 0x6c, 0x61, 0x69, 0x6d, 0x73, 0x1a, 0x47, 0x0a, 0x0e, 0x53, 0x69, 0x74, 0x65,
	0x6c, 0x69, 0x6e, 0x6b, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x1f, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x53, 0x69,
	0x74, 0x65, 0x4c, 0x69, 0x6e, 0x6b, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x1a, 0x42, 0x0a, 0x0b, 0x43, 0x6c, 0x61, 0x69, 0x6d, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x1d, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x07, 0x2e, 0x43, 0x6c, 0x61, 0x69, 0x6d, 0x73, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x46, 0x0a, 0x08, 0x53, 0x69, 0x74, 0x65, 0x4c, 0x69, 0x6e,
	0x6b, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x73, 0x69, 0x74, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x75,
	0x72, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x22, 0x26, 0x0a,
	0x06, 0x43, 0x6c, 0x61, 0x69, 0x6d, 0x73, 0x12, 0x1c, 0x0a, 0x05, 0x63, 0x6c, 0x61, 0x69, 0x6d,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x06, 0x2e, 0x43, 0x6c, 0x61, 0x69, 0x6d, 0x52, 0x05,
	0x63, 0x6c, 0x61, 0x69, 0x6d, 0x22, 0x4d, 0x0a, 0x05, 0x43, 0x6c, 0x61, 0x69, 0x6d, 0x12, 0x1a,
	0x0a, 0x08, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x72, 0x61, 0x6e, 0x6b, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x72, 0x61, 0x6e, 0x6b, 0x42, 0x0e, 0x5a, 0x0c, 0x70, 0x6b, 0x67, 0x2f, 0x65, 0x6e, 0x74, 0x69,
	0x74, 0x69, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_entities_entities_proto_rawDescOnce sync.Once
	file_entities_entities_proto_rawDescData = file_entities_entities_proto_rawDesc
)

func file_entities_entities_proto_rawDescGZIP() []byte {
	file_entities_entities_proto_rawDescOnce.Do(func() {
		file_entities_entities_proto_rawDescData = protoimpl.X.CompressGZIP(file_entities_entities_proto_rawDescData)
	})
	return file_entities_entities_proto_rawDescData
}

var file_entities_entities_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_entities_entities_proto_goTypes = []interface{}{
	(*Entity)(nil),   // 0: Entity
	(*SiteLink)(nil), // 1: SiteLink
	(*Claims)(nil),   // 2: Claims
	(*Claim)(nil),    // 3: Claim
	nil,              // 4: Entity.SitelinksEntry
	nil,              // 5: Entity.ClaimsEntry
}
var file_entities_entities_proto_depIdxs = []int32{
	4, // 0: Entity.sitelinks:type_name -> Entity.SitelinksEntry
	5, // 1: Entity.claims:type_name -> Entity.ClaimsEntry
	3, // 2: Claims.claim:type_name -> Claim
	1, // 3: Entity.SitelinksEntry.value:type_name -> SiteLink
	2, // 4: Entity.ClaimsEntry.value:type_name -> Claims
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_entities_entities_proto_init() }
func file_entities_entities_proto_init() {
	if File_entities_entities_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_entities_entities_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Entity); i {
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
		file_entities_entities_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SiteLink); i {
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
		file_entities_entities_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Claims); i {
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
		file_entities_entities_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Claim); i {
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
			RawDescriptor: file_entities_entities_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_entities_entities_proto_goTypes,
		DependencyIndexes: file_entities_entities_proto_depIdxs,
		MessageInfos:      file_entities_entities_proto_msgTypes,
	}.Build()
	File_entities_entities_proto = out.File
	file_entities_entities_proto_rawDesc = nil
	file_entities_entities_proto_goTypes = nil
	file_entities_entities_proto_depIdxs = nil
}