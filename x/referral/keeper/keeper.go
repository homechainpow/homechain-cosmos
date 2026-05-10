package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/homechain/homechain/x/referral/types"
)

// Keeper of the referral store
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	// paramSpace    paramstypes.Subspace // TODO: Uncomment when params functionality is re-enabled
	bankKeeper bankkeeper.Keeper
	// accountKeeper sdk.AccountKeeper // TODO: Uncomment when AccountKeeper type is fixed
}

// NewKeeper creates a new reward Keeper instance
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bankKeeper bankkeeper.Keeper) Keeper {
	// TODO: Re-enable params functionality when available
	/*
		if !paramSpace.HasKeyTable() {
			paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
		}
	*/

	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		bankKeeper: bankKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) interface{} {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the total set of reward parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	// TODO: Re-enable when params module is available
	// k.paramSpace.Get(ctx, &params)
	return types.DefaultParams()
}

// SetParams sets the reward parameters to the paramspace.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	// TODO: Re-enable when params module is available
	// k.paramSpace.Set(ctx, &params)
}

// GetRewardPool retrieves a reward pool by name
func (k Keeper) GetRewardPool(ctx sdk.Context, name string) (types.RewardPool, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.RewardPoolKey(name)
	bz := store.Get(key)
	if bz == nil {
		// Return default pool if not found
		return types.RewardPool{
			Name:          name,
			Balance:       sdk.NewCoins(),
			LastSettledAt: 0,
			IsActive:      true,
			Permissions:   []string{},
		}, nil
	}

	var pool types.RewardPool
	k.cdc.MustUnmarshal(bz, &pool)
	return pool, nil
}

// SetRewardPool stores a reward pool
func (k Keeper) SetRewardPool(ctx sdk.Context, pool types.RewardPool) {
	store := ctx.KVStore(k.storeKey)
	key := types.RewardPoolKey(pool.Name)
	store.Set(key, k.cdc.MustMarshal(&pool))
}

// GetReferral retrieves referral information from the store
func (k Keeper) GetReferral(ctx sdk.Context, referrer string) (types.ReferralInfo, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ReferralKey(referrer))
	if bz == nil {
		return types.ReferralInfo{}, fmt.Errorf("referral not found")
	}

	var referralInfo types.ReferralInfo
	k.cdc.MustUnmarshal(bz, &referralInfo)
	return referralInfo, nil
}

// SetReferral stores referral information in the store
func (k Keeper) SetReferral(ctx sdk.Context, referralInfo types.ReferralInfo) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ReferralKey(referralInfo.Referrer), k.cdc.MustMarshal(&referralInfo))
}

// HasReferral checks if a referral exists
func (k Keeper) HasReferral(ctx sdk.Context, referrer string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ReferralKey(referrer))
}

// GetReferrals returns all referrals
func (k Keeper) GetReferrals(ctx sdk.Context) []types.ReferralInfo {
	var referrals []types.ReferralInfo
	store := ctx.KVStore(k.storeKey)

	// Use iterator directly from store (SDK v0.54 compatible)
	iterator := store.Iterator([]byte(types.ReferralKeyPrefix), nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var referralInfo types.ReferralInfo
		k.cdc.MustUnmarshal(iterator.Value(), &referralInfo)
		referrals = append(referrals, referralInfo)
	}

	return referrals
}

