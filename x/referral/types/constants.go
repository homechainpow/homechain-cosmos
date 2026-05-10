package types

// Event types for the reward module
const (
	EventTypeReferralUpdated   = "referral_updated"
	EventTypeRewardsClaimed    = "rewards_claimed"
	EventTypeRewardsSettled    = "rewards_settled"
	EventTypePoolCreated       = "pool_created"
	EventTypeRewardsDistributed = "rewards_distributed"
)

// Attribute keys for the reward module events
const (
	AttributeKeyReferrer   = "referrer"
	AttributeKeyReferred   = "referred"
	AttributeKeyRecipient  = "recipient"
	AttributeKeyAmount     = "amount"
	AttributeKeyPeriod     = "period"
	AttributeKeyPoolName   = "pool_name"
	AttributeKeyClaimType  = "claim_type"
	AttributeKeyClaimID    = "claim_id"
	AttributeKeyLevel      = "level"
)
