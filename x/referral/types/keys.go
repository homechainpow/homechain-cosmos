package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "referral"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for referral
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_referral"

	// ReferralKeyPrefix is the prefix for storing referral data
	ReferralKeyPrefix = "referral_value"

	// RewardClaimKeyPrefix is the prefix for storing reward claim data
	RewardClaimKeyPrefix = "reward_claim_value"

	// SettlementKeyPrefix is the prefix for storing settlement data
	SettlementKeyPrefix = "settlement_value"

	// ModuleAccount names for reward pools
	MiningRewardPool = "mining_reward_pool"
	ReferralPool     = "referral_pool"
	TreasuryPool     = "treasury_pool"

	// Default reward distribution percentages
	DefaultMiningRewardPercentage = uint64(70) // 70% to miners
	DefaultReferralPercentage     = uint64(20) // 20% to referrals
	DefaultTreasuryPercentage     = uint64(10) // 10% to treasury

	// Settlement period (in blocks)
	DefaultSettlementPeriod = uint64(100) // Every 100 blocks
)

// ModuleAccount permissions for reward pools
var (
	// MiningRewardPool permissions - can mint and burn mining rewards
	MiningRewardPoolPermissions = []string{"minter", "burner"}

	// ReferralPool permissions - can mint and burn referral rewards
	ReferralPoolPermissions = []string{"minter", "burner"}

	// TreasuryPool permissions - can mint and burn treasury funds
	TreasuryPoolPermissions = []string{"minter", "burner"}
)

// ReferralKey returns the store key for a specific referral
func ReferralKey(referrer string) []byte {
	return []byte(ReferralKeyPrefix + ":" + referrer)
}

// RewardClaimKey returns the store key for a specific reward claim
func RewardClaimKey(recipient string, claimID string) []byte {
	return []byte(RewardClaimKeyPrefix + ":" + recipient + ":" + claimID)
}

// SettlementKey returns the store key for a specific settlement
func SettlementKey(period uint64) []byte {
	return []byte(SettlementKeyPrefix + ":" + strconv.FormatUint(period, 10))
}

// RewardPoolKey returns the store key for a specific reward pool
func RewardPoolKey(name string) []byte {
	return []byte("reward_pool:" + name)
}

// GetMiningRewardPoolAddress returns the address of the mining reward pool
func GetMiningRewardPoolAddress() sdk.AccAddress {
	return sdk.AccAddress([]byte(MiningRewardPool))
}

// GetReferralPoolAddress returns the address of the referral pool
func GetReferralPoolAddress() sdk.AccAddress {
	return sdk.AccAddress([]byte(ReferralPool))
}

// GetTreasuryPoolAddress returns the address of the treasury pool
func GetTreasuryPoolAddress() sdk.AccAddress {
	return sdk.AccAddress([]byte(TreasuryPool))
}
