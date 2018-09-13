// Code generated by protoc-gen-go. DO NOT EDIT.
// source: bean/consensus.proto

package bean

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import crypto "crypto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Setup struct {
	Msg                  []byte            `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	PubKey               *crypto.Point     `protobuf:"bytes,2,opt,name=pub_key,json=pubKey,proto3" json:"pub_key,omitempty"`
	Sig                  *crypto.Signature `protobuf:"bytes,3,opt,name=sig,proto3" json:"sig,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Setup) Reset()         { *m = Setup{} }
func (m *Setup) String() string { return proto.CompactTextString(m) }
func (*Setup) ProtoMessage()    {}
func (*Setup) Descriptor() ([]byte, []int) {
	return fileDescriptor_consensus_f9624335503358af, []int{0}
}
func (m *Setup) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Setup.Unmarshal(m, b)
}
func (m *Setup) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Setup.Marshal(b, m, deterministic)
}
func (dst *Setup) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Setup.Merge(dst, src)
}
func (m *Setup) XXX_Size() int {
	return xxx_messageInfo_Setup.Size(m)
}
func (m *Setup) XXX_DiscardUnknown() {
	xxx_messageInfo_Setup.DiscardUnknown(m)
}

var xxx_messageInfo_Setup proto.InternalMessageInfo

func (m *Setup) GetMsg() []byte {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (m *Setup) GetPubKey() *crypto.Point {
	if m != nil {
		return m.PubKey
	}
	return nil
}

func (m *Setup) GetSig() *crypto.Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type Commitment struct {
	PubKey               *crypto.Point `protobuf:"bytes,1,opt,name=pub_key,json=pubKey,proto3" json:"pub_key,omitempty"`
	Q                    *crypto.Point `protobuf:"bytes,2,opt,name=q,proto3" json:"q,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Commitment) Reset()         { *m = Commitment{} }
func (m *Commitment) String() string { return proto.CompactTextString(m) }
func (*Commitment) ProtoMessage()    {}
func (*Commitment) Descriptor() ([]byte, []int) {
	return fileDescriptor_consensus_f9624335503358af, []int{1}
}
func (m *Commitment) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Commitment.Unmarshal(m, b)
}
func (m *Commitment) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Commitment.Marshal(b, m, deterministic)
}
func (dst *Commitment) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Commitment.Merge(dst, src)
}
func (m *Commitment) XXX_Size() int {
	return xxx_messageInfo_Commitment.Size(m)
}
func (m *Commitment) XXX_DiscardUnknown() {
	xxx_messageInfo_Commitment.DiscardUnknown(m)
}

var xxx_messageInfo_Commitment proto.InternalMessageInfo

func (m *Commitment) GetPubKey() *crypto.Point {
	if m != nil {
		return m.PubKey
	}
	return nil
}

func (m *Commitment) GetQ() *crypto.Point {
	if m != nil {
		return m.Q
	}
	return nil
}

type Challenge struct {
	SigmaPubKey          *crypto.Point `protobuf:"bytes,1,opt,name=sigma_pub_key,json=sigmaPubKey,proto3" json:"sigma_pub_key,omitempty"`
	SigmaQ               *crypto.Point `protobuf:"bytes,2,opt,name=sigma_q,json=sigmaQ,proto3" json:"sigma_q,omitempty"`
	R                    []byte        `protobuf:"bytes,3,opt,name=r,proto3" json:"r,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Challenge) Reset()         { *m = Challenge{} }
func (m *Challenge) String() string { return proto.CompactTextString(m) }
func (*Challenge) ProtoMessage()    {}
func (*Challenge) Descriptor() ([]byte, []int) {
	return fileDescriptor_consensus_f9624335503358af, []int{2}
}
func (m *Challenge) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Challenge.Unmarshal(m, b)
}
func (m *Challenge) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Challenge.Marshal(b, m, deterministic)
}
func (dst *Challenge) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Challenge.Merge(dst, src)
}
func (m *Challenge) XXX_Size() int {
	return xxx_messageInfo_Challenge.Size(m)
}
func (m *Challenge) XXX_DiscardUnknown() {
	xxx_messageInfo_Challenge.DiscardUnknown(m)
}

var xxx_messageInfo_Challenge proto.InternalMessageInfo

func (m *Challenge) GetSigmaPubKey() *crypto.Point {
	if m != nil {
		return m.SigmaPubKey
	}
	return nil
}

func (m *Challenge) GetSigmaQ() *crypto.Point {
	if m != nil {
		return m.SigmaQ
	}
	return nil
}

func (m *Challenge) GetR() []byte {
	if m != nil {
		return m.R
	}
	return nil
}

