package types

// Event types for the mining module
const (
	EventTypeMinerRegistered   = "miner_registered"
	EventTypeShareSubmitted    = "share_submitted"
	EventTypeMinerHeartbeat    = "miner_heartbeat"
	EventTypeMinerDeactivated  = "miner_deactivated"
)

// Attribute keys for the mining module events
const (
	AttributeKeyMiner      = "miner"
	AttributeKeyDeviceInfo = "device_info"
	AttributeKeyReferrer   = "referrer"
	AttributeKeyShareHash  = "share_hash"
	AttributeKeyDifficulty = "difficulty"
	AttributeKeyTimestamp  = "timestamp"
	AttributeKeyNonce      = "nonce"
	AttributeKeyAccepted   = "accepted"
	AttributeKeyReason     = "reason"
)
