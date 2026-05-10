package types

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Proposal represents a governance proposal with ECDSA signature support
type Proposal struct {
	ProposalID        uint64         `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty" yaml:"proposal_id"`
	Title             string         `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty" yaml:"title"`
	Description       string         `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty" yaml:"description"`
	Proposer          string         `protobuf:"bytes,4,opt,name=proposer,proto3" json:"proposer,omitempty" yaml:"proposer"`
	ProposerSignature string         `protobuf:"bytes,5,opt,name=proposer_signature,json=proposerSignature,proto3" json:"proposer_signature,omitempty" yaml:"proposer_signature"`
	Nonce             uint64         `protobuf:"varint,6,opt,name=nonce,proto3" json:"nonce,omitempty" yaml:"nonce"`
	Status            ProposalStatus `protobuf:"varint,7,opt,name=status,proto3,enum=homechain.gov.v1.ProposalStatus" json:"status,omitempty" yaml:"status"`
	SubmitTime        uint64         `protobuf:"varint,8,opt,name=submit_time,json=submitTime,proto3" json:"submit_time,omitempty" yaml:"submit_time"`
	DepositEndTime    uint64         `protobuf:"varint,9,opt,name=deposit_end_time,json=depositEndTime,proto3" json:"deposit_end_time,omitempty" yaml:"deposit_end_time"`
	VotingStartTime   uint64         `protobuf:"varint,10,opt,name=voting_start_time,json=votingStartTime,proto3" json:"voting_start_time,omitempty" yaml:"voting_start_time"`
	VotingEndTime     uint64         `protobuf:"varint,11,opt,name=voting_end_time,json=votingEndTime,proto3" json:"voting_end_time,omitempty" yaml:"voting_end_time"`
	TotalDeposit      sdk.Coins      `protobuf:"bytes,12,rep,name=total_deposit,json=totalDeposit,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"total_deposit,omitempty" yaml:"total_deposit"`
	Votes             []Vote         `protobuf:"bytes,13,rep,name=votes,proto3" json:"votes,omitempty" yaml:"votes"`
}

// NewProposal creates a new Proposal instance
func NewProposal(proposalID uint64, title, description, proposer, proposerSignature string, nonce uint64, submitTime, depositEndTime, votingStartTime, votingEndTime uint64) Proposal {
	return Proposal{
		ProposalID:        proposalID,
		Title:             title,
		Description:       description,
		Proposer:          proposer,
		ProposerSignature: proposerSignature,
		Nonce:             nonce,
		Status:            ProposalStatusDepositPeriod,
		SubmitTime:        submitTime,
		DepositEndTime:    depositEndTime,
		VotingStartTime:   votingStartTime,
		VotingEndTime:     votingEndTime,
		TotalDeposit:      sdk.NewCoins(),
		Votes:             []Vote{},
	}
}

// Validate validates the Proposal
func (p Proposal) Validate() error {
	if p.ProposalID == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "proposal_id must be greater than 0")
	}
	if len(p.Title) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "title cannot be empty")
	}
	if len(p.Description) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "description cannot be empty")
	}
	if !IsValidEthereumAddress(p.Proposer) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid proposer address")
	}
	if len(p.ProposerSignature) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "proposer_signature cannot be empty")
	}
	if p.SubmitTime == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "submit_time must be greater than 0")
	}
	if p.DepositEndTime <= p.SubmitTime {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "deposit_end_time must be after submit_time")
	}
	if p.VotingStartTime <= p.DepositEndTime {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "voting_start_time must be after deposit_end_time")
	}
	if p.VotingEndTime <= p.VotingStartTime {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "voting_end_time must be after voting_start_time")
	}
	return nil
}

// Vote represents a vote on a proposal with ECDSA signature
type Vote struct {
	ProposalID uint64     `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty" yaml:"proposal_id"`
	Voter      string     `protobuf:"bytes,2,opt,name=voter,proto3" json:"voter,omitempty" yaml:"voter"`
	Signature  string     `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty" yaml:"signature"`
	Nonce      uint64     `protobuf:"varint,4,opt,name=nonce,proto3" json:"nonce,omitempty" yaml:"nonce"`
	Option     VoteOption `protobuf:"varint,5,opt,name=option,proto3,enum=homechain.gov.v1.VoteOption" json:"option,omitempty" yaml:"option"`
	Timestamp  uint64     `protobuf:"varint,6,opt,name=timestamp,proto3" json:"timestamp,omitempty" yaml:"timestamp"`
}

