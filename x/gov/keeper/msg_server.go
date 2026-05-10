package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/homechain/homechain/x/gov/types"
)

// MsgServer implements the MsgServer interface for the gov module
type MsgServer struct {
	Keeper
}

// NewMsgServerImpl creates a new MsgServer instance
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &MsgServer{Keeper: keeper}
}

// SubmitProposal handles the submission of a new proposal
func (k MsgServer) SubmitProposal(goCtx context.Context, msg *types.MsgSubmitProposal) (*types.MsgSubmitProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Submit proposal with ECDSA signature verification
	proposalID, err := k.Keeper.SubmitProposal(ctx, msg.Title, msg.Description, msg.Proposer, msg.ProposerSignature, msg.Nonce)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalSubmitted,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute(types.AttributeKeyProposer, msg.Proposer),
			sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
		),
	)

	return &types.MsgSubmitProposalResponse{
		ProposalID: proposalID,
	}, nil
}

// Vote handles voting on a proposal
func (k MsgServer) Vote(goCtx context.Context, msg *types.MsgVote) (*types.MsgVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Vote on proposal with ECDSA signature verification
	if err := k.Keeper.Vote(ctx, msg.ProposalID, msg.Voter, msg.Signature, msg.Nonce, msg.Option); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVoteCast,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalID)),
			sdk.NewAttribute(types.AttributeKeyVoter, msg.Voter),
			sdk.NewAttribute(types.AttributeKeyOption, msg.Option.String()),
		),
	)

	return &types.MsgVoteResponse{}, nil
}

// Deposit handles depositing on a proposal
func (k MsgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Deposit on proposal
	if err := k.Keeper.Deposit(ctx, msg.ProposalID, msg.Depositor, msg.Amount); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDepositMade,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalID)),
			sdk.NewAttribute(types.AttributeKeyDepositor, msg.Depositor),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
		),
	)

	return &types.MsgDepositResponse{}, nil
}
