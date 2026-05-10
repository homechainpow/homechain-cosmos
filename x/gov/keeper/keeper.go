package keeper

import (
	"encoding/binary"
	"fmt"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/homechain/homechain/x/gov/types"
)

// prefixKey adds a prefix to a key
func prefixKey(prefix, key []byte) []byte {
	return append(prefix, key...)
}

// Keeper of the governance store
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramSpace paramstypes.Subspace
	bankKeeper bankkeeper.Keeper
}

// NewKeeper creates a new governance Keeper instance
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, paramSpace paramstypes.Subspace, bankKeeper bankkeeper.Keeper) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramSpace: paramSpace,
		bankKeeper: bankKeeper,
	}
}

// GetProposal retrieves a proposal from the store
func (k Keeper) GetProposal(ctx sdk.Context, proposalID uint64) (types.Proposal, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ProposalKey(proposalID))
	if bz == nil {
		return types.Proposal{}, errorsmod.Wrap(sdkerrors.ErrNotFound, "proposal not found")
	}

	var proposal types.Proposal
	k.cdc.MustUnmarshal(bz, &proposal)
	return proposal, nil
}

// SetProposal stores a proposal in the store
func (k Keeper) SetProposal(ctx sdk.Context, proposal types.Proposal) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ProposalKey(proposal.ProposalID), k.cdc.MustMarshal(&proposal))
}

// HasProposal checks if a proposal exists
func (k Keeper) HasProposal(ctx sdk.Context, proposalID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ProposalKey(proposalID))
}

// GetProposals returns all proposals
func (k Keeper) GetProposals(ctx sdk.Context) []types.Proposal {
	var proposals []types.Proposal
	store := ctx.KVStore(k.storeKey)

	iterator := store.Iterator(prefixKey([]byte(types.ProposalKeyPrefix), nil), nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var proposal types.Proposal
		k.cdc.MustUnmarshal(iterator.Value(), &proposal)
		proposals = append(proposals, proposal)
	}

	return proposals
}

// GetNextProposalID returns the next proposal ID
func (k Keeper) GetNextProposalID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get([]byte("next_proposal_id"))
	if bz == nil {
		return 1
	}

	return binary.BigEndian.Uint64(bz)
}

// SetNextProposalID sets the next proposal ID
func (k Keeper) SetNextProposalID(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, proposalID)
	store.Set([]byte("next_proposal_id"), bz)
}

// SubmitProposal submits a new proposal with ECDSA signature verification
func (k Keeper) SubmitProposal(ctx sdk.Context, title, description, proposer, proposerSignature string, nonce uint64) (uint64, error) {
	// Validate Ethereum address
	if !types.IsValidEthereumAddress(proposer) {
		return 0, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid proposer address")
	}

	// Create EIP-191 message for verification
	message := types.CreateEIP191Message("submit-proposal", proposer, []string{title, description}, nonce)

	// Verify ECDSA signature
	if err := types.VerifyECDSASignature(proposer, proposerSignature, message); err != nil {
		return 0, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("signature verification failed: %s", err))
	}

	// Check for nonce replay
	if k.HasECDSASignature(ctx, proposer, nonce) {
		return 0, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "nonce already used")
	}

	// Get next proposal ID
	proposalID := k.GetNextProposalID(ctx)

	// Calculate timing
	submitTime := uint64(ctx.BlockTime().Unix())
	params := k.GetParams(ctx)
	depositEndTime := submitTime + (params.VotingPeriod / 4) // Deposit period is 1/4 of voting period
	votingStartTime := depositEndTime
	votingEndTime := votingStartTime + params.VotingPeriod

	// Create proposal
	proposal := types.NewProposal(
		proposalID,
		title,
		description,
		proposer,
		proposerSignature,
		nonce,
		submitTime,
		depositEndTime,
		votingStartTime,
		votingEndTime,
	)

	// Store proposal
	k.SetProposal(ctx, proposal)

	// Store ECDSA signature to prevent replay
	k.SetECDSASignature(ctx, types.NewECDSASignature(proposer, proposerSignature, message, nonce, submitTime))

	// Update next proposal ID
	k.SetNextProposalID(ctx, proposalID+1)

	// Log successful proposal submission
	ctx.Logger().Info("Proposal submitted successfully",
		"proposal_id", proposalID,
		"proposer", proposer,
		"title", title)

	return proposalID, nil
}

