// Code generated by protoc-gen-go. DO NOT EDIT.
// source: data.proto

package bean

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type BlockHeader struct {
	Version              int32    `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	PreviousHash         []byte   `protobuf:"bytes,2,opt,name=previous_hash,json=previousHash,proto3" json:"previous_hash,omitempty"`
	GasLimit             []byte   `protobuf:"bytes,3,opt,name=gas_limit,json=gasLimit,proto3" json:"gas_limit,omitempty"`
	GasUsed              []byte   `protobuf:"bytes,4,opt,name=gas_used,json=gasUsed,proto3" json:"gas_used,omitempty"`
	Height               []byte   `protobuf:"bytes,5,opt,name=height,proto3" json:"height,omitempty"`
	Timestamp            int64    `protobuf:"varint,6,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	MerkleRoot           []byte   `protobuf:"bytes,7,opt,name=merkle_root,json=merkleRoot,proto3" json:"merkle_root,omitempty"`
	TxHashes             [][]byte `protobuf:"bytes,8,rep,name=tx_hashes,json=txHashes,proto3" json:"tx_hashes,omitempty"`
	LeaderPubKey         *Point   `protobuf:"bytes,9,opt,name=leader_pub_key,json=leaderPubKey,proto3" json:"leader_pub_key,omitempty"`
	MinorPubKeys         []*Point `protobuf:"bytes,10,rep,name=minor_pub_keys,json=minorPubKeys,proto3" json:"minor_pub_keys,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockHeader) Reset()         { *m = BlockHeader{} }
func (m *BlockHeader) String() string { return proto.CompactTextString(m) }
func (*BlockHeader) ProtoMessage()    {}
func (*BlockHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_data_7d5731588313d330, []int{0}
}
func (m *BlockHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockHeader.Unmarshal(m, b)
}
func (m *BlockHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockHeader.Marshal(b, m, deterministic)
}
func (dst *BlockHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockHeader.Merge(dst, src)
}
func (m *BlockHeader) XXX_Size() int {
	return xxx_messageInfo_BlockHeader.Size(m)
}
func (m *BlockHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockHeader.DiscardUnknown(m)
}

var xxx_messageInfo_BlockHeader proto.InternalMessageInfo

func (m *BlockHeader) GetVersion() int32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *BlockHeader) GetPreviousHash() []byte {
	if m != nil {
		return m.PreviousHash
	}
	return nil
}

func (m *BlockHeader) GetGasLimit() []byte {
	if m != nil {
		return m.GasLimit
	}
	return nil
}

func (m *BlockHeader) GetGasUsed() []byte {
	if m != nil {
		return m.GasUsed
	}
	return nil
}

func (m *BlockHeader) GetHeight() []byte {
	if m != nil {
		return m.Height
	}
	return nil
}

func (m *BlockHeader) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *BlockHeader) GetMerkleRoot() []byte {
	if m != nil {
		return m.MerkleRoot
	}
	return nil
}

func (m *BlockHeader) GetTxHashes() [][]byte {
	if m != nil {
		return m.TxHashes
	}
	return nil
}

func (m *BlockHeader) GetLeaderPubKey() *Point {
	if m != nil {
		return m.LeaderPubKey
	}
	return nil
}

func (m *BlockHeader) GetMinorPubKeys() []*Point {
	if m != nil {
		return m.MinorPubKeys
	}
	return nil
}

type BlockData struct {
	TxCount              int32          `protobuf:"varint,1,opt,name=tx_count,json=txCount,proto3" json:"tx_count,omitempty"`
	TxList               []*Transaction `protobuf:"bytes,2,rep,name=tx_list,json=txList,proto3" json:"tx_list,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *BlockData) Reset()         { *m = BlockData{} }
func (m *BlockData) String() string { return proto.CompactTextString(m) }
func (*BlockData) ProtoMessage()    {}
func (*BlockData) Descriptor() ([]byte, []int) {
	return fileDescriptor_data_7d5731588313d330, []int{1}
}
func (m *BlockData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockData.Unmarshal(m, b)
}
func (m *BlockData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockData.Marshal(b, m, deterministic)
}
func (dst *BlockData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockData.Merge(dst, src)
}
func (m *BlockData) XXX_Size() int {
	return xxx_messageInfo_BlockData.Size(m)
}
func (m *BlockData) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockData.DiscardUnknown(m)
}

var xxx_messageInfo_BlockData proto.InternalMessageInfo

func (m *BlockData) GetTxCount() int32 {
	if m != nil {
		return m.TxCount
	}
	return 0
}

func (m *BlockData) GetTxList() []*Transaction {
	if m != nil {
		return m.TxList
	}
	return nil
}