// GetReferralTree returns the referral tree for a given root address
// TODO: Uncomment when protobuf types are generated
/*
func (k Keeper) GetReferralTree(ctx sdk.Context, rootAddress string, maxDepth uint64) []types.ReferralInfo {
	var tree []types.ReferralInfo
	visited := make(map[string]bool)

	k.buildReferralTree(ctx, rootAddress, 0, maxDepth, visited, &tree)

	return tree
}

// buildReferralTree recursively builds the referral tree
func (k Keeper) buildReferralTree(ctx sdk.Context, address string, currentDepth, maxDepth uint64, visited map[string]bool, tree *[]types.ReferralInfo) {
	if currentDepth >= maxDepth || visited[address] {
		return
	}

	visited[address] = true

	// Find all referrals where this address is the referrer
	referrals := k.GetReferrals(ctx)
	for _, referral := range referrals {
		if referral.Referrer == address && referral.IsActive {
			*tree = append(*tree, referral)
			k.buildReferralTree(ctx, referral.Referred, currentDepth+1, maxDepth, visited, tree)
		}
	}
}

// UpdateReferral updates or creates a referral relationship
func (k Keeper) UpdateReferral(ctx sdk.Context, referrer, referred string) error {
	// Validate addresses
	_, err := sdk.AccAddressFromBech32(referrer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid referrer address: %s", err)
	}

	_, err = sdk.AccAddressFromBech32(referred)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid referred address: %s", err)
	}

	// Check if referral already exists
	if k.HasReferral(ctx, referrer) {
		// Update existing referral
		referralInfo, err := k.GetReferral(ctx, referrer)
		if err != nil {
			return err
		}

		referralInfo.Referred = referred
		referralInfo.IsActive = true
		k.SetReferral(ctx, referralInfo)
	} else {
		// Create new referral
		referralInfo := types.NewReferralInfo(referrer, referred, uint64(ctx.BlockTime().Unix()), 1)
		k.SetReferral(ctx, referralInfo)
	}

	// Log successful referral update
	k.Logger(ctx).Info("Referral updated successfully",
		"referrer", referrer,
		"referred", referred)

	return nil
}

// GetRewardPool retrieves reward pool information
func (k Keeper) GetRewardPool(ctx sdk.Context, poolName string) (types.RewardPool, error) {
	// Get the module account for the pool
	moduleAccount := k.accountKeeper.GetModuleAccount(ctx, poolName)
	if moduleAccount == nil {
		return types.RewardPool{}, errorsmod.Wrapf(sdkerrors.ErrNotFound, "module account %s not found", poolName)
	}

	// Get balance
	balance := k.bankKeeper.GetAllBalances(ctx, moduleAccountAddr)

	// Get permissions
	permissions := k.GetPoolPermissions(poolName)

	// Get last settled time from store
	store := ctx.KVStore(k.storeKey)
	lastSettledKey := []byte(fmt.Sprintf("pool_%s_last_settled", poolName))
	lastSettledAt := uint64(0)
	if bz := store.Get(lastSettledKey); bz != nil {
		k.cdc.MustUnmarshal(bz, &lastSettledAt)
	}

	return types.NewRewardPool(poolName, balance, lastSettledAt, true, permissions), nil
}

// GetPoolAddress returns the address of a reward pool
func (k Keeper) GetPoolAddress(poolName string) sdk.AccAddress {
	switch poolName {
	case types.MiningRewardPool:
		return types.GetMiningRewardPoolAddress()
	case types.ReferralPool:
		return types.GetReferralPoolAddress()
	case types.TreasuryPool:
		return types.GetTreasuryPoolAddress()
	default:
		panic(fmt.Sprintf("unknown pool: %s", poolName))
	}
}

// GetPoolPermissions returns the permissions for a reward pool
func (k Keeper) GetPoolPermissions(poolName string) []string {
	switch poolName {
	case types.MiningRewardPool:
		return types.MiningRewardPoolPermissions
	case types.ReferralPool:
		return types.ReferralPoolPermissions
	case types.TreasuryPool:
		return types.TreasuryPoolPermissions
	default:
		return []string{}
	}
}

// MintMiningRewards mints mining rewards to the mining reward pool
func (k Keeper) MintMiningRewards(ctx sdk.Context, amount sdk.Coins) error {
	if !amount.IsValid() || amount.IsZero() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid amount")
	}

	// Mint coins to mining reward pool
	miningPoolAddr := types.GetMiningRewardPoolAddress()
	if err := k.bankKeeper.MintCoins(ctx, types.MiningRewardPool, amount); err != nil {
		return err
	}

	// Send from module account to mining pool
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.MiningRewardPool, types.MiningRewardPool, amount); err != nil {
		return err
	}

	k.Logger(ctx).Info("Mining rewards minted successfully",
		"amount", amount.String(),
		"pool", types.MiningRewardPool)

	return nil
}

// DistributeRewards distributes rewards from a pool to a recipient
func (k Keeper) DistributeRewards(ctx sdk.Context, poolName string, recipient sdk.AccAddress, amount sdk.Coins) error {
	if !amount.IsValid() || amount.IsZero() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid amount")
	}

	// Get module account for the pool
	poolAddr := k.GetPoolAddress(poolName)

	// Check if pool has sufficient balance
	poolBalance := k.bankKeeper.GetAllBalances(ctx, poolAddr)
	if !poolBalance.IsAllGTE(amount) {
		return errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "insufficient pool balance")
	}

	// Send coins from pool to recipient
	if err := k.bankKeeper.SendCoins(ctx, poolAddr, recipient, amount); err != nil {
		return err
	}

	k.Logger(ctx).Info("Rewards distributed successfully",
		"amount", amount.String(),
		"pool", poolName,
		"recipient", recipient.String())

	return nil
}

// CreateRewardClaim creates a reward claim for a recipient
func (k Keeper) CreateRewardClaim(ctx sdk.Context, recipient string, amount sdk.Coins, claimType string) (string, error) {
	if !amount.IsValid() || amount.IsZero() {
		return "", errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid amount")
	}

	// Generate claim ID
	claimID := fmt.Sprintf("%s_%d_%d", recipient, ctx.BlockHeight(), ctx.BlockTime().Unix())

	// Create reward claim
	rewardClaim := types.NewRewardClaim(claimID, recipient, amount, uint64(ctx.BlockTime().Unix()), claimType)

	// Store claim
	store := ctx.KVStore(k.storeKey)
	key := types.RewardClaimKey(recipient, claimID)
	store.Set(key, k.cdc.MustMarshal(&rewardClaim))

	return claimID, nil
}

// GetRewardClaim retrieves a reward claim
func (k Keeper) GetRewardClaim(ctx sdk.Context, recipient, claimID string) (types.RewardClaim, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.RewardClaimKey(recipient, claimID)
	bz := store.Get(key)
	if bz == nil {
		return types.RewardClaim{}, errorsmod.Wrap(sdkerrors.ErrNotFound, "reward claim not found")
	}

	var rewardClaim types.RewardClaim
	k.cdc.MustUnmarshal(bz, &rewardClaim)
	return rewardClaim, nil
}

// GetPendingRewards returns pending rewards for a recipient
func (k Keeper) GetPendingRewards(ctx sdk.Context, recipient string) (sdk.Coins, error) {
	var totalRewards sdk.Coins
	store := ctx.KVStore(k.storeKey)

	iterator := store.Iterator([]byte(types.RewardClaimKeyPrefix), nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var rewardClaim types.RewardClaim
		k.cdc.MustUnmarshal(iterator.Value(), &rewardClaim)

		if rewardClaim.Recipient == recipient && !rewardClaim.IsClaimed {
			totalRewards = totalRewards.Add(rewardClaim.Amount...)
		}
	}

	return totalRewards, nil
}

// ClaimRewards claims pending rewards for a recipient
func (k Keeper) ClaimRewards(ctx sdk.Context, recipient string) (sdk.Coins, error) {
	recipientAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recipient address: %s", err)
	}

	// Get pending rewards
	pendingRewards, err := k.GetPendingRewards(ctx, recipient)
	if err != nil {
		return nil, err
	}

	if pendingRewards.IsZero() {
		return sdk.Coins{}, nil
	}

	// Distribute rewards from respective pools
	miningRewards := sdk.NewCoins()
	referralRewards := sdk.NewCoins()
	treasuryRewards := sdk.NewCoins()

	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator([]byte(types.RewardClaimKeyPrefix), nil)
	defer iterator.Close()

	// Collect rewards by type and mark as claimed
	for ; iterator.Valid(); iterator.Next() {
		var rewardClaim types.RewardClaim
		k.cdc.MustUnmarshal(iterator.Value(), &rewardClaim)

		if rewardClaim.Recipient == recipient && !rewardClaim.IsClaimed {
			// Mark as claimed
			rewardClaim.IsClaimed = true
			store.Set(iterator.Key(), k.cdc.MustMarshal(&rewardClaim))

			// Add to respective pool rewards
			switch rewardClaim.ClaimType {
			case "mining":
				miningRewards = miningRewards.Add(rewardClaim.Amount...)
			case "referral":
				referralRewards = referralRewards.Add(rewardClaim.Amount...)
			case "treasury":
				treasuryRewards = treasuryRewards.Add(rewardClaim.Amount...)
			}
		}
	}

	// Distribute rewards from each pool
	if !miningRewards.IsZero() {
		if err := k.DistributeRewards(ctx, types.MiningRewardPool, recipientAddr, miningRewards); err != nil {
			return nil, err
		}
	}

	if !referralRewards.IsZero() {
		if err := k.DistributeRewards(ctx, types.ReferralPool, recipientAddr, referralRewards); err != nil {
			return nil, err
		}
	}

	if !treasuryRewards.IsZero() {
		if err := k.DistributeRewards(ctx, types.TreasuryPool, recipientAddr, treasuryRewards); err != nil {
			return nil, err
		}
	}

	k.Logger(ctx).Info("Rewards claimed successfully",
		"recipient", recipient,
		"total_amount", pendingRewards.String())

	return pendingRewards, nil
}

// SettleRewards settles rewards for a period
func (k Keeper) SettleRewards(ctx sdk.Context, period uint64) error {
	params := k.GetParams(ctx)

	// Calculate total rewards to settle (this would come from mining activity)
	totalMined := sdk.NewCoins(sdk.NewCoin("uhome", sdk.NewInt(1000000))) // 1 HOME token per block as example

	// Calculate reward distribution
	miningRewards := sdk.NewCoins()
	referralRewards := sdk.NewCoins()
	treasuryRewards := sdk.NewCoins()

	for _, coin := range totalMined {
		miningAmount := sdk.NewCoin(coin.Denom, coin.Amount.Mul(sdk.NewInt(int64(params.MiningRewardPercentage))).Quo(sdk.NewInt(100)))
		referralAmount := sdk.NewCoin(coin.Denom, coin.Amount.Mul(sdk.NewInt(int64(params.ReferralPercentage))).Quo(sdk.NewInt(100)))
		treasuryAmount := sdk.NewCoin(coin.Denom, coin.Amount.Mul(sdk.NewInt(int64(params.TreasuryPercentage))).Quo(sdk.NewInt(100)))

		miningRewards = miningRewards.Add(miningAmount)
		referralRewards = referralRewards.Add(referralAmount)
		treasuryRewards = treasuryRewards.Add(treasuryAmount)
	}

	// Mint rewards to respective pools
	if !miningRewards.IsZero() {
		if err := k.MintMiningRewards(ctx, miningRewards); err != nil {
			return err
		}
	}

	if !referralRewards.IsZero() {
		if err := k.bankKeeper.MintCoins(ctx, types.ReferralPool, referralRewards); err != nil {
			return err
		}
	}

	if !treasuryRewards.IsZero() {
		if err := k.bankKeeper.MintCoins(ctx, types.TreasuryPool, treasuryRewards); err != nil {
			return err
		}
	}

	// Create settlement record
	settlementInfo := types.NewSettlementInfo(period, totalMined, miningRewards, referralRewards, treasuryRewards, uint64(ctx.BlockTime().Unix()))

	// Store settlement
	store := ctx.KVStore(k.storeKey)
	key := types.SettlementKey(period)
	store.Set(key, k.cdc.MustMarshal(&settlementInfo))

	// Update pool last settled times
	k.updatePoolLastSettled(ctx, types.MiningRewardPool)
	k.updatePoolLastSettled(ctx, types.ReferralPool)
	k.updatePoolLastSettled(ctx, types.TreasuryPool)

	k.Logger(ctx).Info("Rewards settled successfully",
		"period", period,
		"total_mined", totalMined.String(),
		"mining_rewards", miningRewards.String(),
		"referral_rewards", referralRewards.String(),
		"treasury_rewards", treasuryRewards.String())

	return nil
}

// updatePoolLastSettled updates the last settled time for a pool
func (k Keeper) updatePoolLastSettled(ctx sdk.Context, poolName string) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("pool_%s_last_settled", poolName))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(ctx.BlockTime().Unix()))
	store.Set(key, bz)
}

// GetParams returns the total set of reward parameters.
// TODO: Uncomment when params functionality is re-enabled
/*
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.Get(ctx, &params)
	return params
}

// SetParams sets the reward parameters to the paramspace.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.Set(ctx, &params)
}
*/