// Vote on a proposal with ECDSA signature verification
func (k Keeper) Vote(ctx sdk.Context, proposalID uint64, voter, signature string, nonce uint64, option types.VoteOption) error {
	// Validate Ethereum address
	if !types.IsValidEthereumAddress(voter) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid voter address")
	}

	// Get proposal
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	// Check if proposal is in voting period
	if proposal.Status != types.ProposalStatusVotingPeriod {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "proposal is not in voting period")
	}

	// Check if voting period has ended
	if uint64(ctx.BlockTime().Unix()) > proposal.VotingEndTime {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "voting period has ended")
	}

	// Check if already voted
	for _, vote := range proposal.Votes {
		if vote.Voter == voter {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "already voted on this proposal")
		}
	}

	// Create EIP-191 message for verification
	message := types.CreateEIP191Message("vote", voter, []string{strconv.FormatUint(proposalID, 10), option.String()}, nonce)

	// Verify ECDSA signature
	if err := types.VerifyECDSASignature(voter, signature, message); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("signature verification failed: %s", err))
	}

	// Check for nonce replay
	if k.HasECDSASignature(ctx, voter, nonce) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "nonce already used")
	}

	// Create vote
	vote := types.NewVote(proposalID, voter, signature, nonce, option, uint64(ctx.BlockTime().Unix()))

	// Add vote to proposal
	proposal.Votes = append(proposal.Votes, vote)

	// Update proposal
	k.SetProposal(ctx, proposal)

	// Store ECDSA signature to prevent replay
	k.SetECDSASignature(ctx, types.NewECDSASignature(voter, signature, message, nonce, uint64(ctx.BlockTime().Unix())))

	// Log successful vote
	ctx.Logger().Info("Vote cast successfully",
		"proposal_id", proposalID,
		"voter", voter,
		"option", option.String())

	return nil
}

// Deposit on a proposal
func (k Keeper) Deposit(ctx sdk.Context, proposalID uint64, depositor string, amount sdk.Coins) error {
	// Validate depositor address
	if !types.IsValidEthereumAddress(depositor) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid depositor address")
	}

	// Get proposal
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	// Check if proposal is in deposit period
	if proposal.Status != types.ProposalStatusDepositPeriod {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "proposal is not in deposit period")
	}

	// Check if deposit period has ended
	if uint64(ctx.BlockTime().Unix()) > proposal.DepositEndTime {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "deposit period has ended")
	}

	// Parse minimum deposit
	params := k.GetParams(ctx)
	minDeposit, err := sdk.ParseCoinsNormalized(params.MinDeposit)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("invalid min_deposit: %s", err))
	}

	// Check if total deposit meets minimum
	newTotal := proposal.TotalDeposit.Add(amount...)
	if !newTotal.IsAllGTE(minDeposit) {
		return errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "total deposit does not meet minimum requirement")
	}

	// Create deposit (unused for now)
	_ = types.NewDeposit(proposalID, depositor, amount, uint64(ctx.BlockTime().Unix()))

	// Update proposal total deposit
	proposal.TotalDeposit = newTotal

	// Check if minimum deposit is met to move to voting period
	if proposal.TotalDeposit.IsAllGTE(minDeposit) {
		proposal.Status = types.ProposalStatusVotingPeriod
		proposal.VotingStartTime = uint64(ctx.BlockTime().Unix())
		proposal.VotingEndTime = proposal.VotingStartTime + params.VotingPeriod
	}

	// Update proposal
	k.SetProposal(ctx, proposal)

	// Log successful deposit
	ctx.Logger().Info("Deposit made successfully",
		"proposal_id", proposalID,
		"depositor", depositor,
		"amount", amount.String(),
		"total_deposit", proposal.TotalDeposit.String())

	return nil
}

// GetECDSASignature retrieves an ECDSA signature from the store
func (k Keeper) GetECDSASignature(ctx sdk.Context, address string, nonce uint64) (types.ECDSASignature, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ECDSASignatureKey(address, nonce))
	if bz == nil {
		return types.ECDSASignature{}, errorsmod.Wrap(sdkerrors.ErrNotFound, "ECDSA signature not found")
	}

	var signature types.ECDSASignature
	k.cdc.MustUnmarshal(bz, &signature)
	return signature, nil
}

// SetECDSASignature stores an ECDSA signature in the store
func (k Keeper) SetECDSASignature(ctx sdk.Context, signature types.ECDSASignature) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ECDSASignatureKey(signature.Address, signature.Nonce), k.cdc.MustMarshal(&signature))
}

// HasECDSASignature checks if an ECDSA signature exists
func (k Keeper) HasECDSASignature(ctx sdk.Context, address string, nonce uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ECDSASignatureKey(address, nonce))
}

// ProcessExpiredProposals processes expired proposals and updates their status
func (k Keeper) ProcessExpiredProposals(ctx sdk.Context) {
	proposals := k.GetProposals(ctx)
	currentTime := uint64(ctx.BlockTime().Unix())

	for _, proposal := range proposals {
		switch proposal.Status {
		case types.ProposalStatusDepositPeriod:
			if currentTime > proposal.DepositEndTime {
				// Deposit period expired, reject proposal
				proposal.Status = types.ProposalStatusRejected
				k.SetProposal(ctx, proposal)
				ctx.Logger().Info("Proposal rejected due to expired deposit period",
					"proposal_id", proposal.ProposalID)
			}

		case types.ProposalStatusVotingPeriod:
			if currentTime > proposal.VotingEndTime {
				// Voting period expired, tally votes
				k.TallyVotes(ctx, proposal)
			}
		}
	}
}

