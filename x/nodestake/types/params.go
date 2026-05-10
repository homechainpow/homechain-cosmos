package types

import (
	"fmt"

	"cosmossdk.io/math"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// ParamStoreKey defines the key for nodestake parameters
var ParamStoreKey = []byte("Params")

// ParamKeyTable defines the param key table
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Params defines the parameters for the nodestake module
type Params struct {
	MinStakeAmount math.Int `protobuf:"bytes,1,opt,name=min_stake_amount,json=minStakeAmount,proto3" json:"min_stake_amount,omitempty"`
	MaxBoost       uint64   `protobuf:"varint,2,opt,name=max_boost,json=maxBoost,proto3" json:"max_boost,omitempty"`
}

// ParamSetPairs implements the ParamSet interface
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(ParamStoreKey, &p.MinStakeAmount, func(value interface{}) error { return nil }),
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MinStakeAmount: math.ZeroInt(),
		MaxBoost:       101,
	}
}

// Validate validates the params
func (p Params) Validate() error {
	if p.MaxBoost == 0 {
		return fmt.Errorf("max_boost must be greater than 0")
	}
	return nil
}

// String returns string representation
func (p Params) String() string {
	return fmt.Sprintf("Params{MinStakeAmount: %s, MaxBoost: %d}", p.MinStakeAmount, p.MaxBoost)
}

// ProtoMessage stub
func (p *Params) ProtoMessage() {}

// Reset stub
func (p *Params) Reset() { *p = Params{} }

// NewParams creates a new Params instance
func NewParams(minStakeAmount math.Int, maxBoost uint64) Params {
	return Params{
		MinStakeAmount: minStakeAmount,
		MaxBoost:       maxBoost,
	}
}