type Block struct {
	Header               *BlockHeader `protobuf:"bytes,2,opt,name=header,proto3" json:"header,omitempty"`
	Data                 *BlockData   `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	MultiSig             *Signature   `protobuf:"bytes,4,opt,name=multi_sig,json=multiSig,proto3" json:"multi_sig,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Block) Reset()         { *m = Block{} }
func (m *Block) String() string { return proto.CompactTextString(m) }
func (*Block) ProtoMessage()    {}
func (*Block) Descriptor() ([]byte, []int) {
	return fileDescriptor_data_7d5731588313d330, []int{2}
}
func (m *Block) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Block.Unmarshal(m, b)
}
func (m *Block) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Block.Marshal(b, m, deterministic)
}
func (dst *Block) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Block.Merge(dst, src)
}
func (m *Block) XXX_Size() int {
	return xxx_messageInfo_Block.Size(m)
}
func (m *Block) XXX_DiscardUnknown() {
	xxx_messageInfo_Block.DiscardUnknown(m)
}

var xxx_messageInfo_Block proto.InternalMessageInfo

func (m *Block) GetHeader() *BlockHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *Block) GetData() *BlockData {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Block) GetMultiSig() *Signature {
	if m != nil {
		return m.MultiSig
	}
	return nil
}

type TransactionData struct {
	Version              int32    `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Nonce                int64    `protobuf:"varint,2,opt,name=nonce,proto3" json:"nonce,omitempty"`
	To                   string   `protobuf:"bytes,3,opt,name=to,proto3" json:"to,omitempty"`
	Amount               []byte   `protobuf:"bytes,4,opt,name=amount,proto3" json:"amount,omitempty"`
	GasPrice             []byte   `protobuf:"bytes,5,opt,name=gas_price,json=gasPrice,proto3" json:"gas_price,omitempty"`
	GasLimit             []byte   `protobuf:"bytes,6,opt,name=gas_limit,json=gasLimit,proto3" json:"gas_limit,omitempty"`
	Timestamp            int64    `protobuf:"varint,7,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	PubKey               *Point   `protobuf:"bytes,8,opt,name=pub_key,json=pubKey,proto3" json:"pub_key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TransactionData) Reset()         { *m = TransactionData{} }
func (m *TransactionData) String() string { return proto.CompactTextString(m) }
func (*TransactionData) ProtoMessage()    {}
func (*TransactionData) Descriptor() ([]byte, []int) {
	return fileDescriptor_data_7d5731588313d330, []int{3}
}
func (m *TransactionData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TransactionData.Unmarshal(m, b)
}
func (m *TransactionData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TransactionData.Marshal(b, m, deterministic)
}
func (dst *TransactionData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TransactionData.Merge(dst, src)
}
func (m *TransactionData) XXX_Size() int {
	return xxx_messageInfo_TransactionData.Size(m)
}
func (m *TransactionData) XXX_DiscardUnknown() {
	xxx_messageInfo_TransactionData.DiscardUnknown(m)
}

var xxx_messageInfo_TransactionData proto.InternalMessageInfo

func (m *TransactionData) GetVersion() int32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *TransactionData) GetNonce() int64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

func (m *TransactionData) GetTo() string {
	if m != nil {
		return m.To
	}
	return ""
}

func (m *TransactionData) GetAmount() []byte {
	if m != nil {
		return m.Amount
	}
	return nil
}

func (m *TransactionData) GetGasPrice() []byte {
	if m != nil {
		return m.GasPrice
	}
	return nil
}

func (m *TransactionData) GetGasLimit() []byte {
	if m != nil {
		return m.GasLimit
	}
	return nil
}

func (m *TransactionData) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *TransactionData) GetPubKey() *Point {
	if m != nil {
		return m.PubKey
	}
	return nil
}

