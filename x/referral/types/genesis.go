package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers the amino codec for the module
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// TODO: Register message types when implemented
}

var ModuleCdc *codec.LegacyAmino

func init() {
	cdc := codec.NewLegacyAmino()
	RegisterLegacyAminoCodec(cdc)
	ModuleCdc = cdc
}

// ProtoMessage implements proto.Message interface
func (p *Params) ProtoMessage() {}

// Reset implements proto.Message interface
func (p *Params) Reset() { *p = Params{} }

// String implements proto.Message interface
func (p Params) String() string { return "Params" }

// Params defines the parameters for the referral module
type Params struct{}

// DefaultParams returns default referral parameters
func DefaultParams() Params {
	return Params{}
}

// RewardPool represents a reward pool
type RewardPool struct {
	Name          string
	Balance       sdk.Coins
	LastSettledAt uint64
	IsActive      bool
	Permissions   []string
}

// ProtoMessage implements proto.Message interface
func (p *RewardPool) ProtoMessage() {}

// Reset implements proto.Message interface
func (p *RewardPool) Reset() { *p = RewardPool{} }

// String implements proto.Message interface
func (p RewardPool) String() string { return p.Name }

// ReferralInfo represents referral information
type ReferralInfo struct {
	Referrer       string
	Referred       string
	JoinedAt       uint64
	Level          uint32
	TotalReferrals uint64
}

// NewReferralInfo creates a new ReferralInfo instance
func NewReferralInfo(referrer, referred string, joinedAt uint64, level uint32) ReferralInfo {
	return ReferralInfo{
		Referrer:       referrer,
		Referred:       referred,
		JoinedAt:       joinedAt,
		Level:          level,
		TotalReferrals: 0,
	}
}

// ProtoMessage implements proto.Message interface
func (r *ReferralInfo) ProtoMessage() {}

// Reset implements proto.Message interface
func (r *ReferralInfo) Reset() { *r = ReferralInfo{} }

// String implements proto.Message interface
func (r ReferralInfo) String() string { return r.Referrer }

// RewardClaim represents a reward claim
type RewardClaim struct {
	ClaimID   string
	Recipient string
	Amount    sdk.Coins
	ClaimedAt uint64
}

// ProtoMessage implements proto.Message interface
func (rc *RewardClaim) ProtoMessage() {}

// Reset implements proto.Message interface
func (rc *RewardClaim) Reset() { *rc = RewardClaim{} }

// String implements proto.Message interface
func (rc RewardClaim) String() string { return "RewardClaim" }

// GenesisState defines the referral module's genesis state
type GenesisState struct {
	Params       Params
	Pools        []RewardPool
	Referrals    []ReferralInfo
	RewardClaims []RewardClaim
	Settlements  []SettlementInfo
}

// SettlementInfo represents settlement information
type SettlementInfo struct {
	Period          uint64
	TotalMined      uint64
	MiningRewards   sdk.Coins
	ReferralRewards sdk.Coins
	TreasuryRewards sdk.Coins
	SettledAt       uint64
}

// NewSettlementInfo creates a new SettlementInfo instance
func NewSettlementInfo(period, totalMined uint64, miningRewards, referralRewards, treasuryRewards sdk.Coins, settledAt uint64) SettlementInfo {
	return SettlementInfo{
		Period:          period,
		TotalMined:      totalMined,
		MiningRewards:   miningRewards,
		ReferralRewards: referralRewards,
		TreasuryRewards: treasuryRewards,
		SettledAt:       settledAt,
	}
}

// ProtoMessage implements proto.Message interface
func (s *SettlementInfo) ProtoMessage() {}

// Reset implements proto.Message interface
func (s *SettlementInfo) Reset() { *s = SettlementInfo{} }

// String implements proto.Message interface
func (s SettlementInfo) String() string { return "SettlementInfo" }

// NewGenesisState creates a new GenesisState instance
func NewGenesisState(params Params, referrals []ReferralInfo, rewardPools []RewardPool, rewardClaims []RewardClaim, settlements []SettlementInfo) *GenesisState {
	return &GenesisState{
		Params:    params,
		Pools:     rewardPools,
		Referrals: referrals,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:    DefaultParams(),
		Pools:     []RewardPool{},
		Referrals: []ReferralInfo{},
	}
}

// Validate performs genesis state validation
func (gs GenesisState) Validate() error {
	return nil
}

// ProtoMessage implements proto.Message interface
func (gs *GenesisState) ProtoMessage() {}

// Reset implements proto.Message interface
func (gs *GenesisState) Reset() { *gs = GenesisState{} }

// String implements proto.Message interface
func (gs GenesisState) String() string { return "GenesisState" }
