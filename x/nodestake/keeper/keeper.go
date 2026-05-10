package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/homechain/homechain/x/nodestake/types"
)

// Keeper of the nodestake store
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramSpace paramstypes.Subspace
	bankKeeper bankkeeper.Keeper
}

// NewKeeper creates a new nodestake Keeper instance
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, paramSpace paramstypes.Subspace, bankKeeper bankkeeper.Keeper) Keeper {
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

// GetParams returns the total set of nodestake parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.Get(ctx, types.ParamStoreKey, &params)
	return params
}

// SetParams sets the nodestake parameters to the paramspace.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.Set(ctx, types.ParamStoreKey, params)
}

// InitGenesis initializes the genesis state
func (k Keeper) InitGenesis(ctx sdk.Context, genesisState *types.GenesisState) {
	if err := genesisState.Validate(); err != nil {
		panic(err)
	}
	k.SetParams(ctx, genesisState.Params)
}

// ExportGenesis exports the genesis state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}

// GetStake returns stake info for an address
func (k Keeper) GetStake(ctx sdk.Context, address string) (types.StakeInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.StakeKey(address))
	if bz == nil {
		return types.StakeInfo{}, false
	}
	var stake types.StakeInfo
	k.cdc.MustUnmarshal(bz, &stake)
	return stake, true
}

// SetStake stores stake info
func (k Keeper) SetStake(ctx sdk.Context, stake types.StakeInfo) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.StakeKey(stake.Address), k.cdc.MustMarshal(&stake))
}

// RemoveStake deletes stake info
func (k Keeper) RemoveStake(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.StakeKey(address))
}

// ValidateBoost checks if boost is within valid range (1-101)
func (k Keeper) ValidateBoost(boost uint64) error {
	if boost == 0 || boost > 101 {
		return fmt.Errorf("boost must be between 1 and 101, got %d", boost)
	}
	return nil
}