// NewVote creates a new Vote instance
func NewVote(proposalID uint64, voter, signature string, nonce uint64, option VoteOption, timestamp uint64) Vote {
	return Vote{
		ProposalID: proposalID,
		Voter:      voter,
		Signature:  signature,
		Nonce:      nonce,
		Option:     option,
		Timestamp:  timestamp,
	}
}

// Validate validates the Vote
func (v Vote) Validate() error {
	if v.ProposalID == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "proposal_id must be greater than 0")
	}
	if !IsValidEthereumAddress(v.Voter) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid voter address")
	}
	if len(v.Signature) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "signature cannot be empty")
	}
	if v.Timestamp == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "timestamp must be greater than 0")
	}
	return nil
}

// Deposit represents a deposit on a proposal
type Deposit struct {
	ProposalID uint64    `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty" yaml:"proposal_id"`
	Depositor  string    `protobuf:"bytes,2,opt,name=depositor,proto3" json:"depositor,omitempty" yaml:"depositor"`
	Amount     sdk.Coins `protobuf:"bytes,3,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount,omitempty" yaml:"amount"`
	Timestamp  uint64    `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp,omitempty" yaml:"timestamp"`
}

// NewDeposit creates a new Deposit instance
func NewDeposit(proposalID uint64, depositor string, amount sdk.Coins, timestamp uint64) Deposit {
	return Deposit{
		ProposalID: proposalID,
		Depositor:  depositor,
		Amount:     amount,
		Timestamp:  timestamp,
	}
}

// Validate validates the Deposit
func (d Deposit) Validate() error {
	if d.ProposalID == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "proposal_id must be greater than 0")
	}
	if len(d.Depositor) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "depositor cannot be empty")
	}
	if !d.Amount.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount must be valid coins")
	}
	if d.Amount.IsZero() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount cannot be zero")
	}
	if d.Timestamp == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "timestamp must be greater than 0")
	}
	return nil
}

