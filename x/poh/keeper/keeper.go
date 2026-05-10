package keeper

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/crypto/argon2"

	pohtypes "github.com/homechain/homechain/x/poh/types"
)

// Keeper of the poh store
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey types.StoreKey
	// paramSpace paramstypes.Subspace
}

// NewKeeper creates a new poh Keeper instance
func NewKeeper(cdc codec.BinaryCodec, storeKey types.StoreKey) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", pohtypes.ModuleName))
}

// GetCurrentPoH retrieves the current PoH data from the store
func (k Keeper) GetCurrentPoH(ctx sdk.Context) (*pohtypes.PoHData, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(pohtypes.CurrentPoHKeyBytes())
	if bz == nil {
		// Return genesis PoH if no current PoH exists
		return &pohtypes.PoHData{
			PrevHash:   pohtypes.GenesisPoHHash,
			NewHash:    pohtypes.GenesisPoHHash,
			Difficulty: pohtypes.DefaultDifficulty,
			Timestamp:  uint64(ctx.BlockTime().Unix()),
		}, nil
	}

	var pohData pohtypes.PoHData
	k.cdc.MustUnmarshal(bz, &pohData)
	return &pohData, nil
}

// SetCurrentPoH stores the current PoH data in the store
func (k Keeper) SetCurrentPoH(ctx sdk.Context, pohData pohtypes.PoHData) {
	store := ctx.KVStore(k.storeKey)
	store.Set(pohtypes.CurrentPoHKeyBytes(), k.cdc.MustMarshal(&pohData))
}

// GetPoHSequence retrieves PoH data for a specific height
func (k Keeper) GetPoHSequence(ctx sdk.Context, height uint64) (*pohtypes.PoHData, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(pohtypes.PoHSequenceKey(height))
	if bz == nil {
		return nil, fmt.Errorf("PoH sequence not found at height %d", height)
	}

	var pohData pohtypes.PoHData
	k.cdc.MustUnmarshal(bz, &pohData)
	return &pohData, nil
}

// SetPoHSequence stores PoH data for a specific height
func (k Keeper) SetPoHSequence(ctx sdk.Context, height uint64, pohData pohtypes.PoHData) {
	store := ctx.KVStore(k.storeKey)
	store.Set(pohtypes.PoHSequenceKey(height), k.cdc.MustMarshal(&pohData))
}

// GetCurrentDifficulty retrieves the current difficulty from the store
func (k Keeper) GetCurrentDifficulty(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(pohtypes.DifficultyKeyBytes())
	if bz == nil {
		return pohtypes.DefaultDifficulty
	}

	var difficulty uint64
	buf := bytes.NewReader(bz)
	binary.Read(buf, binary.BigEndian, &difficulty)
	return difficulty
}

// SetCurrentDifficulty stores the current difficulty in the store
func (k Keeper) SetCurrentDifficulty(ctx sdk.Context, difficulty uint64) {
	store := ctx.KVStore(k.storeKey)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, difficulty)
	store.Set(pohtypes.DifficultyKeyBytes(), buf.Bytes())
}

// HashWithArgon2id generates Argon2id hash using pure Go implementation
func (k Keeper) HashWithArgon2id(input string, difficulty uint64) (string, error) {
	config := pohtypes.DefaultArgon2idConfig()

	password := []byte(input)
	salt := []byte(input)

	hash := argon2.IDKey(password, salt, config.Time, config.Memory, uint8(config.Threads), config.HashLen)
	hashHex := fmt.Sprintf("%x", hash)

	// Verify hash meets difficulty requirement
	if !pohtypes.VerifyDifficulty(hashHex, difficulty) {
		return "", fmt.Errorf("generated hash does not meet difficulty requirement: %d", difficulty)
	}

	return hashHex, nil
}

// GeneratePoHHash generates a new PoH hash based on previous hash and current context
func (k Keeper) GeneratePoHHash(ctx sdk.Context, prevHash string) (string, error) {
	// Get current difficulty
	difficulty := k.GetCurrentDifficulty(ctx)

	// Create input for Argon2id - combine previous hash with block context
	input := fmt.Sprintf("%s:%s:%d:%d",
		prevHash,
		ctx.BlockHeader().ProposerAddress,
		ctx.BlockHeight(),
		ctx.BlockTime().Unix(),
	)

	// Generate hash using Argon2id C wrapper
	return k.HashWithArgon2id(input, difficulty)
}

