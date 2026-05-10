package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgStake defines a message for staking tokens on a node
type MsgStake struct {
	Address string    `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Amount  sdk.Coins `protobuf:"bytes,2,rep,name=amount,proto3" json:"amount"`
	Boost   uint64    `protobuf:"varint,3,opt,name=boost,proto3" json:"boost,omitempty"`
}

// MsgStakeResponse defines the response for staking
type MsgStakeResponse struct{}

// MsgUnstake defines a message for unstaking tokens
type MsgUnstake struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Amount  string `protobuf:"bytes,2,opt,name=amount,proto3" json:"amount,omitempty"`
}

// MsgUnstakeResponse defines the response for unstaking
type MsgUnstakeResponse struct{}

// MsgUpdateBoost defines a message for updating boost multiplier
type MsgUpdateBoost struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Boost   uint64 `protobuf:"varint,2,opt,name=boost,proto3" json:"boost,omitempty"`
}

// MsgUpdateBoostResponse defines the response for updating boost
type MsgUpdateBoostResponse struct{}

// StakeInfo stores staking information for an address
type StakeInfo struct {
	Address string    `json:"address"`
	Amount  sdk.Coins `json:"amount"`
	Boost   uint64    `json:"boost"`
}

// ProtoMessage stub
func (s *StakeInfo) ProtoMessage() {}

// Reset stub
func (s *StakeInfo) Reset() { *s = StakeInfo{} }

// String stub
func (s StakeInfo) String() string { return s.Address }

// StakeKey returns the store key for a stake
func StakeKey(address string) []byte {
	return []byte("stake:" + address)
}

// RegisterMsgServer registers the message server
func RegisterMsgServer(server interface{}, msgServer MsgServer) {}

// MsgServer defines the message server interface
type MsgServer interface {
	Stake(goCtx context.Context, msg *MsgStake) (*MsgStakeResponse, error)
	Unstake(goCtx context.Context, msg *MsgUnstake) (*MsgUnstakeResponse, error)
	UpdateBoost(goCtx context.Context, msg *MsgUpdateBoost) (*MsgUpdateBoostResponse, error)
}

// QueryServer defines the query server interface
type QueryServer interface {
}

// ProtoMessage stubs
func (m *MsgStake) ProtoMessage() {}
func (m *MsgStake) Reset()        { *m = MsgStake{} }
func (m MsgStake) String() string { return m.Address }

func (m *MsgStakeResponse) ProtoMessage() {}
func (m *MsgStakeResponse) Reset()        { *m = MsgStakeResponse{} }
func (m MsgStakeResponse) String() string { return "MsgStakeResponse" }

func (m *MsgUnstake) ProtoMessage() {}
func (m *MsgUnstake) Reset()        { *m = MsgUnstake{} }
func (m MsgUnstake) String() string { return m.Address }

func (m *MsgUnstakeResponse) ProtoMessage() {}
func (m *MsgUnstakeResponse) Reset()        { *m = MsgUnstakeResponse{} }
func (m MsgUnstakeResponse) String() string { return "MsgUnstakeResponse" }

func (m *MsgUpdateBoost) ProtoMessage() {}
func (m *MsgUpdateBoost) Reset()        { *m = MsgUpdateBoost{} }
func (m MsgUpdateBoost) String() string { return m.Address }

func (m *MsgUpdateBoostResponse) ProtoMessage() {}
func (m *MsgUpdateBoostResponse) Reset()        { *m = MsgUpdateBoostResponse{} }
func (m MsgUpdateBoostResponse) String() string { return "MsgUpdateBoostResponse" }
