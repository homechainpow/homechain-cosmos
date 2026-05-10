package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/homechain/homechain/x/mining/types"
)

// MsgServer implements the MsgServer interface for the mining module
type MsgServer struct {
	Keeper
}

// NewMsgServerImpl creates a new MsgServer instance
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &MsgServer{Keeper: keeper}
}

// RegisterMiner handles the registration of a new miner
func (k MsgServer) RegisterMiner(goCtx context.Context, msg *types.MsgRegisterMiner) (*types.MsgRegisterMinerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate signer
	_, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid signer address: %w", err)
	}

	// Verify signature (in a real implementation, this would verify the signature)
	// For now, we'll skip signature verification for simplicity
	// if err := k.VerifySignature(ctx, msg); err != nil {
	//     return nil, err
	// }

	// Register miner
	if err := k.Keeper.RegisterMiner(ctx, msg.Address, msg.DeviceInfo, msg.Referrer); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMinerRegistered,
			sdk.NewAttribute(types.AttributeKeyMiner, msg.Address),
			sdk.NewAttribute(types.AttributeKeyDeviceInfo, msg.DeviceInfo),
			sdk.NewAttribute(types.AttributeKeyReferrer, msg.Referrer),
		),
	)

	return &types.MsgRegisterMinerResponse{}, nil
}

// SubmitShare handles the submission of a mining share
func (k MsgServer) SubmitShare(goCtx context.Context, msg *types.MsgSubmitShare) (*types.MsgSubmitShareResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate signer
	_, err := sdk.AccAddressFromBech32(msg.Miner)
	if err != nil {
		return nil, fmt.Errorf("invalid miner address: %w", err)
	}

	// Verify signature (in a real implementation, this would verify the signature)
	// For now, we'll skip signature verification for simplicity
	// if err := k.VerifySignature(ctx, msg); err != nil {
	//     return nil, err
	// }

	// Submit share
	if err := k.Keeper.SubmitShare(ctx, msg.ShareData); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeShareSubmitted,
			sdk.NewAttribute(types.AttributeKeyMiner, msg.Miner),
			sdk.NewAttribute(types.AttributeKeyShareHash, msg.ShareData.Hash),
			sdk.NewAttribute(types.AttributeKeyDifficulty, fmt.Sprintf("%d", msg.ShareData.Difficulty)),
		),
	)

	return &types.MsgSubmitShareResponse{
		Accepted: true,
		Reason:   "Share accepted",
	}, nil
}

// MinerHeartbeat handles miner heartbeat
func (k MsgServer) MinerHeartbeat(goCtx context.Context, msg *types.MsgMinerHeartbeat) (*types.MsgMinerHeartbeatResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate signer
	_, err := sdk.AccAddressFromBech32(msg.Miner)
	if err != nil {
		return nil, fmt.Errorf("invalid miner address: %w", err)
	}

	// Verify signature (in a real implementation, this would verify the signature)
	// For now, we'll skip signature verification for simplicity
	// if err := k.VerifySignature(ctx, msg); err != nil {
	//     return nil, err
	// }

	// Update miner heartbeat
	if err := k.Keeper.MinerHeartbeat(ctx, msg.Miner); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMinerHeartbeat,
			sdk.NewAttribute(types.AttributeKeyMiner, msg.Miner),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", msg.Timestamp)),
		),
	)

	return &types.MsgMinerHeartbeatResponse{}, nil
}
