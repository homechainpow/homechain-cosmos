package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/homechain/homechain/x/mining/types"
	pohkeeper "github.com/homechain/homechain/x/poh/keeper"
)

// Keeper of the mining store
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	// paramSpace paramstypes.Subspace // TODO: Uncomment when params functionality is re-enabled
	bankKeeper bankkeeper.Keeper
	pohKeeper  pohkeeper.Keeper
}

// NewKeeper creates a new mining Keeper instance
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bankKeeper bankkeeper.Keeper, pohKeeper pohkeeper.Keeper) Keeper {
	// TODO: Re-enable params functionality when available
	/*
		if !paramSpace.HasKeyTable() {
			paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
		}
	*/

	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		// paramSpace: paramSpace,
		bankKeeper: bankKeeper,
		pohKeeper:  pohKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) interface{} {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetMiner retrieves miner information from the store
func (k Keeper) GetMiner(ctx sdk.Context, address string) (types.MinerInfo, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.MinerKey(address))
	if bz == nil {
		return types.MinerInfo{}, fmt.Errorf("miner not found")
	}

	var minerInfo types.MinerInfo
	k.cdc.MustUnmarshal(bz, &minerInfo)
	return minerInfo, nil
}

// SetMiner stores miner information in the store
func (k Keeper) SetMiner(ctx sdk.Context, minerInfo types.MinerInfo) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.MinerKey(minerInfo.Address), k.cdc.MustMarshal(&minerInfo))
}

// HasMiner checks if a miner exists
func (k Keeper) HasMiner(ctx sdk.Context, address string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.MinerKey(address))
}

// GetMiners returns all miners
func (k Keeper) GetMiners(ctx sdk.Context) []types.MinerInfo {
	var miners []types.MinerInfo
	store := ctx.KVStore(k.storeKey)

	// Use iterator directly from store (SDK v0.54 compatible)
	iterator := store.Iterator([]byte(types.MinerKeyPrefix), nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var minerInfo types.MinerInfo
		k.cdc.MustUnmarshal(iterator.Value(), &minerInfo)
		miners = append(miners, minerInfo)
	}

	return miners
}

// GetActiveMiners returns all active miners
func (k Keeper) GetActiveMiners(ctx sdk.Context) []types.MinerInfo {
	var activeMiners []types.MinerInfo
	allMiners := k.GetMiners(ctx)

	for _, miner := range allMiners {
		if miner.IsActive {
			activeMiners = append(activeMiners, miner)
		}
	}

	return activeMiners
}

// RegisterMiner registers a new miner
func (k Keeper) RegisterMiner(ctx sdk.Context, address, deviceInfo, referrer string) error {
	// Check if miner already exists
	if k.HasMiner(ctx, address) {
		return fmt.Errorf("miner already registered")
	}

	// Validate address
	_, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	// Create new miner
	minerInfo := types.NewMinerInfo(address, deviceInfo, referrer, uint64(ctx.BlockTime().Unix()))

	// Store miner
	k.SetMiner(ctx, minerInfo)

	// Log successful registration
	// k.Logger(ctx).Info("Miner registered successfully",
	// 	"address", address,
	// 	"device_info", deviceInfo,
	// 	"referrer", referrer)

	return nil
}