type Response struct {
	PubKey               *crypto.Point `protobuf:"bytes,1,opt,name=pub_key,json=pubKey,proto3" json:"pub_key,omitempty"`
	S                    []byte        `protobuf:"bytes,2,opt,name=s,proto3" json:"s,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_consensus_f9624335503358af, []int{3}
}
func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Response.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Response.Marshal(b, m, deterministic)
}
func (dst *Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Response.Merge(dst, src)
}
func (m *Response) XXX_Size() int {
	return xxx_messageInfo_Response.Size(m)
}
func (m *Response) XXX_DiscardUnknown() {
	xxx_messageInfo_Response.DiscardUnknown(m)
}

var xxx_messageInfo_Response proto.InternalMessageInfo

func (m *Response) GetPubKey() *crypto.Point {
	if m != nil {
		return m.PubKey
	}
	return nil
}

func (m *Response) GetS() []byte {
	if m != nil {
		return m.S
	}
	return nil
}

func init() {
	proto.RegisterType((*Setup)(nil), "bean.setup")
	proto.RegisterType((*Commitment)(nil), "bean.commitment")
	proto.RegisterType((*Challenge)(nil), "bean.challenge")
	proto.RegisterType((*Response)(nil), "bean.response")
}

func init() { proto.RegisterFile("bean/consensus.proto", fileDescriptor_consensus_f9624335503358af) }

var fileDescriptor_consensus_f9624335503358af = []byte{
	// 248 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0x4f, 0x4b, 0xc3, 0x40,
	0x10, 0xc5, 0x19, 0xa3, 0xa9, 0x4e, 0x53, 0xa8, 0x8b, 0x87, 0xa0, 0x97, 0x12, 0xa1, 0xf4, 0x94,
	0xa2, 0x7e, 0x01, 0xef, 0x5e, 0x6c, 0xbe, 0x40, 0x49, 0xc2, 0xb8, 0x2e, 0xed, 0xfe, 0xe9, 0xce,
	0xee, 0x21, 0xdf, 0x5e, 0xb2, 0x41, 0xf0, 0xa0, 0xc5, 0xdb, 0x30, 0xf3, 0xe6, 0xf7, 0x78, 0x0f,
	0xef, 0x3a, 0x6a, 0xcd, 0xb6, 0xb7, 0x86, 0xc9, 0x70, 0xe4, 0xda, 0x79, 0x1b, 0xac, 0xb8, 0x1c,
	0xb7, 0xf7, 0xcb, 0xde, 0x0f, 0x2e, 0xd8, 0xed, 0x81, 0x86, 0x69, 0x5f, 0x7d, 0xe0, 0x15, 0x53,
	0x88, 0x4e, 0x2c, 0x31, 0xd3, 0x2c, 0x4b, 0x58, 0xc1, 0xa6, 0x68, 0xc6, 0x51, 0xac, 0x71, 0xe6,
	0x62, 0xb7, 0x3f, 0xd0, 0x50, 0x5e, 0xac, 0x60, 0x33, 0x7f, 0x5e, 0xd4, 0xd3, 0x7b, 0xed, 0xac,
	0x32, 0xa1, 0xc9, 0x5d, 0xec, 0xde, 0x68, 0x10, 0x8f, 0x98, 0xb1, 0x92, 0x65, 0x96, 0x34, 0xb7,
	0xdf, 0x1a, 0x56, 0xd2, 0xb4, 0x21, 0x7a, 0x6a, 0xc6, 0x6b, 0xb5, 0x43, 0xec, 0xad, 0xd6, 0x2a,
	0x68, 0x32, 0xe1, 0x27, 0x1a, 0xce, 0xa1, 0x1f, 0x10, 0x4e, 0xbf, 0x9b, 0xc3, 0xa9, 0x0a, 0x78,
	0xd3, 0x7f, 0xb6, 0xc7, 0x23, 0x19, 0x49, 0xe2, 0x09, 0x17, 0xac, 0xa4, 0x6e, 0xf7, 0x67, 0xb9,
	0xf3, 0xa4, 0x79, 0x9f, 0xe0, 0x6b, 0x9c, 0x4d, 0x2f, 0x7f, 0x58, 0xe4, 0xe9, 0xba, 0x13, 0x05,
	0x82, 0x4f, 0xe9, 0x8a, 0x06, 0x7c, 0xf5, 0x8a, 0xd7, 0x9e, 0xd8, 0x8d, 0xf5, 0xfe, 0x3b, 0x46,
	0x81, 0xc0, 0xc9, 0xa3, 0x68, 0x80, 0xbb, 0x3c, 0x35, 0xff, 0xf2, 0x15, 0x00, 0x00, 0xff, 0xff,
	0xf9, 0x95, 0x94, 0x36, 0xa9, 0x01, 0x00, 0x00,
}