type Transaction struct {
	Data                 *TransactionData `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Sig                  *Signature       `protobuf:"bytes,2,opt,name=sig,proto3" json:"sig,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *Transaction) Reset()         { *m = Transaction{} }
func (m *Transaction) String() string { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()    {}
func (*Transaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_data_7d5731588313d330, []int{4}
}
func (m *Transaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transaction.Unmarshal(m, b)
}
func (m *Transaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transaction.Marshal(b, m, deterministic)
}
func (dst *Transaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transaction.Merge(dst, src)
}
func (m *Transaction) XXX_Size() int {
	return xxx_messageInfo_Transaction.Size(m)
}
func (m *Transaction) XXX_DiscardUnknown() {
	xxx_messageInfo_Transaction.DiscardUnknown(m)
}

var xxx_messageInfo_Transaction proto.InternalMessageInfo

func (m *Transaction) GetData() *TransactionData {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Transaction) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type MultiSignature struct {
	Sig                  *Signature `protobuf:"bytes,1,opt,name=sig,proto3" json:"sig,omitempty"`
	Bitmap               []byte     `protobuf:"bytes,2,opt,name=bitmap,proto3" json:"bitmap,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *MultiSignature) Reset()         { *m = MultiSignature{} }
func (m *MultiSignature) String() string { return proto.CompactTextString(m) }
func (*MultiSignature) ProtoMessage()    {}
func (*MultiSignature) Descriptor() ([]byte, []int) {
	return fileDescriptor_data_7d5731588313d330, []int{5}
}
func (m *MultiSignature) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MultiSignature.Unmarshal(m, b)
}
func (m *MultiSignature) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MultiSignature.Marshal(b, m, deterministic)
}
func (dst *MultiSignature) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MultiSignature.Merge(dst, src)
}
func (m *MultiSignature) XXX_Size() int {
	return xxx_messageInfo_MultiSignature.Size(m)
}
func (m *MultiSignature) XXX_DiscardUnknown() {
	xxx_messageInfo_MultiSignature.DiscardUnknown(m)
}

var xxx_messageInfo_MultiSignature proto.InternalMessageInfo

func (m *MultiSignature) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

func (m *MultiSignature) GetBitmap() []byte {
	if m != nil {
		return m.Bitmap
	}
	return nil
}

func init() {
	proto.RegisterType((*BlockHeader)(nil), "bean.block_header")
	proto.RegisterType((*BlockData)(nil), "bean.block_data")
	proto.RegisterType((*Block)(nil), "bean.block")
	proto.RegisterType((*TransactionData)(nil), "bean.transaction_data")
	proto.RegisterType((*Transaction)(nil), "bean.transaction")
	proto.RegisterType((*MultiSignature)(nil), "bean.multi_signature")
}

func init() { proto.RegisterFile("data.proto", fileDescriptor_data_7d5731588313d330) }

var fileDescriptor_data_7d5731588313d330 = []byte{
	// 536 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x53, 0xdf, 0x6e, 0x9b, 0x3e,
	0x18, 0x15, 0xd0, 0x40, 0xf8, 0xc8, 0xaf, 0xed, 0xcf, 0x9a, 0x2a, 0xef, 0x8f, 0x34, 0xc6, 0x7a,
	0x11, 0x55, 0x53, 0xa4, 0x65, 0x8f, 0xb0, 0x9b, 0x4a, 0xcb, 0x45, 0x45, 0xb5, 0xbb, 0x49, 0xc8,
	0x24, 0x16, 0x58, 0x01, 0x8c, 0xf0, 0x47, 0x45, 0x9e, 0x60, 0x8f, 0xb9, 0xdb, 0x3d, 0xc6, 0x64,
	0x1b, 0xba, 0xb4, 0x59, 0xef, 0x38, 0x9f, 0xbf, 0x63, 0x1d, 0x9f, 0x73, 0x00, 0xd8, 0x31, 0x64,
	0xab, 0xb6, 0x93, 0x28, 0xc9, 0x59, 0xce, 0x59, 0xf3, 0x26, 0xdc, 0xf3, 0x83, 0x1d, 0x24, 0xbf,
	0x5c, 0x58, 0xe4, 0x95, 0xdc, 0xee, 0xb3, 0x92, 0xb3, 0x1d, 0xef, 0x08, 0x85, 0xe0, 0x81, 0x77,
	0x4a, 0xc8, 0x86, 0x3a, 0xb1, 0xb3, 0x9c, 0xa5, 0x13, 0x24, 0x1f, 0xe1, 0xbf, 0xb6, 0xe3, 0x0f,
	0x42, 0xf6, 0x2a, 0x2b, 0x99, 0x2a, 0xa9, 0x1b, 0x3b, 0xcb, 0x45, 0xba, 0x98, 0x86, 0xb7, 0x4c,
	0x95, 0xe4, 0x2d, 0x84, 0x05, 0x53, 0x59, 0x25, 0x6a, 0x81, 0xd4, 0x33, 0x0b, 0xf3, 0x82, 0xa9,
	0x8d, 0xc6, 0xe4, 0x35, 0xe8, 0xef, 0xac, 0x57, 0x7c, 0x47, 0xcf, 0xcc, 0x59, 0x50, 0x30, 0xf5,
	0x5d, 0xf1, 0x1d, 0xb9, 0x02, 0xbf, 0xe4, 0xa2, 0x28, 0x91, 0xce, 0xcc, 0xc1, 0x88, 0xc8, 0x3b,
	0x08, 0x51, 0xd4, 0x5c, 0x21, 0xab, 0x5b, 0xea, 0xc7, 0xce, 0xd2, 0x4b, 0xff, 0x0e, 0xc8, 0x7b,
	0x88, 0x6a, 0xde, 0xed, 0x2b, 0x9e, 0x75, 0x52, 0x22, 0x0d, 0x0c, 0x15, 0xec, 0x28, 0x95, 0x12,
	0xb5, 0x1c, 0x1c, 0x8c, 0x5a, 0xae, 0xe8, 0x3c, 0xf6, 0xb4, 0x1c, 0x1c, 0x6e, 0x0d, 0x26, 0x9f,
	0xe1, 0xbc, 0x32, 0x8f, 0xce, 0xda, 0x3e, 0xcf, 0xf6, 0xfc, 0x40, 0xc3, 0xd8, 0x59, 0x46, 0xeb,
	0x68, 0xa5, 0x5d, 0x5a, 0xb5, 0x52, 0x34, 0x98, 0x2e, 0xec, 0xca, 0x5d, 0x9f, 0x7f, 0xe3, 0x07,
	0x4d, 0xa9, 0x45, 0x23, 0x1f, 0x19, 0x8a, 0x42, 0xec, 0x9d, 0x50, 0xcc, 0x8a, 0x65, 0xa8, 0xe4,
	0x1e, 0xc0, 0x1a, 0xac, 0x63, 0xd0, 0x16, 0xe0, 0x90, 0x6d, 0x65, 0xdf, 0xe0, 0xe4, 0x2f, 0x0e,
	0x5f, 0x35, 0x24, 0x37, 0x10, 0xe0, 0x90, 0x55, 0x42, 0x21, 0x75, 0xcd, 0xa5, 0xff, 0xdb, 0x4b,
	0xb1, 0x63, 0x8d, 0x62, 0x5b, 0x14, 0xb2, 0x49, 0x7d, 0x1c, 0x36, 0x42, 0x61, 0xf2, 0xd3, 0x81,
	0x99, 0xb9, 0x95, 0xdc, 0x68, 0xe3, 0xb4, 0x42, 0x13, 0x47, 0xb4, 0x26, 0x96, 0x74, 0x9c, 0x69,
	0x3a, 0x6e, 0x90, 0x6b, 0x38, 0xd3, 0x22, 0x4c, 0x2e, 0xd1, 0xfa, 0xf2, 0x78, 0x53, 0xcf, 0x53,
	0x73, 0x4a, 0x3e, 0x41, 0x58, 0xf7, 0x15, 0x8a, 0x4c, 0x89, 0xc2, 0xc4, 0x14, 0xad, 0x2f, 0xec,
	0xaa, 0x12, 0x45, 0xc3, 0xb0, 0xef, 0x78, 0x3a, 0x37, 0x1b, 0xf7, 0xa2, 0x48, 0x7e, 0x3b, 0x70,
	0x79, 0xa4, 0xd0, 0xbe, 0xf2, 0xe5, 0x12, 0xbd, 0x82, 0x59, 0x23, 0x9b, 0x2d, 0x37, 0x6a, 0xbd,
	0xd4, 0x02, 0x72, 0x0e, 0x2e, 0x4a, 0x23, 0x2b, 0x4c, 0x5d, 0x94, 0xba, 0x0d, 0xac, 0x36, 0x1e,
	0xd9, 0x9a, 0x8c, 0x68, 0x6a, 0x57, 0xdb, 0x89, 0x2d, 0x1f, 0x8b, 0xa2, 0x1b, 0x75, 0xa7, 0xf1,
	0xd3, 0xea, 0xf9, 0xcf, 0xaa, 0xf7, 0xa4, 0x47, 0xc1, 0xf3, 0x1e, 0x5d, 0x43, 0x30, 0x55, 0x60,
	0x7e, 0x5a, 0x01, 0xbf, 0x35, 0x51, 0x26, 0x3f, 0x20, 0x3a, 0x7a, 0x29, 0xb9, 0x19, 0xdd, 0x74,
	0x0c, 0xe3, 0xea, 0x24, 0xac, 0x63, 0x4f, 0x3f, 0x80, 0xa7, 0xdd, 0x74, 0xff, 0xed, 0xa6, 0x3e,
	0x4b, 0x36, 0x70, 0xf1, 0x68, 0xbb, 0x9d, 0x4f, 0x2c, 0xe7, 0x65, 0x96, 0x76, 0x2a, 0x17, 0x58,
	0xb3, 0x76, 0xfc, 0x1b, 0x47, 0x94, 0xfb, 0xe6, 0xf7, 0xfe, 0xf2, 0x27, 0x00, 0x00, 0xff, 0xff,
	0x69, 0x32, 0xb9, 0xe3, 0xfd, 0x03, 0x00, 0x00,
}
