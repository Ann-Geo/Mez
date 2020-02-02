// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/edgenode/edgenode_api.proto

package edgenode

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Url struct {
	Address              string   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Url) Reset()         { *m = Url{} }
func (m *Url) String() string { return proto.CompactTextString(m) }
func (*Url) ProtoMessage()    {}
func (*Url) Descriptor() ([]byte, []int) {
	return fileDescriptor_197815cccd0f3d4f, []int{0}
}

func (m *Url) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Url.Unmarshal(m, b)
}
func (m *Url) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Url.Marshal(b, m, deterministic)
}
func (m *Url) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Url.Merge(m, src)
}
func (m *Url) XXX_Size() int {
	return xxx_messageInfo_Url.Size(m)
}
func (m *Url) XXX_DiscardUnknown() {
	xxx_messageInfo_Url.DiscardUnknown(m)
}

var xxx_messageInfo_Url proto.InternalMessageInfo

func (m *Url) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type Id struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Id) Reset()         { *m = Id{} }
func (m *Id) String() string { return proto.CompactTextString(m) }
func (*Id) ProtoMessage()    {}
func (*Id) Descriptor() ([]byte, []int) {
	return fileDescriptor_197815cccd0f3d4f, []int{1}
}

func (m *Id) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Id.Unmarshal(m, b)
}
func (m *Id) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Id.Marshal(b, m, deterministic)
}
func (m *Id) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Id.Merge(m, src)
}
func (m *Id) XXX_Size() int {
	return xxx_messageInfo_Id.Size(m)
}
func (m *Id) XXX_DiscardUnknown() {
	xxx_messageInfo_Id.DiscardUnknown(m)
}

var xxx_messageInfo_Id proto.InternalMessageInfo