// TallyVotes tallies votes and updates proposal status
func (k Keeper) TallyVotes(ctx sdk.Context, proposal types.Proposal) {
	params := k.GetParams(ctx)

	// Count votes
	yesVotes := uint64(0)
	noVotes := uint64(0)
	abstainVotes := uint64(0)
	noWithVetoVotes := uint64(0)

	for _, vote := range proposal.Votes {
		switch vote.Option {
		case types.VoteOptionYes:
			yesVotes++
		case types.VoteOptionNo:
			noVotes++
		case types.VoteOptionAbstain:
			abstainVotes++
		case types.VoteOptionNoWithVeto:
			noWithVetoVotes++
		}
	}

	totalVotes := yesVotes + noVotes + abstainVotes + noWithVetoVotes

	// Calculate quorum (assuming total voting power is known)
	// For simplicity, we'll use a fixed voting power
	totalVotingPower := uint64(1000000) // This should be calculated from actual staking
	quorumThreshold := totalVotingPower * params.Quorum / 100

	if totalVotes < quorumThreshold {
		// Quorum not met, reject proposal
		proposal.Status = types.ProposalStatusRejected
		k.SetProposal(ctx, proposal)
		ctx.Logger().Info("Proposal rejected due to insufficient quorum",
			"proposal_id", proposal.ProposalID,
			"votes", totalVotes,
			"quorum_threshold", quorumThreshold)
		return
	}

	// Check veto threshold
	vetoThreshold := totalVotingPower * params.VetoThreshold / 100
	if noWithVetoVotes > vetoThreshold {
		// Veto threshold met, reject proposal
		proposal.Status = types.ProposalStatusRejected
		k.SetProposal(ctx, proposal)
		ctx.Logger().Info("Proposal rejected due to veto threshold",
			"proposal_id", proposal.ProposalID,
			"no_with_veto_votes", noWithVetoVotes,
			"veto_threshold", vetoThreshold)
		return
	}

	// Check if proposal passes
	yesThreshold := totalVotingPower * params.Threshold / 100
	if yesVotes > yesThreshold {
		proposal.Status = types.ProposalStatusPassed
		ctx.Logger().Info("Proposal passed",
			"proposal_id", proposal.ProposalID,
			"yes_votes", yesVotes,
			"threshold", yesThreshold)
	} else {
		proposal.Status = types.ProposalStatusRejected
		ctx.Logger().Info("Proposal rejected due to insufficient yes votes",
			"proposal_id", proposal.ProposalID,
			"yes_votes", yesVotes,
			"threshold", yesThreshold)
	}

	k.SetProposal(ctx, proposal)
}

// GetParams returns the total set of governance parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.Get(ctx, types.ParamStoreKey, &params)
	return params
}

// SetParams sets the governance parameters to the paramspace.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.Set(ctx, types.ParamStoreKey, params)
}

// ValidateGenesis validates the genesis state
func (k Keeper) ValidateGenesis(ctx sdk.Context, genesisState *types.GenesisState) error {
	if err := genesisState.Validate(); err != nil {
		return err
	}

	// Validate proposals
	for _, proposal := range genesisState.Proposals {
		if err := proposal.Validate(); err != nil {
			return err
		}
	}

	// Validate ECDSA signatures
	for _, signature := range genesisState.ECDSASignatures {
		if err := signature.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// InitGenesis initializes the genesis state
func (k Keeper) InitGenesis(ctx sdk.Context, genesisState *types.GenesisState) {
	if err := k.ValidateGenesis(ctx, genesisState); err != nil {
		panic(err)
	}

	// Set parameters
	k.SetParams(ctx, genesisState.Params)

	// Initialize proposals
	for _, proposal := range genesisState.Proposals {
		k.SetProposal(ctx, proposal)
	}

	// Initialize ECDSA signatures
	for _, signature := range genesisState.ECDSASignatures {
		k.SetECDSASignature(ctx, signature)
	}

	// Set starting proposal ID
	k.SetNextProposalID(ctx, genesisState.StartingProposalID)
}

// ExportGenesis exports the genesis state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	proposals := k.GetProposals(ctx)

	// Get ECDSA signatures
	var ecdsaSignatures []types.ECDSASignature
	store := ctx.KVStore(k.storeKey)

	iterator := store.Iterator(prefixKey([]byte(types.ECDSASignatureKeyPrefix), nil), nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var signature types.ECDSASignature
		k.cdc.MustUnmarshal(iterator.Value(), &signature)
		ecdsaSignatures = append(ecdsaSignatures, signature)
	}

	return types.NewGenesisState(params, proposals, ecdsaSignatures, k.GetNextProposalID(ctx))
}