// SubmitShare submits a mining share for verification
func (k Keeper) SubmitShare(ctx sdk.Context, shareData types.ShareData) error {
	// Validate share data
	if err := shareData.Validate(); err != nil {
		return err
	}

	// Check if miner exists
	miner, err := k.GetMiner(ctx, shareData.Miner)
	if err != nil {
		return fmt.Errorf("miner not found")
	}

	// Check if miner is active
	if !miner.IsActive {
		return fmt.Errorf("miner is not active")
	}

	// Check rate limit
	if err := k.CheckRateLimit(ctx, shareData.Miner); err != nil {
		return err
	}

	// Check block share limit
	if err := k.CheckBlockShareLimit(ctx); err != nil {
		return err
	}

	// Get current PoH hash
	currentPoH, err := k.pohKeeper.GetCurrentPoH(ctx)
	if err != nil {
		return err
	}

	// Verify share uses correct PoH hash
	if shareData.PrevHash != currentPoH.NewHash {
		return fmt.Errorf("share uses invalid PoH hash")
	}

	// Generate hash for the share (if not already provided)
	if len(shareData.Hash) == 0 {
		hash, err := shareData.HashShare()
		if err != nil {
			return fmt.Errorf("failed to generate share hash: %w", err)
		}
		shareData.Hash = hash
	}

	// Verify share difficulty
	if !shareData.VerifyShare() {
		return fmt.Errorf("share does not meet difficulty requirement")
	}

	// Store share
	k.StoreShare(ctx, shareData)

	// Update miner statistics
	k.UpdateMinerStats(ctx, shareData.Miner, true)

	// Update rate limit
	k.UpdateRateLimit(ctx, shareData.Miner)

	// Log successful share submission
	// k.Logger(ctx).Info("Share submitted successfully",
	// 	"miner", shareData.Miner,
	// 	"hash", shareData.Hash[:8]+"...",
	// 	"difficulty", shareData.Difficulty)

	return nil
}

// StoreShare stores share data in the store
func (k Keeper) StoreShare(ctx sdk.Context, shareData types.ShareData) {
	store := ctx.KVStore(k.storeKey)
	key := types.ShareKey(shareData.Miner, shareData.Timestamp, shareData.Nonce)
	store.Set(key, k.cdc.MustMarshal(&shareData))
}

// UpdateMinerStats updates miner statistics
func (k Keeper) UpdateMinerStats(ctx sdk.Context, minerAddr string, validShare bool) {
	miner, err := k.GetMiner(ctx, minerAddr)
	if err != nil {
		return
	}

	miner.TotalShares++
	miner.LastShareAt = uint64(ctx.BlockTime().Unix())

	if validShare {
		miner.ValidShares++
	}

	k.SetMiner(ctx, miner)
}

// CheckRateLimit checks if miner is within rate limit
func (k Keeper) CheckRateLimit(ctx sdk.Context, minerAddr string) error {
	params := k.GetParams(ctx)
	store := ctx.KVStore(k.storeKey)

	key := types.RateLimitKey(minerAddr)
	bz := store.Get(key)

	if bz == nil {
		// No previous submissions, allow
		return nil
	}

	// Use binary encoding for uint64
	lastSubmissionTime := uint64(0)
	if len(bz) == 8 {
		lastSubmissionTime = uint64(bz[0])<<56 | uint64(bz[1])<<48 | uint64(bz[2])<<40 | uint64(bz[3])<<32 |
			uint64(bz[4])<<24 | uint64(bz[5])<<16 | uint64(bz[6])<<8 | uint64(bz[7])
	}

	currentTime := uint64(ctx.BlockTime().Unix())

	// Check if enough time has passed (60 seconds / rate limit)
	minInterval := uint64(60) / params.RateLimitPerMinute

	if currentTime-lastSubmissionTime < minInterval {
		return fmt.Errorf("rate limit exceeded")
	}

	return nil
}

// UpdateRateLimit updates rate limit for miner
func (k Keeper) UpdateRateLimit(ctx sdk.Context, minerAddr string) {
	store := ctx.KVStore(k.storeKey)
	key := types.RateLimitKey(minerAddr)

	currentTime := uint64(ctx.BlockTime().Unix())
	// Use binary encoding for uint64
	bz := make([]byte, 8)
	bz[0] = byte(currentTime >> 56)
	bz[1] = byte(currentTime >> 48)
	bz[2] = byte(currentTime >> 40)
	bz[3] = byte(currentTime >> 32)
	bz[4] = byte(currentTime >> 24)
	bz[5] = byte(currentTime >> 16)
	bz[6] = byte(currentTime >> 8)
	bz[7] = byte(currentTime)
	store.Set(key, bz)
}

// CheckBlockShareLimit checks if block share limit is reached
func (k Keeper) CheckBlockShareLimit(ctx sdk.Context) error {
	params := k.GetParams(ctx)

	// Count shares in current block
	sharesInBlock := k.CountSharesInBlock(ctx)

	if sharesInBlock >= params.MaxSharesPerBlock {
		return fmt.Errorf("block share limit reached")
	}

	return nil
}

