package types

// Event types for the poh module
const (
	EventTypePoHSubmitted       = "poh_submitted"
	EventTypePoHVerified        = "poh_verified"
	EventTypeDifficultyAdjusted = "difficulty_adjusted"
)

// Attribute keys for the poh module events
const (
	AttributeKeySigner     = "signer"
	AttributeKeyPrevHash   = "prev_hash"
	AttributeKeyNewHash    = "new_hash"
	AttributeKeyDifficulty = "difficulty"
	AttributeKeyHeight     = "height"
	AttributeKeyTimestamp  = "timestamp"
	AttributeKeyValid      = "valid"
)

// Gas configuration constants for PoH operations
// Based on benchmark: Argon2id (t=1, m=64k, p=4) takes ~100ms on modern CPU.
// Standard Cosmos SDK 1ms execution ~ 1,000 Gas.
// So 100ms should be roughly 100,000 - 200,000 Gas.
const (
	DefaultPoHGasLimit uint64 = 250000 // High enough to cover Argon2 verification
)
