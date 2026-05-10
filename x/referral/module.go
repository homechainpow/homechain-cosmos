package referral

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
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/gorilla/mux"

	"github.com/homechain/homechain/x/referral/keeper"
	"github.com/homechain/homechain/x/referral/types"
)

const (
	ModuleName = "referral"
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
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	// TODO: Re-enable when protobuf types are generated
	// types.RegisterInterfaces(reg)
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

// AppModule represents an AppModule for the reward module
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

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// TODO: Re-enable when protobuf types are generated
	// types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	// types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))

	// Register invariants
	// TODO: Fix migration handler for SDK v0.54
	/*
		if err := cfg.RegisterMigration(am.Name(), 1, func(ctx context.Context) error {
			return nil
		}); err != nil {
			panic(err)
		}
	*/
}

// Route returns the module's message routing key.
// TODO: Fix for SDK v0.54 - sdk.Route was removed
/*
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(RouterKey, NewHandler(am.keeper))
}
*/

// QuerierRoute returns the module's query routing key.
func (am AppModule) QuerierRoute() string {
	return RouterKey
}

// LegacyQuerierHandler returns the sdk.Querier for the module.
// TODO: Fix for SDK v0.54 - sdk.Querier was removed
/*
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return keeper.NewQuerier(am.keeper, legacyQuerierCdc)
}
*/

// AppModule implements the AppModule interface
var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ProvideAppModule provides the reward app module.
func ProvideAppModule(cdc codec.Codec, keeper keeper.Keeper) module.AppModule {
	return NewAppModule(cdc, keeper)
}

// ProvideAppModuleBasic provides the reward app module basic.
func ProvideAppModuleBasic(cdc codec.Codec) module.AppModuleBasic {
	return NewAppModuleBasic(cdc)
}

// ProvideKeeper provides the reward keeper.
func ProvideKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, bankKeeper bankkeeper.Keeper) keeper.Keeper {
	return keeper.NewKeeper(cdc, storeKey, bankKeeper)
}

// DepInjectInputConfig defines the input configuration for dependency injection.
type DepInjectInputConfig struct {
	depinject.In

	AppCodec   codec.Codec
	StoreKey   storetypes.StoreKey
	BankKeeper bankkeeper.Keeper
}

// DepInjectOutputConfig defines the output configuration for dependency injection.
type DepInjectOutputConfig struct {
	depinject.Out

	Module      module.AppModule
	ModuleBasic module.AppModuleBasic
	Keeper      keeper.Keeper
}

// DepInject provides the dependency injection configuration for the reward module.
func DepInject(in DepInjectInputConfig) DepInjectOutputConfig {
	keeper := keeper.NewKeeper(in.AppCodec, in.StoreKey, in.BankKeeper)

	return DepInjectOutputConfig{
		Module:      NewAppModule(in.AppCodec, keeper),
		ModuleBasic: NewAppModuleBasic(in.AppCodec),
		Keeper:      keeper,
	}
}
