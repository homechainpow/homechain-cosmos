package types

import (
	"fmt"
)

const (
	// ModuleName defines the module name
	ModuleName = "poh"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_poh"

	// PoHSequenceKeyPrefix is the prefix for storing PoH sequence data
	PoHSequenceKeyPrefix = "poh_sequence_value"

	// CurrentPoHKey is the key for storing current PoH data
	CurrentPoHKey = "current_poh"

	// DifficultyKey is the key for storing current difficulty
	DifficultyKey = "difficulty"

	// GenesisPoHHash is the initial PoH hash for genesis
	GenesisPoHHash = "0000000000000000000000000000000000000000000000000000000000000000"

	// DefaultDifficulty is the default difficulty for PoH
	DefaultDifficulty = uint64(20)

	// HeartbeatPeriod is the number of blocks between PoH heartbeats
	HeartbeatPeriod = uint64(100)
)

// PoHSequenceKey returns the store key for a specific PoH sequence
func PoHSequenceKey(height uint64) []byte {
	return []byte(fmt.Sprintf("%s:%d", PoHSequenceKeyPrefix, height))
}

// CurrentPoHKey returns the key for current PoH data
func CurrentPoHKeyBytes() []byte {
	return []byte(CurrentPoHKey)
}

// DifficultyKeyBytes returns the key for difficulty data
func DifficultyKeyBytes() []byte {
	return []byte(DifficultyKey)
}