func (m *Id) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type Image struct {
	Image                []byte   `protobuf:"bytes,1,opt,name=image,proto3" json:"image,omitempty"`
	Timestamp            string   `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Image) Reset()         { *m = Image{} }
func (m *Image) String() string { return proto.CompactTextString(m) }
func (*Image) ProtoMessage()    {}
func (*Image) Descriptor() ([]byte, []int) {
	return fileDescriptor_197815cccd0f3d4f, []int{2}
}

func (m *Image) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Image.Unmarshal(m, b)
}
func (m *Image) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Image.Marshal(b, m, deterministic)
}
func (m *Image) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Image.Merge(m, src)
}
func (m *Image) XXX_Size() int {
	return xxx_messageInfo_Image.Size(m)
}
func (m *Image) XXX_DiscardUnknown() {
	xxx_messageInfo_Image.DiscardUnknown(m)
}

var xxx_messageInfo_Image proto.InternalMessageInfo

func (m *Image) GetImage() []byte {
	if m != nil {
		return m.Image
	}
	return nil
}

func (m *Image) GetTimestamp() string {
	if m != nil {
		return m.Timestamp
	}
	return ""
}

type CameraInfo struct {
	Camid                []string `protobuf:"bytes,1,rep,name=camid,proto3" json:"camid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CameraInfo) Reset()         { *m = CameraInfo{} }
func (m *CameraInfo) String() string { return proto.CompactTextString(m) }
func (*CameraInfo) ProtoMessage()    {}
func (*CameraInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_197815cccd0f3d4f, []int{3}
}

func (m *CameraInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CameraInfo.Unmarshal(m, b)
}
func (m *CameraInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CameraInfo.Marshal(b, m, deterministic)
}
func (m *CameraInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CameraInfo.Merge(m, src)
}
func (m *CameraInfo) XXX_Size() int {
	return xxx_messageInfo_CameraInfo.Size(m)
}
func (m *CameraInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_CameraInfo.DiscardUnknown(m)
}

var xxx_messageInfo_CameraInfo proto.InternalMessageInfo

func (m *CameraInfo) GetCamid() []string {
	if m != nil {
		return m.Camid
	}
	return nil
}

type ImageStreamParameters struct {
	Camid                string   `protobuf:"bytes,1,opt,name=camid,proto3" json:"camid,omitempty"`
	Latency              string   `protobuf:"bytes,2,opt,name=latency,proto3" json:"latency,omitempty"`
	Accuracy             string   `protobuf:"bytes,3,opt,name=accuracy,proto3" json:"accuracy,omitempty"`
	Start                string   `protobuf:"bytes,4,opt,name=start,proto3" json:"start,omitempty"`
	Stop                 string   `protobuf:"bytes,5,opt,name=stop,proto3" json:"stop,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ImageStreamParameters) Reset()         { *m = ImageStreamParameters{} }
func (m *ImageStreamParameters) String() string { return proto.CompactTextString(m) }
func (*ImageStreamParameters) ProtoMessage()    {}
func (*ImageStreamParameters) Descriptor() ([]byte, []int) {
	return fileDescriptor_197815cccd0f3d4f, []int{4}
}

func (m *ImageStreamParameters) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ImageStreamParameters.Unmarshal(m, b)
}
func (m *ImageStreamParameters) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ImageStreamParameters.Marshal(b, m, deterministic)
}
func (m *ImageStreamParameters) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ImageStreamParameters.Merge(m, src)
}
func (m *ImageStreamParameters) XXX_Size() int {
	return xxx_messageInfo_ImageStreamParameters.Size(m)
}
func (m *ImageStreamParameters) XXX_DiscardUnknown() {
	xxx_messageInfo_ImageStreamParameters.DiscardUnknown(m)
}

var xxx_messageInfo_ImageStreamParameters proto.InternalMessageInfo

func (m *ImageStreamParameters) GetCamid() string {
	if m != nil {
		return m.Camid
	}
	return ""
}

func (m *ImageStreamParameters) GetLatency() string {
	if m != nil {
		return m.Latency
	}
	return ""
}

func (m *ImageStreamParameters) GetAccuracy() string {
	if m != nil {
		return m.Accuracy
	}
	return ""
}

func (m *ImageStreamParameters) GetStart() string {
	if m != nil {
		return m.Start
	}
	return ""
}

func (m *ImageStreamParameters) GetStop() string {
	if m != nil {
		return m.Stop
	}
	return ""
}

type Status struct {
	Status               bool     `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Status) Reset()         { *m = Status{} }
func (m *Status) String() string { return proto.CompactTextString(m) }
func (*Status) ProtoMessage()    {}
func (*Status) Descriptor() ([]byte, []int) {
	return fileDescriptor_197815cccd0f3d4f, []int{5}
}

func (m *Status) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Status.Unmarshal(m, b)
}
func (m *Status) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Status.Marshal(b, m, deterministic)
}
func (m *Status) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Status.Merge(m, src)
}
func (m *Status) XXX_Size() int {
	return xxx_messageInfo_Status.Size(m)
}
func (m *Status) XXX_DiscardUnknown() {
	xxx_messageInfo_Status.DiscardUnknown(m)
}

var xxx_messageInfo_Status proto.InternalMessageInfo

func (m *Status) GetStatus() bool {
	if m != nil {
		return m.Status
	}
	return false
}

