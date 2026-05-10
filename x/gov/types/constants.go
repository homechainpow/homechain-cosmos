package types

// Event types for the governance module
const (
	EventTypeProposalSubmitted = "proposal_submitted"
	EventTypeVoteCast          = "vote_cast"
	EventTypeDepositMade       = "deposit_made"
	EventTypeProposalPassed    = "proposal_passed"
	EventTypeProposalRejected  = "proposal_rejected"
	EventTypeProposalExpired   = "proposal_expired"
)

// Attribute keys for the governance module events
const (
	AttributeKeyProposalID = "proposal_id"
	AttributeKeyProposer   = "proposer"
	AttributeKeyTitle      = "title"
	AttributeKeyVoter      = "voter"
	AttributeKeyOption     = "option"
	AttributeKeyDepositor  = "depositor"
	AttributeKeyAmount     = "amount"
	AttributeKeyStatus     = "status"
)
