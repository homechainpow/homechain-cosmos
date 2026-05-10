package types

const (
	// ModuleName defines the module name
	ModuleName = "mining"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for mining
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_mining"

	// MinerKeyPrefix is the prefix for storing miner data
	MinerKeyPrefix = "miner_value"

	// ShareKeyPrefix is the prefix for storing share data
	ShareKeyPrefix = "share_value"

	// NonceKeyPrefix is the prefix for storing nonce data
	NonceKeyPrefix = "nonce_value"

	// RateLimitKeyPrefix is the prefix for storing rate limit data
	RateLimitKeyPrefix = "rate_limit_value"

	// MiningRewardPool is the name of the mining reward pool
	MiningRewardPool = "mining_reward_pool"

	// DefaultMaxSharesPerBlock is the default maximum shares per block
	DefaultMaxSharesPerBlock = uint64(100)

	// DefaultRateLimitPerMinute is the default rate limit per minute
	DefaultRateLimitPerMinute = uint64(10)

	// DefaultMinerInactiveThreshold is the default inactive threshold for miners (in blocks)
	DefaultMinerInactiveThreshold = uint64(1000)

	// ShareDifficulty is the fixed difficulty for mining shares
	ShareDifficulty = uint64(15)
)

// MinerKey returns the store key for a specific miner
func MinerKey(address string) []byte {
	return []byte(MinerKeyPrefix + ":" + address)
}

// ShareKey returns the store key for a specific share
func ShareKey(miner string, timestamp uint64, nonce string) []byte {
	return []byte(ShareKeyPrefix + ":" + miner + ":" + string(timestamp) + ":" + nonce)
}

// NonceKey returns the store key for a specific nonce
func NonceKey(miner string) []byte {
	return []byte(NonceKeyPrefix + ":" + miner)
}

// RateLimitKey returns the store key for rate limit data
func RateLimitKey(miner string) []byte {
	return []byte(RateLimitKeyPrefix + ":" + miner)
}
