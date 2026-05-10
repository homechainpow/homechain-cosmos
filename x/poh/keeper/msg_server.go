package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/homechain/homechain/x/poh/types"
)

// MsgServer implements the MsgServer interface for the poh module
type MsgServer struct {
	Keeper
}

// NewMsgServerImpl creates a new MsgServer instance
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &MsgServer{Keeper: keeper}
}

// SubmitPoH handles the submission of PoH data
func (k MsgServer) SubmitPoH(goCtx context.Context, msg *types.MsgSubmitPoH) (*types.MsgSubmitPoHResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Consume gas for Argon2 PoH verification to prevent CPU exhaustion attacks
	k.Keeper.ConsumePoHGas(ctx)

	// Validate signer
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, fmt.Errorf("invalid signer address: %s", err)
	}

	// Submit PoH data
	if err := k.Keeper.SubmitPoH(ctx, msg.PohData); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePoHSubmitted,
			sdk.NewAttribute(types.AttributeKeySigner, msg.Signer),
			sdk.NewAttribute(types.AttributeKeyPrevHash, msg.PohData.PrevHash),
			sdk.NewAttribute(types.AttributeKeyNewHash, msg.PohData.NewHash),
			sdk.NewAttribute(types.AttributeKeyDifficulty, fmt.Sprintf("%d", msg.PohData.Difficulty)),
		),
	)

	return &types.MsgSubmitPoHResponse{}, nil
}