// GenesisState defines the governance module's genesis state
type GenesisState struct {
	Params             Params           `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	Proposals          []Proposal       `protobuf:"bytes,2,rep,name=proposals,proto3" json:"proposals"`
	ECDSASignatures    []ECDSASignature `protobuf:"bytes,3,rep,name=ecdsa_signatures,json=ecdsaSignatures,proto3" json:"ecdsa_signatures"`
	StartingProposalID uint64           `protobuf:"varint,4,opt,name=starting_proposal_id,json=startingProposalId,proto3" json:"starting_proposal_id,omitempty"`
}

// NewGenesisState creates a new GenesisState instance
func NewGenesisState(params Params, proposals []Proposal, ecdsaSignatures []ECDSASignature, startingProposalID uint64) *GenesisState {
	return &GenesisState{
		Params:             params,
		Proposals:          proposals,
		ECDSASignatures:    ecdsaSignatures,
		StartingProposalID: startingProposalID,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:             DefaultParams(),
		Proposals:          []Proposal{},
		ECDSASignatures:    []ECDSASignature{},
		StartingProposalID: 1,
	}
}

// Validate performs genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, proposal := range gs.Proposals {
		if err := proposal.Validate(); err != nil {
			return err
		}
	}

	for _, signature := range gs.ECDSASignatures {
		if err := signature.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// DefaultGenesis returns the default genesis state for the module
func DefaultGenesis() *GenesisState {
	return DefaultGenesisState()
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(gs *GenesisState) error {
	return gs.Validate()
}

// Params defines the parameters for the governance module
type Params struct {
	VotingPeriod  uint64 `protobuf:"varint,1,opt,name=voting_period,json=votingPeriod,proto3" json:"voting_period,omitempty"`
	MinDeposit    string `protobuf:"bytes,2,opt,name=min_deposit,json=minDeposit,proto3" json:"min_deposit,omitempty"`
	Quorum        uint64 `protobuf:"varint,3,opt,name=quorum,proto3" json:"quorum,omitempty"`
	Threshold     uint64 `protobuf:"varint,4,opt,name=threshold,proto3" json:"threshold,omitempty"`
	VetoThreshold uint64 `protobuf:"varint,5,opt,name=veto_threshold,json=vetoThreshold,proto3" json:"veto_threshold,omitempty"`
}

// NewParams creates a new Params instance
func NewParams(votingPeriod uint64, minDeposit string, quorum, threshold, vetoThreshold uint64) Params {
	return Params{
		VotingPeriod:  votingPeriod,
		MinDeposit:    minDeposit,
		Quorum:        quorum,
		Threshold:     threshold,
		VetoThreshold: vetoThreshold,
	}
}

// ParamSetPairs implements the ParamSet interface
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(ParamStoreKey, &p.VotingPeriod, func(value interface{}) error { return nil }),
	}
}

// DefaultParams returns default governance parameters
func DefaultParams() Params {
	return NewParams(
		DefaultVotingPeriod,
		DefaultMinDeposit,
		DefaultQuorum,
		DefaultThreshold,
		DefaultVetoThreshold,
	)
}

// Validate validates the parameters
func (p Params) Validate() error {
	if p.VotingPeriod == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "voting_period must be greater than 0")
	}
	if len(p.MinDeposit) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "min_deposit cannot be empty")
	}
	if p.Quorum == 0 || p.Quorum > 100 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "quorum must be between 1 and 100")
	}
	if p.Threshold == 0 || p.Threshold > 100 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "threshold must be between 1 and 100")
	}
	if p.VetoThreshold == 0 || p.VetoThreshold > 100 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "veto_threshold must be between 1 and 100")
	}
	return nil
}

// MsgSubmitProposal defines a message for submitting a proposal with ECDSA signature
type MsgSubmitProposal struct {
	Title             string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description       string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Proposer          string `protobuf:"bytes,3,opt,name=proposer,proto3" json:"proposer,omitempty"`
	ProposerSignature string `protobuf:"bytes,4,opt,name=proposer_signature,json=proposerSignature,proto3" json:"proposer_signature,omitempty"`
	Nonce             uint64 `protobuf:"varint,5,opt,name=nonce,proto3" json:"nonce,omitempty"`
}

// NewMsgSubmitProposal creates a new MsgSubmitProposal instance
func NewMsgSubmitProposal(title, description, proposer, proposerSignature string, nonce uint64) *MsgSubmitProposal {
	return &MsgSubmitProposal{
		Title:             title,
		Description:       description,
		Proposer:          proposer,
		ProposerSignature: proposerSignature,
		Nonce:             nonce,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgSubmitProposal) Route() string { return RouterKey }

// Type implements the sdk.Msg interface
func (msg MsgSubmitProposal) Type() string { return "submit_proposal" }

// GetSigners implements the sdk.Msg interface
func (msg MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	// Convert Ethereum address to Cosmos address for signing
	// This is a simplified conversion - in production, you'd need proper address mapping
	ethAddr := common.HexToAddress(msg.Proposer)
	cosmosAddr := sdk.AccAddress(ethAddr.Bytes())
	return []sdk.AccAddress{cosmosAddr}
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgSubmitProposal) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgSubmitProposal) ValidateBasic() error {
	if len(msg.Title) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "title cannot be empty")
	}
	if len(msg.Description) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "description cannot be empty")
	}
	if !IsValidEthereumAddress(msg.Proposer) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid proposer address: %s", msg.Proposer)
	}
	if len(msg.ProposerSignature) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "proposer_signature cannot be empty")
	}
	return nil
}

// MsgSubmitProposalResponse defines the response for submitting a proposal
type MsgSubmitProposalResponse struct {
	ProposalID uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
}

// MsgVote defines a message for voting on a proposal with ECDSA signature
type MsgVote struct {
	ProposalID uint64     `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
	Voter      string     `protobuf:"bytes,2,opt,name=voter,proto3" json:"voter,omitempty"`
	Signature  string     `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	Nonce      uint64     `protobuf:"varint,4,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Option     VoteOption `protobuf:"varint,5,opt,name=option,proto3,enum=homechain.gov.v1.VoteOption" json:"option,omitempty"`
}

// NewMsgVote creates a new MsgVote instance
func NewMsgVote(proposalID uint64, voter, signature string, nonce uint64, option VoteOption) *MsgVote {
	return &MsgVote{
		ProposalID: proposalID,
		Voter:      voter,
		Signature:  signature,
		Nonce:      nonce,
		Option:     option,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgVote) Route() string { return RouterKey }

// Type implements the sdk.Msg interface
func (msg MsgVote) Type() string { return "vote" }

// GetSigners implements the sdk.Msg interface
func (msg MsgVote) GetSigners() []sdk.AccAddress {
	ethAddr := common.HexToAddress(msg.Voter)
	cosmosAddr := sdk.AccAddress(ethAddr.Bytes())
	return []sdk.AccAddress{cosmosAddr}
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgVote) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgVote) ValidateBasic() error {
	if msg.ProposalID == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "proposal_id must be greater than 0")
	}
	if !IsValidEthereumAddress(msg.Voter) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid voter address: %s", msg.Voter)
	}
	if len(msg.Signature) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "signature cannot be empty")
	}
	return nil
}