// ValidateGenesis validates the genesis state
func (k Keeper) ValidateGenesis(ctx sdk.Context, genesisState *types.GenesisState) error {
	if err := genesisState.Validate(); err != nil {
		return err
	}

	// Validate reward pools exist
	// TODO: Re-enable when accountKeeper is available
	// for _, pool := range genesisState.RewardPools {
	// 	moduleAccount := k.accountKeeper.GetModuleAccount(ctx, pool.Name)
	// 	if moduleAccount == nil {
	// 		return fmt.Errorf("module account %s not found", pool.Name)
	// 	}
	// }

	return nil
}

// InitGenesis initializes the genesis state
func (k Keeper) InitGenesis(ctx sdk.Context, genesisState *types.GenesisState) {
	if err := k.ValidateGenesis(ctx, genesisState); err != nil {
		panic(err)
	}

	// Set parameters
	// TODO: Re-enable when SetParams is available
	// k.SetParams(ctx, genesisState.Params)

	// Initialize referrals
	for _, referral := range genesisState.Referrals {
		k.SetReferral(ctx, referral)
	}

	// Initialize reward pools
	// TODO: Re-enable when accountKeeper is available
	// for _, pool := range genesisState.RewardPools {
	// 	// Ensure module account exists
	// 	if !k.accountKeeper.HasAccount(ctx, k.GetPoolAddress(pool.Name)) {
	// 		panic(fmt.Sprintf("module account %s does not exist", pool.Name))
	// 	}
	// }

	// Initialize reward claims
	for _, claim := range genesisState.RewardClaims {
		store := ctx.KVStore(k.storeKey)
		key := types.RewardClaimKey(claim.Recipient, claim.ClaimID)
		store.Set(key, k.cdc.MustMarshal(&claim))
	}

	// Initialize settlements
	for _, settlement := range genesisState.Settlements {
		store := ctx.KVStore(k.storeKey)
		key := types.SettlementKey(settlement.Period)
		store.Set(key, k.cdc.MustMarshal(&settlement))
	}
}

