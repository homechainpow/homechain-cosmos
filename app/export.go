package app

import (
	"encoding/json"

	"github.com/cometbft/cometbft/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ExportAppStateAndValidators exports the state of the application for a genesis file.
func (app *HomeChainApp) ExportAppStateAndValidators(
	forZeroHeight bool, jailAllowedAddrs []string, modulesToExport []string,
) (servertypes.ExportedApp, error) {
	// Create a context for the export
	ctx := app.NewContext(true)

	// If forZeroHeight is true, we need to prepare the state for export at height 0
	if forZeroHeight {
		app.prepForZeroHeightGenesis(ctx, jailAllowedAddrs)
	}

	// Export the genesis state
	genState, err := app.mm.ExportGenesis(ctx, app.appCodec)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}
	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	// Get the validator set
	validators, err := staking.WriteValidators(ctx, &app.StakingKeeper)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	consensusParams := app.GetConsensusParams(ctx)
	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          ctx.BlockHeight(),
		ConsensusParams: consensusParams.ToProto(),
	}, nil
}

// prepForZeroHeightGenesis prepares the application state for export at height 0.
func (app *HomeChainApp) prepForZeroHeightGenesis(ctx sdk.Context, jailAllowedAddrs []string) {
	applyAllowedAddrs := false

	// Check if there are any allowed addresses
	if len(jailAllowedAddrs) > 0 {
		applyAllowedAddrs = true
	}

	allowedAddrsMap := make(map[string]bool)

	for _, addr := range jailAllowedAddrs {
		_, err := sdk.ValAddressFromBech32(addr)
		if err != nil {
			panic(err)
		}
		allowedAddrsMap[addr] = true
	}

	// Iterate through all validators
	app.StakingKeeper.IterateValidators(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		valAddr, err := sdk.ValAddressFromBech32(validator.GetOperator())
		if err != nil {
			return false
		}
		_, err = app.StakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			return false
		}

		if !applyAllowedAddrs {
			return false
		}

		if allowedAddrsMap[validator.GetOperator()] {
			return false
		}

		return false
	})
}

// ModuleManager returns the app module manager
func (app *HomeChainApp) ModuleManager() *module.Manager {
	return app.mm
}

// GetConsensusParams returns the consensus parameters for the application
func (app *HomeChainApp) GetConsensusParams(ctx sdk.Context) *types.ConsensusParams {
	// Return default consensus params or fetch from store
	return &types.ConsensusParams{
		Block: types.BlockParams{
			MaxBytes: 22020096, // 21MB
			MaxGas:   -1,       // No gas limit
		},
		Evidence: types.EvidenceParams{
			MaxAgeNumBlocks: 100000,
			MaxAgeDuration:  172800000000000, // 48 hours
			MaxBytes:        1048576,         // 1MB
		},
		Validator: types.ValidatorParams{
			PubKeyTypes: []string{"ed25519"},
		},
	}
}

// Ensure interface compliance
var _ servertypes.Application = (*HomeChainApp)(nil)