// MsgVoteResponse defines the response for voting
type MsgVoteResponse struct{}

// MsgDeposit defines a message for depositing on a proposal
type MsgDeposit struct {
	ProposalID uint64    `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
	Depositor  string    `protobuf:"bytes,2,opt,name=depositor,proto3" json:"depositor,omitempty"`
	Amount     sdk.Coins `protobuf:"bytes,3,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount,omitempty"`
}

// NewMsgDeposit creates a new MsgDeposit instance
func NewMsgDeposit(proposalID uint64, depositor string, amount sdk.Coins) *MsgDeposit {
	return &MsgDeposit{
		ProposalID: proposalID,
		Depositor:  depositor,
		Amount:     amount,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgDeposit) Route() string { return RouterKey }

// Type implements the sdk.Msg interface
func (msg MsgDeposit) Type() string { return "deposit" }

// GetSigners implements the sdk.Msg interface
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		// If it's an Ethereum address, convert it
		if IsValidEthereumAddress(msg.Depositor) {
			ethAddr := common.HexToAddress(msg.Depositor)
			depositor = sdk.AccAddress(ethAddr.Bytes())
		} else {
			panic(err)
		}
	}
	return []sdk.AccAddress{depositor}
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgDeposit) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgDeposit) ValidateBasic() error {
	if msg.ProposalID == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "proposal_id must be greater than 0")
	}
	if len(msg.Depositor) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "depositor cannot be empty")
	}
	if !msg.Amount.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount must be valid coins")
	}
	if msg.Amount.IsZero() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount cannot be zero")
	}
	return nil
}

// MsgDepositResponse defines the response for depositing
type MsgDepositResponse struct{}

// VerifyECDSASignature verifies an ECDSA signature
func VerifyECDSASignature(address, signature, message string) error {
	// Convert hex signature to bytes
	sigBytes, err := hex.DecodeString(signature[2:]) // Remove 0x prefix
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}

	if len(sigBytes) != 65 {
		return fmt.Errorf("invalid signature length: expected 65, got %d", len(sigBytes))
	}

	// Recover public key from signature
	hash := sha256.Sum256([]byte(message))
	pubKey, err := ethcrypto.SigToPub(hash[:], sigBytes)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get recovered address
	recoveredAddr := ethcrypto.PubkeyToAddress(*pubKey)

	// Compare with provided address
	providedAddr := common.HexToAddress(address)

	if recoveredAddr != providedAddr {
		return fmt.Errorf("signature verification failed: recovered address %s != provided address %s",
			recoveredAddr.Hex(), providedAddr.Hex())
	}

	return nil
}

// CreateEIP191Message creates an EIP-191 compliant message
func CreateEIP191Message(messageType, address string, params []string, nonce uint64) string {
	// EIP-191 prefix: "\x19Ethereum Signed Message:\n32"
	message := fmt.Sprintf("%s:%s:%s:%d", messageType, address, strings.Join(params, ":"), nonce)
	return fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
}

// MsgServer defines the message server interface for the gov module
type MsgServer interface {
	SubmitProposal(goCtx context.Context, msg *MsgSubmitProposal) (*MsgSubmitProposalResponse, error)
	Vote(goCtx context.Context, msg *MsgVote) (*MsgVoteResponse, error)
	Deposit(goCtx context.Context, msg *MsgDeposit) (*MsgDepositResponse, error)
}

// QueryServer defines the query server interface for the gov module
type QueryServer interface {
	Proposal(goCtx context.Context, req *QueryProposalRequest) (*QueryProposalResponse, error)
	Proposals(goCtx context.Context, req *QueryProposalsRequest) (*QueryProposalsResponse, error)
	Vote(goCtx context.Context, req *QueryVoteRequest) (*QueryVoteResponse, error)
	Votes(goCtx context.Context, req *QueryVotesRequest) (*QueryVotesResponse, error)
	Deposit(goCtx context.Context, req *QueryDepositRequest) (*QueryDepositResponse, error)
	Deposits(goCtx context.Context, req *QueryDepositsRequest) (*QueryDepositsResponse, error)
	Params(goCtx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
}

// Query request/response types for gRPC queries

// QueryProposalRequest is the request type for the Query/Proposal RPC method
type QueryProposalRequest struct {
	ProposalId uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
}

// QueryProposalResponse is the response type for the Query/Proposal RPC method
type QueryProposalResponse struct {
	Proposal Proposal `protobuf:"bytes,1,opt,name=proposal,proto3" json:"proposal"`
}

// QueryProposalsRequest is the request type for the Query/Proposals RPC method
type QueryProposalsRequest struct{}

// QueryProposalsResponse is the response type for the Query/Proposals RPC method
type QueryProposalsResponse struct {
	Proposals []Proposal `protobuf:"bytes,1,rep,name=proposals,proto3" json:"proposals"`
}

// QueryVoteRequest is the request type for the Query/Vote RPC method
type QueryVoteRequest struct {
	ProposalId uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
	Voter      string `protobuf:"bytes,2,opt,name=voter,proto3" json:"voter,omitempty"`
}

// QueryVoteResponse is the response type for the Query/Vote RPC method
type QueryVoteResponse struct {
	Vote Vote `protobuf:"bytes,1,opt,name=vote,proto3" json:"vote"`
}

// QueryVotesRequest is the request type for the Query/Votes RPC method
type QueryVotesRequest struct {
	ProposalId uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
}

// QueryVotesResponse is the response type for the Query/Votes RPC method
type QueryVotesResponse struct {
	Votes []Vote `protobuf:"bytes,1,rep,name=votes,proto3" json:"votes"`
}

// QueryDepositRequest is the request type for the Query/Deposit RPC method
type QueryDepositRequest struct {
	ProposalId uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
	Depositor  string `protobuf:"bytes,2,opt,name=depositor,proto3" json:"depositor,omitempty"`
}

// QueryDepositResponse is the response type for the Query/Deposit RPC method
type QueryDepositResponse struct {
	Deposit Deposit `protobuf:"bytes,1,opt,name=deposit,proto3" json:"deposit"`
}

// QueryDepositsRequest is the request type for the Query/Deposits RPC method
type QueryDepositsRequest struct {
	ProposalId uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
}

// QueryDepositsResponse is the response type for the Query/Deposits RPC method
type QueryDepositsResponse struct {
	Deposits []Deposit `protobuf:"bytes,1,rep,name=deposits,proto3" json:"deposits"`
}

// QueryParamsRequest is the request type for the Query/Params RPC method
type QueryParamsRequest struct{}

// QueryParamsResponse is the response type for the Query/Params RPC method
type QueryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

// ParamStoreKey defines the key for governance parameters in the paramspace
var ParamStoreKey = []byte("Params")

// ParamKeyTable defines the param key table for the gov module
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// RegisterInterfaces registers the interfaces for the module
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitProposal{},
		&MsgVote{},
		&MsgDeposit{})
}

// RegisterLegacyAminoCodec registers the amino codec for the module
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitProposal{}, "homechain/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(&MsgVote{}, "homechain/MsgVote", nil)
	cdc.RegisterConcrete(&MsgDeposit{}, "homechain/MsgDeposit", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc = codec.NewLegacyAmino()

// ProtoMessage stub implementations for codec compatibility
func (p *Proposal) ProtoMessage() {}
func (p *Proposal) Reset()        { *p = Proposal{} }
func (p Proposal) String() string { return p.Title }

func (v *Vote) ProtoMessage() {}
func (v *Vote) Reset()        { *v = Vote{} }
func (v Vote) String() string { return v.Voter }

func (d *Deposit) ProtoMessage() {}
func (d *Deposit) Reset()        { *d = Deposit{} }
func (d Deposit) String() string { return d.Depositor }

func (p *Params) ProtoMessage() {}
func (p *Params) Reset()        { *p = Params{} }
func (p Params) String() string { return "Params" }

func (m *MsgSubmitProposal) ProtoMessage() {}
func (m *MsgSubmitProposal) Reset()        { *m = MsgSubmitProposal{} }
func (m MsgSubmitProposal) String() string { return m.Title }

func (m *MsgSubmitProposalResponse) ProtoMessage() {}
func (m *MsgSubmitProposalResponse) Reset()        { *m = MsgSubmitProposalResponse{} }
func (m MsgSubmitProposalResponse) String() string { return "MsgSubmitProposalResponse" }

func (m *MsgVote) ProtoMessage() {}
func (m *MsgVote) Reset()        { *m = MsgVote{} }
func (m MsgVote) String() string { return m.Voter }

func (m *MsgVoteResponse) ProtoMessage() {}
func (m *MsgVoteResponse) Reset()        { *m = MsgVoteResponse{} }
func (m MsgVoteResponse) String() string { return "MsgVoteResponse" }

func (m *MsgDeposit) ProtoMessage() {}
func (m *MsgDeposit) Reset()        { *m = MsgDeposit{} }
func (m MsgDeposit) String() string { return m.Depositor }

func (m *MsgDepositResponse) ProtoMessage() {}
func (m *MsgDepositResponse) Reset()        { *m = MsgDepositResponse{} }
func (m MsgDepositResponse) String() string { return "MsgDepositResponse" }

func (q *QueryProposalRequest) ProtoMessage() {}
func (q *QueryProposalRequest) Reset()        { *q = QueryProposalRequest{} }
func (q QueryProposalRequest) String() string { return "QueryProposalRequest" }

func (q *QueryProposalResponse) ProtoMessage() {}
func (q *QueryProposalResponse) Reset()        { *q = QueryProposalResponse{} }
func (q QueryProposalResponse) String() string { return "QueryProposalResponse" }

func (q *QueryProposalsRequest) ProtoMessage() {}
func (q *QueryProposalsRequest) Reset()        { *q = QueryProposalsRequest{} }
func (q QueryProposalsRequest) String() string { return "QueryProposalsRequest" }

func (q *QueryProposalsResponse) ProtoMessage() {}
func (q *QueryProposalsResponse) Reset()        { *q = QueryProposalsResponse{} }
func (q QueryProposalsResponse) String() string { return "QueryProposalsResponse" }

func (q *QueryVoteRequest) ProtoMessage() {}
func (q *QueryVoteRequest) Reset()        { *q = QueryVoteRequest{} }
func (q QueryVoteRequest) String() string { return "QueryVoteRequest" }

func (q *QueryVoteResponse) ProtoMessage() {}
func (q *QueryVoteResponse) Reset()        { *q = QueryVoteResponse{} }
func (q QueryVoteResponse) String() string { return "QueryVoteResponse" }

func (q *QueryVotesRequest) ProtoMessage() {}
func (q *QueryVotesRequest) Reset()        { *q = QueryVotesRequest{} }
func (q QueryVotesRequest) String() string { return "QueryVotesRequest" }

func (q *QueryVotesResponse) ProtoMessage() {}
func (q *QueryVotesResponse) Reset()        { *q = QueryVotesResponse{} }
func (q QueryVotesResponse) String() string { return "QueryVotesResponse" }

func (q *QueryDepositRequest) ProtoMessage() {}
func (q *QueryDepositRequest) Reset()        { *q = QueryDepositRequest{} }
func (q QueryDepositRequest) String() string { return "QueryDepositRequest" }

func (q *QueryDepositResponse) ProtoMessage() {}
func (q *QueryDepositResponse) Reset()        { *q = QueryDepositResponse{} }
func (q QueryDepositResponse) String() string { return "QueryDepositResponse" }

func (q *QueryDepositsRequest) ProtoMessage() {}
func (q *QueryDepositsRequest) Reset()        { *q = QueryDepositsRequest{} }
func (q QueryDepositsRequest) String() string { return "QueryDepositsRequest" }

func (q *QueryDepositsResponse) ProtoMessage() {}
func (q *QueryDepositsResponse) Reset()        { *q = QueryDepositsResponse{} }
func (q QueryDepositsResponse) String() string { return "QueryDepositsResponse" }

func (q *QueryParamsRequest) ProtoMessage() {}
func (q *QueryParamsRequest) Reset()        { *q = QueryParamsRequest{} }
func (q QueryParamsRequest) String() string { return "QueryParamsRequest" }

func (q *QueryParamsResponse) ProtoMessage() {}
func (q *QueryParamsResponse) Reset()        { *q = QueryParamsResponse{} }
func (q QueryParamsResponse) String() string { return "QueryParamsResponse" }

func (gs *GenesisState) ProtoMessage() {}
func (gs *GenesisState) Reset()        { *gs = GenesisState{} }
func (gs GenesisState) String() string { return "GenesisState" }

// RegisterMsgServer registers the message server
func RegisterMsgServer(server interface{}, msgServer MsgServer) {}

// RegisterQueryServer registers the query server
func RegisterQueryServer(server interface{}, queryServer QueryServer) {}
