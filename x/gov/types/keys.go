package types

import (
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	// ModuleName defines the module name
	ModuleName = "gov"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for governance
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_gov"

	// ECDSASignatureKeyPrefix is the prefix for storing ECDSA signatures
	ECDSASignatureKeyPrefix = "ecdsa_signature_value"

	// ProposalKeyPrefix is the prefix for storing proposals
	ProposalKeyPrefix = "proposal_value"

	// VoteKeyPrefix is the prefix for storing votes
	VoteKeyPrefix = "vote_value"

	// DepositKeyPrefix is the prefix for storing deposits
	DepositKeyPrefix = "deposit_value"

	// DefaultVotingPeriod is the default voting period in blocks
	DefaultVotingPeriod = uint64(10080) // 1 week at 5s per block

	// DefaultMinDeposit is the default minimum deposit
	DefaultMinDeposit = "1000000uhome" // 1 HOME token

	// DefaultQuorum is the default quorum percentage
	DefaultQuorum = uint64(33) // 33%

	// DefaultThreshold is the default threshold percentage
	DefaultThreshold = uint64(50) // 50%

	// DefaultVetoThreshold is the default veto threshold percentage
	DefaultVetoThreshold = uint64(33) // 33%
)

// ECDSASignatureKey returns the store key for a specific ECDSA signature
func ECDSASignatureKey(address string, nonce uint64) []byte {
	return []byte(fmt.Sprintf("%s:%s:%d", ECDSASignatureKeyPrefix, address, nonce))
}

// ProposalKey returns the store key for a specific proposal
func ProposalKey(proposalID uint64) []byte {
	return []byte(fmt.Sprintf("%s:%d", ProposalKeyPrefix, proposalID))
}

// VoteKey returns the store key for a specific vote
func VoteKey(proposalID uint64, voter string) []byte {
	return []byte(fmt.Sprintf("%s:%d:%s", VoteKeyPrefix, proposalID, voter))
}

// DepositKey returns the store key for a specific deposit
func DepositKey(proposalID uint64, depositor string) []byte {
	return []byte(fmt.Sprintf("%s:%d:%s", DepositKeyPrefix, proposalID, depositor))
}

// IsValidEthereumAddress checks if the address is a valid Ethereum address
func IsValidEthereumAddress(address string) bool {
	if len(address) != 42 {
		return false
	}
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	_, err := hex.DecodeString(address[2:])
	return err == nil
}

// NormalizeEthereumAddress normalizes an Ethereum address to lowercase
func NormalizeEthereumAddress(address string) string {
	return strings.ToLower(address)
}

// ECDSASignature represents an ECDSA signature with recovery parameters
type ECDSASignature struct {
	Address   string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty" yaml:"address"`
	Signature string `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty" yaml:"signature"`
	Nonce     uint64 `protobuf:"varint,3,opt,name=nonce,proto3" json:"nonce,omitempty" yaml:"nonce"`
	Message   string `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty" yaml:"message"`
	Timestamp uint64 `protobuf:"varint,5,opt,name=timestamp,proto3" json:"timestamp,omitempty" yaml:"timestamp"`
}

// NewECDSASignature creates a new ECDSASignature instance
func NewECDSASignature(address, signature, message string, nonce, timestamp uint64) ECDSASignature {
	return ECDSASignature{
		Address:   address,
		Signature: signature,
		Nonce:     nonce,
		Message:   message,
		Timestamp: timestamp,
	}
}

// Validate validates the ECDSASignature
func (e ECDSASignature) Validate() error {
	if !IsValidEthereumAddress(e.Address) {
		return fmt.Errorf("invalid ethereum address: %s", e.Address)
	}
	if len(e.Signature) != 132 { // 0x + 130 hex chars (65 bytes)
		return fmt.Errorf("invalid signature length: expected 132, got %d", len(e.Signature))
	}
	if !strings.HasPrefix(e.Signature, "0x") {
		return fmt.Errorf("signature must start with 0x")
	}
	if len(e.Message) == 0 {
		return fmt.Errorf("message cannot be empty")
	}
	return nil
}

// ProposalStatus represents the status of a proposal
type ProposalStatus int32

const (
	// PROPOSAL_STATUS_UNSPECIFIED defines the default proposal status
	ProposalStatusUnspecified ProposalStatus = 0
	// PROPOSAL_STATUS_DEPOSIT_PERIOD defines a proposal in deposit period
	ProposalStatusDepositPeriod ProposalStatus = 1
	// PROPOSAL_STATUS_VOTING_PERIOD defines a proposal in voting period
	ProposalStatusVotingPeriod ProposalStatus = 2
	// PROPOSAL_STATUS_PASSED defines a proposal that has passed
	ProposalStatusPassed ProposalStatus = 3
	// PROPOSAL_STATUS_REJECTED defines a proposal that has been rejected
	ProposalStatusRejected ProposalStatus = 4
	// PROPOSAL_STATUS_FAILED defines a proposal that has failed
	ProposalStatusFailed ProposalStatus = 5
)

// String implements the Stringer interface for ProposalStatus
func (ps ProposalStatus) String() string {
	switch ps {
	case ProposalStatusUnspecified:
		return "PROPOSAL_STATUS_UNSPECIFIED"
	case ProposalStatusDepositPeriod:
		return "PROPOSAL_STATUS_DEPOSIT_PERIOD"
	case ProposalStatusVotingPeriod:
		return "PROPOSAL_STATUS_VOTING_PERIOD"
	case ProposalStatusPassed:
		return "PROPOSAL_STATUS_PASSED"
	case ProposalStatusRejected:
		return "PROPOSAL_STATUS_REJECTED"
	case ProposalStatusFailed:
		return "PROPOSAL_STATUS_FAILED"
	default:
		return "PROPOSAL_STATUS_UNKNOWN"
	}
}

// VoteOption represents a vote option
type VoteOption int32

const (
	// VOTE_OPTION_UNSPECIFIED defines an unspecified vote option
	VoteOptionUnspecified VoteOption = 0
	// VOTE_OPTION_YES defines a yes vote option
	VoteOptionYes VoteOption = 1
	// VOTE_OPTION_ABSTAIN defines an abstain vote option
	VoteOptionAbstain VoteOption = 2
	// VOTE_OPTION_NO defines a no vote option
	VoteOptionNo VoteOption = 3
	// VOTE_OPTION_NO_WITH_VETO defines a no with veto vote option
	VoteOptionNoWithVeto VoteOption = 4
)

// String implements the Stringer interface for VoteOption
func (vo VoteOption) String() string {
	switch vo {
	case VoteOptionUnspecified:
		return "VOTE_OPTION_UNSPECIFIED"
	case VoteOptionYes:
		return "VOTE_OPTION_YES"
	case VoteOptionAbstain:
		return "VOTE_OPTION_ABSTAIN"
	case VoteOptionNo:
		return "VOTE_OPTION_NO"
	case VoteOptionNoWithVeto:
		return "VOTE_OPTION_NO_WITH_VETO"
	default:
		return "VOTE_OPTION_UNKNOWN"
	}
}

// ProtoMessage stub implementations for codec compatibility
func (e *ECDSASignature) ProtoMessage() {}
func (e *ECDSASignature) Reset()        { *e = ECDSASignature{} }
func (e ECDSASignature) String() string { return e.Address }
