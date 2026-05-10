package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// GenesisState defines the nodestake module's genesis state
type GenesisState struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return DefaultGenesisState()
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(gs *GenesisState) error {
	return gs.Validate()
}

// ProtoMessage stub
func (gs *GenesisState) ProtoMessage() {}

// Reset stub
func (gs *GenesisState) Reset() { *gs = GenesisState{} }

// String stub
func (gs GenesisState) String() string { return "GenesisState" }

// RegisterLegacyAminoCodec registers amino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgStake{}, "homechain/MsgStake", nil)
	cdc.RegisterConcrete(&MsgUnstake{}, "homechain/MsgUnstake", nil)
	cdc.RegisterConcrete(&MsgUpdateBoost{}, "homechain/MsgUpdateBoost", nil)
}
