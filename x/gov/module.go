package gov

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
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/gorilla/mux"

	"github.com/homechain/homechain/x/gov/keeper"
	"github.com/homechain/homechain/x/gov/types"
)

const (
	ModuleName = "gov"
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
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns the module's default genesis state
func (b AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the module
func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return err
	}
	return types.ValidateGenesis(&genState)
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
	return NewTxCmd()
}

// GetQueryCmd returns the module's root query command
func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return NewQueryCmd()
}

// AppModule represents an AppModule for the gov module
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
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// NewTxCmd returns a stub tx command
func NewTxCmd() *cobra.Command {
	return &cobra.Command{Use: "gov"}
}

// NewQueryCmd returns a stub query command
func NewQueryCmd() *cobra.Command {
	return &cobra.Command{Use: "gov"}
}

// AppModule implements the AppModule interface
var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ProvideAppModule provides the gov app module.
func ProvideAppModule(cdc codec.Codec, keeper keeper.Keeper) module.AppModule {
	return NewAppModule(cdc, keeper)
}

// ProvideAppModuleBasic provides the gov app module basic.
func ProvideAppModuleBasic(cdc codec.Codec) module.AppModuleBasic {
	return NewAppModuleBasic(cdc)
}

// ProvideKeeper provides the gov keeper.
func ProvideKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, ps paramstypes.Subspace, bankKeeper bankkeeper.Keeper) keeper.Keeper {
	return keeper.NewKeeper(cdc, storeKey, ps, bankKeeper)
}

// DepInjectInputConfig defines the input configuration for dependency injection.
type DepInjectInputConfig struct {
	depinject.In

	AppCodec   codec.Codec
	StoreKey   storetypes.StoreKey
	Subspace   paramstypes.Subspace
	BankKeeper bankkeeper.Keeper
}

// DepInjectOutputConfig defines the output configuration for dependency injection.
type DepInjectOutputConfig struct {
	depinject.Out

	Module      module.AppModule
	ModuleBasic module.AppModuleBasic
	Keeper      keeper.Keeper
}

// DepInject provides the dependency injection configuration for the gov module.
func DepInject(in DepInjectInputConfig) DepInjectOutputConfig {
	keeper := keeper.NewKeeper(in.AppCodec, in.StoreKey, in.Subspace, in.BankKeeper)

	return DepInjectOutputConfig{
		Module:      NewAppModule(in.AppCodec, keeper),
		ModuleBasic: NewAppModuleBasic(in.AppCodec),
		Keeper:      keeper,
	}
}
