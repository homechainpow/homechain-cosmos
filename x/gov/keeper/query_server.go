package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/homechain/homechain/x/gov/types"
)

// QueryServer implements the QueryServer interface for the gov module
type QueryServer struct {
	Keeper
}

// NewQueryServerImpl creates a new QueryServer instance
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &QueryServer{Keeper: keeper}
}

// Proposal returns proposal information for a specific proposal ID
func (k QueryServer) Proposal(goCtx context.Context, req *types.QueryProposalRequest) (*types.QueryProposalResponse, error) {
	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal_id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	proposal, err := k.Keeper.GetProposal(ctx, req.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryProposalResponse{
		Proposal: proposal,
	}, nil
}

// Proposals returns all proposals
func (k QueryServer) Proposals(goCtx context.Context, req *types.QueryProposalsRequest) (*types.QueryProposalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	proposals := k.Keeper.GetProposals(ctx)

	return &types.QueryProposalsResponse{
		Proposals: proposals,
	}, nil
}

// Vote returns vote information for a specific proposal and voter
func (k QueryServer) Vote(goCtx context.Context, req *types.QueryVoteRequest) (*types.QueryVoteResponse, error) {
	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal_id cannot be 0")
	}
	if req.Voter == "" {
		return nil, status.Error(codes.InvalidArgument, "voter cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	proposal, err := k.Keeper.GetProposal(ctx, req.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Find vote for the specified voter
	for _, vote := range proposal.Votes {
		if vote.Voter == req.Voter {
			return &types.QueryVoteResponse{
				Vote: vote,
			}, nil
		}
	}

	return nil, status.Error(codes.NotFound, "vote not found")
}

// Votes returns all votes for a specific proposal
func (k QueryServer) Votes(goCtx context.Context, req *types.QueryVotesRequest) (*types.QueryVotesResponse, error) {
	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal_id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	proposal, err := k.Keeper.GetProposal(ctx, req.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryVotesResponse{
		Votes: proposal.Votes,
	}, nil
}

// Deposit returns deposit information for a specific proposal and depositor
func (k QueryServer) Deposit(goCtx context.Context, req *types.QueryDepositRequest) (*types.QueryDepositResponse, error) {
	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal_id cannot be 0")
	}
	if req.Depositor == "" {
		return nil, status.Error(codes.InvalidArgument, "depositor cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := k.Keeper.GetProposal(ctx, req.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Find deposit for the specified depositor
	// Note: In a real implementation, you would store deposits separately
	// For now, we'll return an empty deposit as the deposit logic is simplified
	return &types.QueryDepositResponse{
		Deposit: types.Deposit{
			ProposalID: req.ProposalId,
			Depositor:  req.Depositor,
			Amount:     sdk.NewCoins(),
			Timestamp:  uint64(ctx.BlockTime().Unix()),
		},
	}, nil
}

// Deposits returns all deposits for a specific proposal
func (k QueryServer) Deposits(goCtx context.Context, req *types.QueryDepositsRequest) (*types.QueryDepositsResponse, error) {
	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal_id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := k.Keeper.GetProposal(ctx, req.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Note: In a real implementation, you would store deposits separately
	// For now, we'll return an empty list as the deposit logic is simplified
	return &types.QueryDepositsResponse{
		Deposits: []types.Deposit{},
	}, nil
}

// Params returns the governance parameters
func (k QueryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