type LatencyMeasured struct {
	CurrentLat           string   `protobuf:"bytes,1,opt,name=current_lat,json=currentLat,proto3" json:"current_lat,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LatencyMeasured) Reset()         { *m = LatencyMeasured{} }
func (m *LatencyMeasured) String() string { return proto.CompactTextString(m) }
func (*LatencyMeasured) ProtoMessage()    {}
func (*LatencyMeasured) Descriptor() ([]byte, []int) {
	return fileDescriptor_197815cccd0f3d4f, []int{6}
}

func (m *LatencyMeasured) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LatencyMeasured.Unmarshal(m, b)
}
func (m *LatencyMeasured) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LatencyMeasured.Marshal(b, m, deterministic)
}
func (m *LatencyMeasured) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LatencyMeasured.Merge(m, src)
}
func (m *LatencyMeasured) XXX_Size() int {
	return xxx_messageInfo_LatencyMeasured.Size(m)
}
func (m *LatencyMeasured) XXX_DiscardUnknown() {
	xxx_messageInfo_LatencyMeasured.DiscardUnknown(m)
}

var xxx_messageInfo_LatencyMeasured proto.InternalMessageInfo

func (m *LatencyMeasured) GetCurrentLat() string {
	if m != nil {
		return m.CurrentLat
	}
	return ""
}

func init() {
	proto.RegisterType((*Url)(nil), "edgenode.Url")
	proto.RegisterType((*Id)(nil), "edgenode.Id")
	proto.RegisterType((*Image)(nil), "edgenode.Image")
	proto.RegisterType((*CameraInfo)(nil), "edgenode.CameraInfo")
	proto.RegisterType((*ImageStreamParameters)(nil), "edgenode.ImageStreamParameters")
	proto.RegisterType((*Status)(nil), "edgenode.Status")
	proto.RegisterType((*LatencyMeasured)(nil), "edgenode.LatencyMeasured")
}

func init() { proto.RegisterFile("api/edgenode/edgenode_api.proto", fileDescriptor_197815cccd0f3d4f) }

var fileDescriptor_197815cccd0f3d4f = []byte{
	// 404 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x92, 0xcd, 0x6a, 0xdb, 0x40,
	0x10, 0xc7, 0x2d, 0x39, 0xfe, 0x1a, 0xa7, 0x4d, 0x19, 0xdc, 0xa2, 0x9a, 0x82, 0xc3, 0x9e, 0x7c,
	0x72, 0x43, 0x7a, 0xe8, 0xa1, 0xbd, 0x14, 0x9f, 0x0c, 0x29, 0x18, 0x1b, 0x9f, 0xc3, 0x68, 0x77,
	0x9a, 0x2e, 0xe8, 0x8b, 0xfd, 0x38, 0xe4, 0x21, 0xfa, 0x0a, 0x7d, 0xd6, 0xa0, 0x95, 0x6c, 0x05,
	0x91, 0xdb, 0xfc, 0xfe, 0x33, 0xf3, 0x9f, 0xd1, 0xac, 0x60, 0x45, 0x95, 0xfe, 0xca, 0xea, 0x89,
	0x8b, 0x52, 0xf1, 0x25, 0x78, 0xa4, 0x4a, 0x6f, 0x2a, 0x53, 0xba, 0x12, 0xa7, 0x67, 0x4d, 0xac,
	0x60, 0x78, 0x32, 0x19, 0x26, 0x30, 0x21, 0xa5, 0x0c, 0x5b, 0x9b, 0x44, 0xb7, 0xd1, 0x7a, 0x76,
	0x38, 0xa3, 0x58, 0x40, 0xbc, 0x53, 0xf8, 0x1e, 0x62, 0xad, 0xda, 0x54, 0xac, 0x95, 0xf8, 0x01,
	0xa3, 0x5d, 0x4e, 0x4f, 0x8c, 0x0b, 0x18, 0xe9, 0x3a, 0x08, 0xb9, 0xeb, 0x43, 0x03, 0xf8, 0x05,
	0x66, 0x4e, 0xe7, 0x6c, 0x1d, 0xe5, 0x55, 0x12, 0x87, 0xae, 0x4e, 0x10, 0x02, 0x60, 0x4b, 0x39,
	0x1b, 0xda, 0x15, 0x7f, 0xca, 0xda, 0x41, 0x52, 0x1e, 0xdc, 0x87, 0xeb, 0xd9, 0xa1, 0x01, 0xf1,
	0x2f, 0x82, 0x8f, 0x61, 0xc2, 0xd1, 0x19, 0xa6, 0x7c, 0x4f, 0x86, 0x72, 0x76, 0x6c, 0xec, 0xeb,
	0xfa, 0xe8, 0x52, 0x5f, 0x7f, 0x40, 0x46, 0x8e, 0x0b, 0xf9, 0xdc, 0xce, 0x3b, 0x23, 0x2e, 0x61,
	0x4a, 0x52, 0x7a, 0x43, 0xf2, 0x39, 0x19, 0x86, 0xd4, 0x85, 0x6b, 0x2f, 0xeb, 0xc8, 0xb8, 0xe4,
	0xaa, 0xf1, 0x0a, 0x80, 0x08, 0x57, 0xd6, 0x95, 0x55, 0x32, 0x0a, 0x62, 0x88, 0xc5, 0x2d, 0x8c,
	0x8f, 0x8e, 0x9c, 0xb7, 0xf8, 0x09, 0xc6, 0x36, 0x44, 0x61, 0x81, 0xe9, 0xa1, 0x25, 0x71, 0x0f,
	0x37, 0x0f, 0xcd, 0xc8, 0xdf, 0x4c, 0xd6, 0x1b, 0x56, 0xb8, 0x82, 0xb9, 0xf4, 0xc6, 0x70, 0xe1,
	0x1e, 0x33, 0x72, 0xed, 0xc2, 0xd0, 0x4a, 0x0f, 0xe4, 0xee, 0xff, 0xc7, 0x30, 0xde, 0xfb, 0xf4,
	0xe8, 0x53, 0x5c, 0xc3, 0x64, 0x5b, 0x16, 0x05, 0x4b, 0x87, 0xef, 0x36, 0xe7, 0xe7, 0xd9, 0x9c,
	0x4c, 0xb6, 0xbc, 0xee, 0x70, 0xa7, 0xc4, 0x00, 0xef, 0x60, 0xb2, 0xf7, 0x69, 0xa6, 0xed, 0x5f,
	0xbc, 0x79, 0x95, 0xaa, 0x8f, 0xb5, 0xfc, 0xd0, 0x09, 0xcd, 0xba, 0x62, 0xb0, 0x8e, 0xf0, 0x17,
	0xcc, 0x8e, 0x3e, 0xb5, 0xd2, 0xe8, 0x94, 0x71, 0xd5, 0xeb, 0xe9, 0x1f, 0x78, 0xd9, 0x37, 0x15,
	0x83, 0xbb, 0x08, 0xbf, 0xc3, 0xfc, 0x54, 0xd8, 0x8b, 0xc9, 0xa2, 0xab, 0xe9, 0x9e, 0xf2, 0xad,
	0xe9, 0xf8, 0x13, 0xe6, 0xed, 0x59, 0xb6, 0x94, 0x49, 0xfc, 0xdc, 0x95, 0xf4, 0xae, 0xf5, 0x56,
	0x77, 0x3a, 0x0e, 0xff, 0xeb, 0xb7, 0x97, 0x00, 0x00, 0x00, 0xff, 0xff, 0x05, 0xcc, 0x83, 0x6c,
	0xd2, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// PubSubClient is the client API for PubSub service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PubSubClient interface {
	Connect(ctx context.Context, in *Url, opts ...grpc.CallOption) (*Id, error)
	Publish(ctx context.Context, opts ...grpc.CallOption) (PubSub_PublishClient, error)
	Subscribe(ctx context.Context, in *ImageStreamParameters, opts ...grpc.CallOption) (PubSub_SubscribeClient, error)
	Unsubscribe(ctx context.Context, in *CameraInfo, opts ...grpc.CallOption) (*Status, error)
	LatencyCalc(ctx context.Context, in *LatencyMeasured, opts ...grpc.CallOption) (*Status, error)
}

type pubSubClient struct {
	cc *grpc.ClientConn
}

func NewPubSubClient(cc *grpc.ClientConn) PubSubClient {
	return &pubSubClient{cc}
}

func (c *pubSubClient) Connect(ctx context.Context, in *Url, opts ...grpc.CallOption) (*Id, error) {
	out := new(Id)
	err := c.cc.Invoke(ctx, "/edgenode.PubSub/Connect", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pubSubClient) Publish(ctx context.Context, opts ...grpc.CallOption) (PubSub_PublishClient, error) {
	stream, err := c.cc.NewStream(ctx, &_PubSub_serviceDesc.Streams[0], "/edgenode.PubSub/Publish", opts...)
	if err != nil {
		return nil, err
	}
	x := &pubSubPublishClient{stream}
	return x, nil
}

type PubSub_PublishClient interface {
	Send(*Image) error
	CloseAndRecv() (*Status, error)
	grpc.ClientStream
}

type pubSubPublishClient struct {
	grpc.ClientStream
}

func (x *pubSubPublishClient) Send(m *Image) error {
	return x.ClientStream.SendMsg(m)
}

func (x *pubSubPublishClient) CloseAndRecv() (*Status, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(Status)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *pubSubClient) Subscribe(ctx context.Context, in *ImageStreamParameters, opts ...grpc.CallOption) (PubSub_SubscribeClient, error) {
	stream, err := c.cc.NewStream(ctx, &_PubSub_serviceDesc.Streams[1], "/edgenode.PubSub/Subscribe", opts...)
	if err != nil {
		return nil, err
	}
	x := &pubSubSubscribeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type PubSub_SubscribeClient interface {
	Recv() (*Image, error)
	grpc.ClientStream
}

type pubSubSubscribeClient struct {
	grpc.ClientStream
}

func (x *pubSubSubscribeClient) Recv() (*Image, error) {
	m := new(Image)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *pubSubClient) Unsubscribe(ctx context.Context, in *CameraInfo, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/edgenode.PubSub/Unsubscribe", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pubSubClient) LatencyCalc(ctx context.Context, in *LatencyMeasured, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/edgenode.PubSub/LatencyCalc", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PubSubServer is the server API for PubSub service.
type PubSubServer interface {
	Connect(context.Context, *Url) (*Id, error)
	Publish(PubSub_PublishServer) error
	Subscribe(*ImageStreamParameters, PubSub_SubscribeServer) error
	Unsubscribe(context.Context, *CameraInfo) (*Status, error)
	LatencyCalc(context.Context, *LatencyMeasured) (*Status, error)
}

func RegisterPubSubServer(s *grpc.Server, srv PubSubServer) {
	s.RegisterService(&_PubSub_serviceDesc, srv)
}

func _PubSub_Connect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Url)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PubSubServer).Connect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgenode.PubSub/Connect",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PubSubServer).Connect(ctx, req.(*Url))
	}
	return interceptor(ctx, in, info, handler)
}

func _PubSub_Publish_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(PubSubServer).Publish(&pubSubPublishServer{stream})
}

type PubSub_PublishServer interface {
	SendAndClose(*Status) error
	Recv() (*Image, error)
	grpc.ServerStream
}

type pubSubPublishServer struct {
	grpc.ServerStream
}

func (x *pubSubPublishServer) SendAndClose(m *Status) error {
	return x.ServerStream.SendMsg(m)
}

func (x *pubSubPublishServer) Recv() (*Image, error) {
	m := new(Image)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _PubSub_Subscribe_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ImageStreamParameters)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PubSubServer).Subscribe(m, &pubSubSubscribeServer{stream})
}

type PubSub_SubscribeServer interface {
	Send(*Image) error
	grpc.ServerStream
}

type pubSubSubscribeServer struct {
	grpc.ServerStream
}

func (x *pubSubSubscribeServer) Send(m *Image) error {
	return x.ServerStream.SendMsg(m)
}

func _PubSub_Unsubscribe_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CameraInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PubSubServer).Unsubscribe(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgenode.PubSub/Unsubscribe",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PubSubServer).Unsubscribe(ctx, req.(*CameraInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _PubSub_LatencyCalc_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LatencyMeasured)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PubSubServer).LatencyCalc(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgenode.PubSub/LatencyCalc",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PubSubServer).LatencyCalc(ctx, req.(*LatencyMeasured))
	}
	return interceptor(ctx, in, info, handler)
}

var _PubSub_serviceDesc = grpc.ServiceDesc{
	ServiceName: "edgenode.PubSub",
	HandlerType: (*PubSubServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Connect",
			Handler:    _PubSub_Connect_Handler,
		},
		{
			MethodName: "Unsubscribe",
			Handler:    _PubSub_Unsubscribe_Handler,
		},
		{
			MethodName: "LatencyCalc",
			Handler:    _PubSub_LatencyCalc_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Publish",
			Handler:       _PubSub_Publish_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "Subscribe",
			Handler:       _PubSub_Subscribe_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "api/edgenode/edgenode_api.proto",
}
