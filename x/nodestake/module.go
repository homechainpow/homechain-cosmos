package nodestake

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

	"github.com/homechain/homechain/x/nodestake/keeper"
	"github.com/homechain/homechain/x/nodestake/types"
)

const (
	ModuleName = "nodestake"
	StoreKey   = ModuleName
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.AppModule      = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic defines the basic application module
type AppModuleBasic struct{}

func NewAppModuleBasic() AppModuleBasic {
	return AppModuleBasic{}
}

// Name returns the module name
func (AppModuleBasic) Name() string { return ModuleName }

// RegisterLegacyAminoCodec registers the module's types
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {}

// DefaultGenesis returns the module's default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis validates the genesis state
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return err
	}
	return types.ValidateGenesis(&genState)
}

// GetTxCmd returns the module's root tx command
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return &cobra.Command{Use: ModuleName}
}

// GetQueryCmd returns the module's root query command
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return &cobra.Command{Use: ModuleName}
}

// RegisterGRPCGatewayRoutes registers gRPC gateway routes
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {}

// AppModule defines the application module
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule
func NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(),
		keeper:         keeper,
	}
}

// IsOnePerModuleType implements depinject.OnePerModuleType
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements appmodule.AppModule
func (am AppModule) IsAppModule() {}

// Name returns the module name
func (am AppModule) Name() string { return ModuleName }

// RegisterServices registers module services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
}

// InitGenesis performs genesis initialization
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(gs, &genState); err != nil {
		panic(err)
	}
	am.keeper.InitGenesis(ctx, &genState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns exported genesis state
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion returns consensus version
func (am AppModule) ConsensusVersion() uint64 { return 1 }

// ProvideAppModule provides the app module
func ProvideAppModule(keeper keeper.Keeper) module.AppModule {
	return NewAppModule(keeper)
}

// ProvideAppModuleBasic provides the app module basic
func ProvideAppModuleBasic() module.AppModuleBasic {
	return NewAppModuleBasic()
}

// ProvideKeeper provides the keeper
func ProvideKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, ps paramstypes.Subspace, bk bankkeeper.Keeper) keeper.Keeper {
	return keeper.NewKeeper(cdc, storeKey, ps, bk)
}

// DepInjectInput defines the input configuration
type DepInjectInput struct {
	depinject.In
	AppCodec   codec.Codec
	StoreKey   storetypes.StoreKey
	Subspace   paramstypes.Subspace
	BankKeeper bankkeeper.Keeper
}

// DepInjectOutput defines the output configuration
type DepInjectOutput struct {
	depinject.Out
	Module      module.AppModule
	ModuleBasic module.AppModuleBasic
	Keeper      keeper.Keeper
}

// DepInject provides the dependency injection configuration
func DepInject(in DepInjectInput) DepInjectOutput {
	k := keeper.NewKeeper(in.AppCodec, in.StoreKey, in.Subspace, in.BankKeeper)
	return DepInjectOutput{
		Module:      NewAppModule(k),
		ModuleBasic: NewAppModuleBasic(),
		Keeper:      k,
	}
}
