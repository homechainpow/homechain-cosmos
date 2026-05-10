package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/homechain/homechain/x/poh/types"
)

// QueryServer implements the QueryServer interface for the poh module
type QueryServer struct {
	Keeper
}

// NewQueryServerImpl creates a new QueryServer instance
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &QueryServer{Keeper: keeper}
}

// CurrentPoH returns the current PoH data
func (k QueryServer) CurrentPoH(goCtx context.Context, req *types.QueryCurrentPoHRequest) (*types.QueryCurrentPoHResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pohData, err := k.Keeper.GetCurrentPoH(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryCurrentPoHResponse{
		PohData: pohData,
	}, nil
}

// PoHSequence returns PoH data for a specific height
func (k QueryServer) PoHSequence(goCtx context.Context, req *types.QueryPoHSequenceRequest) (*types.QueryPoHSequenceResponse, error) {
	if req.Height == 0 {
		return nil, status.Error(codes.InvalidArgument, "height cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pohData, err := k.Keeper.GetPoHSequence(ctx, req.Height)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryPoHSequenceResponse{
		PohData: pohData,
	}, nil
}