// VerifyPoHSequence verifies the PoH sequence is valid
func (k Keeper) VerifyPoHSequence(ctx sdk.Context, prevHash, newHash string, difficulty uint64) bool {
	// Basic format validation
	if len(prevHash) != 64 || len(newHash) != 64 {
		k.Logger(ctx).Error("Invalid hash format", "prev_hash_len", len(prevHash), "new_hash_len", len(newHash))
		return false
	}

	// Verify difficulty requirement
	if !pohtypes.VerifyDifficulty(newHash, difficulty) {
		k.Logger(ctx).Error("Hash does not meet difficulty requirement",
			"hash", newHash, "difficulty", difficulty)
		return false
	}

	// Get current PoH to verify sequence continuity
	currentPoH, err := k.GetCurrentPoH(ctx)
	if err != nil {
		k.Logger(ctx).Error("Failed to get current PoH", "error", err)
		return false
	}

	// Verify previous hash matches current state
	if currentPoH.NewHash != prevHash {
		k.Logger(ctx).Error("Previous hash mismatch",
			"expected", currentPoH.NewHash, "received", prevHash)
		return false
	}

	return true
}

// SubmitPoH submits new PoH data to the store
func (k Keeper) SubmitPoH(ctx sdk.Context, pohData pohtypes.PoHData) error {
	// Validate PoH data
	if err := pohData.Validate(); err != nil {
		return err
	}

	// Verify PoH sequence
	if !k.VerifyPoHSequence(ctx, pohData.PrevHash, pohData.NewHash, pohData.Difficulty) {
		return fmt.Errorf("invalid PoH sequence")
	}

	// Store PoH sequence for current height
	k.SetPoHSequence(ctx, uint64(ctx.BlockHeight()), pohData)

	// Update current PoH
	k.SetCurrentPoH(ctx, pohData)

	// Log successful PoH submission
	k.Logger(ctx).Info("PoH submitted successfully",
		"height", ctx.BlockHeight(),
		"prev_hash", pohData.PrevHash[:8]+"...",
		"new_hash", pohData.NewHash[:8]+"...",
		"difficulty", pohData.Difficulty)

	return nil
}

// GetParams returns the total set of poh parameters.
// TODO: Uncomment when params functionality is re-enabled
/*
func (k Keeper) GetParams(ctx sdk.Context) (params pohtypes.Params) {
	k.paramSpace.Get(ctx, &params)
	return params
}

// SetParams sets the poh parameters to the paramspace.
func (k Keeper) SetParams(ctx sdk.Context, params pohtypes.Params) {
	k.paramSpace.Set(ctx, &params)
}
*/

// GetPoHHistory returns PoH data for a range of heights
func (k Keeper) GetPoHHistory(ctx sdk.Context, startHeight, endHeight uint64) ([]pohtypes.PoHData, error) {
	var pohHistory []pohtypes.PoHData

	for height := startHeight; height <= endHeight; height++ {
		pohData, err := k.GetPoHSequence(ctx, height)
		if err != nil {
			// Skip missing heights
			continue
		}
		pohHistory = append(pohHistory, *pohData)
	}

	return pohHistory, nil
}

// GetLastNPoH returns the last N PoH entries
func (k Keeper) GetLastNPoH(ctx sdk.Context, n uint64) ([]pohtypes.PoHData, error) {
	currentHeight := uint64(ctx.BlockHeight())
	if currentHeight == 0 {
		return []pohtypes.PoHData{}, nil
	}

	startHeight := currentHeight - n + 1
	if startHeight < 1 {
		startHeight = 1
	}

	return k.GetPoHHistory(ctx, startHeight, currentHeight)
}