// ExportGenesis exports the genesis state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	referrals := k.GetReferrals(ctx)

	// Get reward pools
	var rewardPools []types.RewardPool
	poolNames := []string{types.MiningRewardPool, types.ReferralPool, types.TreasuryPool}
	for _, poolName := range poolNames {
		pool, err := k.GetRewardPool(ctx, poolName)
		if err != nil {
			panic(err)
		}
		rewardPools = append(rewardPools, pool)
	}

	// Get reward claims
	var rewardClaims []types.RewardClaim
	store := ctx.KVStore(k.storeKey)
	// Use iterator directly from store (SDK v0.54 compatible)
	iterator := store.Iterator([]byte(types.RewardClaimKeyPrefix), nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var rewardClaim types.RewardClaim
		k.cdc.MustUnmarshal(iterator.Value(), &rewardClaim)
		rewardClaims = append(rewardClaims, rewardClaim)
	}

	// Get settlements
	var settlements []types.SettlementInfo
	iterator2 := store.Iterator([]byte(types.SettlementKeyPrefix), nil)
	defer iterator2.Close()

	for ; iterator2.Valid(); iterator2.Next() {
		var settlement types.SettlementInfo
		k.cdc.MustUnmarshal(iterator2.Value(), &settlement)
		settlements = append(settlements, settlement)
	}

	return types.NewGenesisState(params, referrals, rewardPools, rewardClaims, settlements)
}
