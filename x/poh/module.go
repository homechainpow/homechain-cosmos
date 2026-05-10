package poh

import (
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"

	"github.com/homechain/homechain/x/poh/keeper"
	"github.com/homechain/homechain/x/poh/types"
)

const (
	ModuleName = "poh"
	StoreKey   = ModuleName
	RouterKey  = ModuleName
)

var (
	_ module.AppModuleBasic = AppModule{}
	_ module.AppModule      = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

type AppModuleBasic struct {
	cdc codec.Codec
}

func NewAppModuleBasic(cdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

// Name returns the module name
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterLegacyAminoCodec registers the module's types on the given LegacyAmino codec.
func (b AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// types.RegisterLegacyAminoCodec(cdc) // TODO: Uncomment when protobuf is generated
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	// types.RegisterInterfaces(reg) // TODO: Uncomment when protobuf is generated
}

// DefaultGenesis returns the module's default genesis state
func (b AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the module
func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return err
	}
	// TODO: Re-enable when ValidateGenesis is implemented
	// return types.ValidateGenesis(&genState)
	return genState.Validate()
}

// RegisterRESTRoutes registers the module's REST routes
func (b AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	// Register REST routes if needed
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC Gateway routes if needed
}

// GetTxCmd returns the module's root tx command
func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	// TODO: Re-enable when CLI commands are implemented
	// return NewTxCmd()
	return nil
}

// GetQueryCmd returns the module's root query command
func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	// TODO: Re-enable when CLI commands are implemented
	// return NewQueryCmd()
	return nil
}

// AppModule represents an AppModule for the poh module
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Name returns the module name
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// InitGenesis performs genesis initialization for the module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(gs, &genState); err != nil {
		panic(err)
	}
	am.keeper.InitGenesis(ctx, &genState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (am AppModule) ConsensusVersion() uint64 { return 1 }

// LegacyQuerierHandler returns the sdk.Querier for the module.
// TODO: Fix for SDK v0.54.3 - sdk.Querier was removed
/*
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return keeper.NewQuerier(am.keeper, legacyQuerierCdc)
}
*/

// RegisterServices implements module.AppModule interface
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// TODO: Uncomment when protobuf is generated
	// types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	// types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))

	// TODO: Fix migration handler for SDK v0.54
	// Register invariants
	// if err := cfg.RegisterMigration(am.Name(), 1, func(ctx context.Context) error {
	// 	return nil
	// }); err != nil {
	// 	panic(err)
	// }
}

// AppModule implements the AppModule interface
var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ProvideAppModule provides the poh app module.
func ProvideAppModule(cdc codec.Codec, keeper keeper.Keeper) module.AppModule {
	return NewAppModule(cdc, keeper)
}

// ProvideAppModuleBasic provides the poh app module basic.
func ProvideAppModuleBasic(cdc codec.Codec) module.AppModuleBasic {
	return NewAppModuleBasic(cdc)
}

// ProvideKeeper provides the poh keeper.
func ProvideKeeper(cdc codec.Codec, storeKey storetypes.StoreKey) keeper.Keeper {
	// ps paramtypes.Subspace removed - params module deprecated in SDK v0.50+
	return keeper.NewKeeper(cdc, storeKey)
}

// DepInjectInputConfig defines the input configuration for dependency injection.
type DepInjectInputConfig struct {
	depinject.In

	AppCodec codec.Codec
	StoreKey storetypes.StoreKey
	// Subspace paramtypes.Subspace // removed - params module deprecated
}

// DepInjectOutputConfig defines the output configuration for dependency injection.
type DepInjectOutputConfig struct {
	depinject.Out

	Module      module.AppModule
	ModuleBasic module.AppModuleBasic
	Keeper      keeper.Keeper
	// Hooks types.StakingHooks // TODO: Uncomment when StakingHooks is defined
}

// DepInject provides the dependency injection configuration for the poh module.
func DepInject(in DepInjectInputConfig) DepInjectOutputConfig {
	k := keeper.NewKeeper(in.AppCodec, in.StoreKey)

	return DepInjectOutputConfig{
		Module:      NewAppModule(in.AppCodec, k),
		ModuleBasic: NewAppModuleBasic(in.AppCodec),
		Keeper:      k,
	}
}