// AdjustDifficulty adjusts the mining difficulty based on block time
// TODO: Uncomment when params functionality is re-enabled
/*
func (k Keeper) AdjustDifficulty(ctx sdk.Context) {
	params := k.GetParams(ctx)
	currentDifficulty := k.GetCurrentDifficulty(ctx)

	// Get last N blocks to calculate average block time
	lastNPoH, err := k.GetLastNPoH(ctx, params.HeartbeatPeriod)
	if err != nil || len(lastNPoH) < 2 {
		// Not enough data to adjust difficulty
		return
	}

	// Calculate average block time
	var totalTime uint64
	for i := 1; i < len(lastNPoH); i++ {
		totalTime += lastNPoH[i].Timestamp - lastNPoH[i-1].Timestamp
	}
	avgBlockTime := totalTime / uint64(len(lastNPoH)-1)

	// Target block time is 5 seconds
	targetBlockTime := uint64(5)

	// Adjust difficulty based on average block time
	newDifficulty := currentDifficulty
	if avgBlockTime > targetBlockTime {
		// Blocks are too slow, decrease difficulty
		if newDifficulty > params.MinDifficulty {
			newDifficulty--
		}
	} else if avgBlockTime < targetBlockTime {
		// Blocks are too fast, increase difficulty
		if newDifficulty < params.MaxDifficulty {
			newDifficulty++
		}
	}

	// Update difficulty if it changed
	if newDifficulty != currentDifficulty {
		k.SetCurrentDifficulty(ctx, newDifficulty)
		k.Logger(ctx).Info("Difficulty adjusted",
			"old", currentDifficulty,
			"new", newDifficulty,
			"avg_block_time", avgBlockTime,
			"target_block_time", targetBlockTime)
	}
}
*/

// ValidateGenesis validates the genesis state
func (k Keeper) ValidateGenesis(ctx sdk.Context, genesisState *pohtypes.GenesisState) error {
	if err := genesisState.Validate(); err != nil {
		return err
	}

	// Validate PoH data sequence
	for i, poh := range genesisState.PoHData {
		if i > 0 {
			prevPoH := genesisState.PoHData[i-1]
			if poh.PrevHash != prevPoH.NewHash {
				return fmt.Errorf("PoH sequence broken at index %d: expected prev_hash %s, got %s",
					i, prevPoH.NewHash, poh.PrevHash)
			}
		}
	}

	return nil
}

// InitGenesis initializes the genesis state
func (k Keeper) InitGenesis(ctx sdk.Context, genesisState *pohtypes.GenesisState) {
	if err := k.ValidateGenesis(ctx, genesisState); err != nil {
		panic(err)
	}

	// Set parameters
	// k.SetParams(ctx, genesisState.Params) // TODO: Uncomment when params functionality is re-enabled

	// Set difficulty
	k.SetCurrentDifficulty(ctx, genesisState.Difficulty)

	// Initialize PoH data
	if len(genesisState.PoHData) > 0 {
		// Set the last PoH data as current
		lastPoH := genesisState.PoHData[len(genesisState.PoHData)-1]
		k.SetCurrentPoH(ctx, lastPoH)
	} else {
		// Set genesis PoH
		genesisPoH := pohtypes.NewPoHData(
			pohtypes.GenesisPoHHash,
			pohtypes.GenesisPoHHash,
			pohtypes.DefaultDifficulty,
			uint64(ctx.BlockTime().Unix()),
		)
		k.SetCurrentPoH(ctx, genesisPoH)
	}
}

// ExportGenesis exports the genesis state
func (k Keeper) ExportGenesis(ctx sdk.Context) *pohtypes.GenesisState {
	// params := k.GetParams(ctx) // TODO: Uncomment when params functionality is re-enabled
	params := pohtypes.DefaultParams() // Use default params for now
	difficulty := k.GetCurrentDifficulty(ctx)

	// Get PoH history
	pohHistory, err := k.GetPoHHistory(ctx, 1, uint64(ctx.BlockHeight()))
	if err != nil {
		pohHistory = []pohtypes.PoHData{}
	}

	return pohtypes.NewGenesisState(params, pohHistory, difficulty)
}

// ConsumePoHGas consumes gas for PoH operations to prevent CPU exhaustion attacks
func (k Keeper) ConsumePoHGas(ctx sdk.Context) {
	ctx.GasMeter().ConsumeGas(pohtypes.DefaultPoHGasLimit, "Argon2 PoH Verification")
}
