package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/homechain/homechain/x/nodestake/types"
)

// MsgServer implements the MsgServer interface for the nodestake module
type MsgServer struct {
	Keeper
}

// NewMsgServerImpl creates a new MsgServer instance
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &MsgServer{Keeper: keeper}
}

// Stake handles staking tokens on a node
func (k MsgServer) Stake(goCtx context.Context, msg *types.MsgStake) (*types.MsgStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate address
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Validate boost
	if err := k.Keeper.ValidateBoost(msg.Boost); err != nil {
		return nil, err
	}

	// Transfer coins from user to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, msg.Amount); err != nil {
		return nil, fmt.Errorf("failed to transfer stake: %w", err)
	}

	// Get existing stake or create new
	stake, found := k.Keeper.GetStake(ctx, msg.Address)
	if found {
		stake.Amount = stake.Amount.Add(msg.Amount...)
		stake.Boost = msg.Boost
	} else {
		stake = types.StakeInfo{
			Address: msg.Address,
			Amount:  msg.Amount,
			Boost:   msg.Boost,
		}
	}

	// Store stake info
	k.Keeper.SetStake(ctx, stake)

	ctx.Logger().Info("Stake successful",
		"address", msg.Address,
		"amount", msg.Amount.String(),
		"boost", msg.Boost,
	)

	return &types.MsgStakeResponse{}, nil
}

// Unstake handles unstaking tokens from a node
func (k MsgServer) Unstake(goCtx context.Context, msg *types.MsgUnstake) (*types.MsgUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate address
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Parse amount
	amount, err := sdk.ParseCoinsNormalized(msg.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	// Check stake exists
	stake, found := k.Keeper.GetStake(ctx, msg.Address)
	if !found {
		return nil, fmt.Errorf("no stake found for address %s", msg.Address)
	}

	// Validate unstake amount does not exceed staked amount
	if !stake.Amount.IsAllGTE(amount) {
		return nil, fmt.Errorf("insufficient staked amount: has %s, wants %s", stake.Amount.String(), amount.String())
	}

	// Transfer coins from module account back to user
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, amount); err != nil {
		return nil, fmt.Errorf("failed to return unstaked coins: %w", err)
	}

	// Update stake info
	stake.Amount = stake.Amount.Sub(amount...)
	if stake.Amount.IsZero() {
		k.Keeper.RemoveStake(ctx, msg.Address)
	} else {
		k.Keeper.SetStake(ctx, stake)
	}

	ctx.Logger().Info("Unstake successful",
		"address", msg.Address,
		"amount", amount.String(),
	)

	return &types.MsgUnstakeResponse{}, nil
}

// UpdateBoost handles updating the boost multiplier
func (k MsgServer) UpdateBoost(goCtx context.Context, msg *types.MsgUpdateBoost) (*types.MsgUpdateBoostResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate boost
	if err := k.Keeper.ValidateBoost(msg.Boost); err != nil {
		return nil, err
	}

	// Check stake exists
	stake, found := k.Keeper.GetStake(ctx, msg.Address)
	if !found {
		return nil, fmt.Errorf("no stake found for address %s", msg.Address)
	}

	// Update boost
	stake.Boost = msg.Boost
	k.Keeper.SetStake(ctx, stake)

	ctx.Logger().Info("Boost updated",
		"address", msg.Address,
		"boost", msg.Boost,
	)

	return &types.MsgUpdateBoostResponse{}, nil
}