// CountSharesInBlock counts shares in current block
func (k Keeper) CountSharesInBlock(ctx sdk.Context) uint64 {
	var count uint64
	store := ctx.KVStore(k.storeKey)

	// Use iterator directly from store (SDK v0.54 compatible)
	iterator := store.Iterator([]byte(types.ShareKeyPrefix), nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var shareData types.ShareData
		k.cdc.MustUnmarshal(iterator.Value(), &shareData)

		// Check if share is from current block (within 5 seconds tolerance)
		if uint64(ctx.BlockTime().Unix())-shareData.Timestamp <= 5 {
			count++
		}
	}

	return count
}

// MinerHeartbeat updates miner heartbeat
func (k Keeper) MinerHeartbeat(ctx sdk.Context, minerAddr string) error {
	miner, err := k.GetMiner(ctx, minerAddr)
	if err != nil {
		return fmt.Errorf("miner not found")
	}

	// Update last activity
	miner.LastShareAt = uint64(ctx.BlockTime().Unix())
	miner.IsActive = true

	k.SetMiner(ctx, miner)

	return nil
}

// DeactivateInactiveMiners deactivates miners that have been inactive
func (k Keeper) DeactivateInactiveMiners(ctx sdk.Context) {
	params := k.GetParams(ctx)
	miners := k.GetMiners(ctx)

	for _, miner := range miners {
		if !miner.IsActive {
			continue
		}

		inactiveTime := uint64(ctx.BlockTime().Unix()) - miner.LastShareAt
		maxInactiveTime := params.MinerInactiveThreshold * 5 // 5 seconds per block

		if inactiveTime > maxInactiveTime {
			miner.IsActive = false
			k.SetMiner(ctx, miner)

			// k.Logger(ctx).Info("Miner deactivated due to inactivity",
			// 	"address", miner.Address,
			// 	"inactive_time", inactiveTime)
		}
	}
}

// GetShareHistory returns share history for a miner
func (k Keeper) GetShareHistory(ctx sdk.Context, minerAddr string, limit uint64) ([]types.ShareData, error) {
	var shares []types.ShareData
	store := ctx.KVStore(k.storeKey)

	// Use iterator directly from store (SDK v0.54 compatible)
	iterator := store.Iterator([]byte(types.ShareKeyPrefix), nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var shareData types.ShareData
		k.cdc.MustUnmarshal(iterator.Value(), &shareData)

		if shareData.Miner == minerAddr {
			shares = append(shares, shareData)

			if len(shares) >= int(limit) {
				break
			}
		}
	}

	return shares, nil
}

// GetParams returns the total set of mining parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	// TODO: Re-enable when params module is available
	// k.paramSpace.Get(ctx, &params)
	return types.DefaultParams()
}

// SetParams sets the mining parameters to the paramspace.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	// TODO: Re-enable when params module is available
	// k.paramSpace.Set(ctx, &params)
}

// ValidateGenesis validates the genesis state
func (k Keeper) ValidateGenesis(ctx sdk.Context, genesisState *types.GenesisState) error {
	if err := genesisState.Validate(); err != nil {
		return err
	}

	// Validate miners
	for _, miner := range genesisState.Miners {
		if err := miner.Validate(); err != nil {
			return err
		}
	}

	// Validate shares
	for _, share := range genesisState.Shares {
		if err := share.Validate(); err != nil {
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

	// Initialize miners
	for _, miner := range genesisState.Miners {
		k.SetMiner(ctx, miner)
	}

	// Initialize shares
	for _, share := range genesisState.Shares {
		k.StoreShare(ctx, share)
	}
}

// ExportGenesis exports the genesis state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	miners := k.GetMiners(ctx)

	// Get all shares
	var shares []types.ShareData
	store := ctx.KVStore(k.storeKey)

	// Use iterator directly from store (SDK v0.54 compatible)
	iterator := store.Iterator([]byte(types.ShareKeyPrefix), nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var shareData types.ShareData
		k.cdc.MustUnmarshal(iterator.Value(), &shareData)
		shares = append(shares, shareData)
	}

	return types.NewGenesisState(params, miners, shares)
}
