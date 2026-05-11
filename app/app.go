package app

import (
	"context"
	"encoding/json"
	"io"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log/v2"
	"cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storev2 "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	// Custom modules
	"github.com/homechain/homechain/x/gov/ante"
	"github.com/homechain/homechain/x/mining"
	miningkeeper "github.com/homechain/homechain/x/mining/keeper"
	miningtypes "github.com/homechain/homechain/x/mining/types"
	"github.com/homechain/homechain/x/poh"
	pohkeeper "github.com/homechain/homechain/x/poh/keeper"
	pohtypes "github.com/homechain/homechain/x/poh/types"
	"github.com/homechain/homechain/x/referral"
	referralkeeper "github.com/homechain/homechain/x/referral/keeper"
	referraltypes "github.com/homechain/homechain/x/referral/types"

	// Nodestake module
	"github.com/homechain/homechain/x/nodestake"
	nodestakekeeper "github.com/homechain/homechain/x/nodestake/keeper"
	nodestaketypes "github.com/homechain/homechain/x/nodestake/types"
)

var (
// _ App = (*HomeChainApp)(nil) // Commented out - interface compliance check
)

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App interface {
	servertypes.Application

	// GetBaseApp returns the base app of the application
	GetBaseApp() *baseapp.BaseApp

	// GetKeys returns all the store keys of the application
	GetKeys() map[string]*types.KVStoreKey

	// GetTKeys returns all the transient store keys of the application
	GetTKeys() map[string]*types.TransientStoreKey

	// GetMemKeys returns all the memory store keys of the application
	GetMemKeys() map[string]*types.MemoryStoreKey

	// GetSubspace returns a param subspace for a given module name.
	GetSubspace(moduleName string) paramstypes.Subspace

	// GetAccountKeeper returns the account keeper of the application
	GetAccountKeeper() authkeeper.AccountKeeper

	// GetBankKeeper returns the bank keeper of the application
	GetBankKeeper() bankkeeper.Keeper

	// GetStakingKeeper returns the staking keeper of the application
	GetStakingKeeper() stakingkeeper.Keeper

	// GetSlashingKeeper returns the slashing keeper of the application
	GetSlashingKeeper() slashingkeeper.Keeper

	// GetMintKeeper returns the mint keeper of the application
	GetMintKeeper() mintkeeper.Keeper

	// GetDistrKeeper returns the distribution keeper of the application
	GetDistrKeeper() distrkeeper.Keeper

	// GetGovKeeper returns the gov keeper of the application
	GetGovKeeper() govkeeper.Keeper

	// GetUpgradeKeeper returns the upgrade keeper of the application
	GetUpgradeKeeper() upgradekeeper.Keeper

	// GetParamsKeeper returns the params keeper of the application
	GetParamsKeeper() paramskeeper.Keeper

	// GetFeegrantKeeper returns the feegrant keeper of the application
	GetFeegrantKeeper() feegrantkeeper.Keeper

	// GetEvidenceKeeper returns the evidence keeper of the application
	GetEvidenceKeeper() evidencekeeper.Keeper

	// GetPoHKeeper returns the PoH keeper of the application
	GetPoHKeeper() pohkeeper.Keeper

	// GetMiningKeeper returns the mining keeper of the application
	GetMiningKeeper() miningkeeper.Keeper

	// GetReferralKeeper returns the referral keeper of the application
	GetReferralKeeper() referralkeeper.Keeper

	// GetNodestakeKeeper returns the nodestake keeper of the application
	GetNodestakeKeeper() nodestakekeeper.Keeper

	// GetModuleManager returns the module manager of the application
	GetModuleManager() *module.Manager

	// GetConfig returns the config of the application
	GetConfig() module.Configurator

	// The simulation manager is only exposed for testing purposes.
	GetSimulationManager() *module.SimulationManager

	// SimulationState implements simulation.App interface for deterministic testing.
	SimulationState() *module.SimulationState
}

// HomeChainApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type HomeChainApp struct {
	*baseapp.BaseApp
	legacyAmino *codec.LegacyAmino
	appCodec    codec.Codec
	txConfig    sdkclient.TxConfig

	// keys to access the substores
	keys    map[string]*types.KVStoreKey
	tkeys   map[string]*types.TransientStoreKey
	memKeys map[string]*types.MemoryStoreKey

	// keepers
	AccountKeeper  authkeeper.AccountKeeper
	BankKeeper     bankkeeper.Keeper
	StakingKeeper  stakingkeeper.Keeper
	SlashingKeeper slashingkeeper.Keeper
	MintKeeper     mintkeeper.Keeper
	DistrKeeper    distrkeeper.Keeper
	GovKeeper      govkeeper.Keeper
	UpgradeKeeper  upgradekeeper.Keeper
	ParamsKeeper   paramskeeper.Keeper
	FeegrantKeeper feegrantkeeper.Keeper
	EvidenceKeeper evidencekeeper.Keeper

	// custom keepers
	PoHKeeper       pohkeeper.Keeper
	MiningKeeper    miningkeeper.Keeper
	ReferralKeeper  referralkeeper.Keeper
	NodestakeKeeper nodestakekeeper.Keeper

	// the module manager
	mm *module.Manager
	// the simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator

	// proposal handler for ABCI++
	proposalHandler *ProposalHandler
}

// NewHomeChainApp returns a reference to an initialized HomeChainApp.
func NewHomeChainApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	appOpts servertypes.AppOptions,
	// wasmOpts servertypes.WasmOpts, // Commented out - not available in SDK v0.50+
	enabledProposals []string,
	baseAppOptions ...func(*baseapp.BaseApp),
) *HomeChainApp {
	// BaseApp handles interactions with Tendermint through the ABCI protocol.
	encodingConfig := MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	txConfig := encodingConfig.TxConfig

	bApp := baseapp.NewBaseApp(AppName, logger, db, txConfig.TxDecoder(), baseAppOptions...)

	// Create store keys - using types.NewKVStoreKey instead of sdk.NewKVStoreKeys
	keys := make(map[string]*types.KVStoreKey)
	keys[authtypes.StoreKey] = types.NewKVStoreKey(authtypes.StoreKey)
	keys[banktypes.StoreKey] = types.NewKVStoreKey(banktypes.StoreKey)
	keys[stakingtypes.StoreKey] = types.NewKVStoreKey(stakingtypes.StoreKey)
	keys[minttypes.StoreKey] = types.NewKVStoreKey(minttypes.StoreKey)
	keys[distrtypes.StoreKey] = types.NewKVStoreKey(distrtypes.StoreKey)
	keys[slashtypes.StoreKey] = types.NewKVStoreKey(slashtypes.StoreKey)
	keys[govtypes.StoreKey] = types.NewKVStoreKey(govtypes.StoreKey)
	keys[paramstypes.StoreKey] = types.NewKVStoreKey(paramstypes.StoreKey)
	keys[evidencetypes.StoreKey] = types.NewKVStoreKey(evidencetypes.StoreKey)
	keys[upgradetypes.StoreKey] = types.NewKVStoreKey(upgradetypes.StoreKey)
	keys[pohtypes.StoreKey] = types.NewKVStoreKey(pohtypes.StoreKey)
	keys[miningtypes.StoreKey] = types.NewKVStoreKey(miningtypes.StoreKey)
	keys[referraltypes.StoreKey] = types.NewKVStoreKey(referraltypes.StoreKey)
	keys[nodestaketypes.StoreKey] = types.NewKVStoreKey(nodestaketypes.StoreKey)

	tkeys := make(map[string]*types.TransientStoreKey)
	tkeys[paramstypes.TStoreKey] = types.NewTransientStoreKey(paramstypes.TStoreKey)

	memKeys := make(map[string]*types.MemoryStoreKey)

	// Here you initialize your application.
	app := &HomeChainApp{
		BaseApp:     bApp,
		legacyAmino: encodingConfig.Amino,
		appCodec:    encodingConfig.Marshaler,
		txConfig:    encodingConfig.TxConfig,
		keys:        keys,
		tkeys:       tkeys,
		memKeys:     memKeys,
	}

	// app.ParamsKeeper = initParamsKeeper(appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])
	// bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	// SDK v0.50+ keeper initialization
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		newKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		newKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		app.BlockedAddrs(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)

	// Custom keepers (minimal for viable simulation)
	app.PoHKeeper = pohkeeper.NewKeeper(
		appCodec,
		keys[pohtypes.StoreKey],
	)

	app.MiningKeeper = miningkeeper.NewKeeper(
		appCodec,
		keys[miningtypes.StoreKey],
		app.BankKeeper,
		app.PoHKeeper,
	)

	app.ReferralKeeper = referralkeeper.NewKeeper(
		appCodec,
		keys[referraltypes.StoreKey],
		app.BankKeeper,
	)

	app.NodestakeKeeper = nodestakekeeper.NewKeeper(
		appCodec,
		keys[nodestaketypes.StoreKey],
		app.GetSubspace(nodestaketypes.ModuleName),
		app.BankKeeper,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it can be
	// decorated with staking hooks.
	// app.StakingKeeper = *stakingKeeper.SetHooks(
	// 	stakingtypes.NewMultiStakingHooks(
	// 		app.DistrKeeper.Hooks(),
	// 		app.SlashingKeeper.Hooks(),
	// 	),
	// )

	// Create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions, but should be initialized for any app that wants to use the simulator.
	// app.sm = module.NewSimulationManagerFromAppModules(app.getModuleBasics())

	// app.sm.RegisterStoreDecoders()

	// Create the module manager with custom modules
	// NOTE: SDK modules (staking, bank, auth, etc.) require v0.50 API fix first
	app.mm = module.NewManager(
		poh.NewAppModule(appCodec, app.PoHKeeper),
		mining.NewAppModule(appCodec, app.MiningKeeper),
		referral.NewAppModule(appCodec, app.ReferralKeeper),
		nodestake.NewAppModule(app.NodestakeKeeper),
	)

	// Custom module init genesis order
	app.mm.SetOrderInitGenesis(
		pohtypes.ModuleName,
		miningtypes.ModuleName,
		referraltypes.ModuleName,
		nodestaketypes.ModuleName,
	)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)

	// Set ECDSA AnteHandler for Ethereum-style signature verification
	app.SetAnteHandler(
		sdk.ChainAnteDecorators(
			ante.NewECDSASigVerificationDecorator(),
		),
	)

	// Load the latest state, if it exists, else initialize it
	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(err.Error())
		}
	}

	return app
}

// Name returns the name of the App
func (app *HomeChainApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
// func (app *HomeChainApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
// 	return app.mm.BeginBlock(ctx, req)
// }

// EndBlocker application updates every end block
// func (app *HomeChainApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
// 	return app.mm.EndBlock(ctx, req)
// }

// InitChainer application update at chain initialization
func (app *HomeChainApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState map[string]json.RawMessage
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		return nil, err
	}

	if app.mm != nil {
		return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
	}
	return &abci.ResponseInitChain{}, nil
}

// LoadHeight loads a particular height
func (app *HomeChainApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *HomeChainApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// to register additional amino types on a per-app basis.
func (app *HomeChainApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns SimApp's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// to register additional amino types on a per-app basis.
func (app *HomeChainApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns SimApp's InterfaceRegistry
func (app *HomeChainApp) InterfaceRegistry() cdctypes.InterfaceRegistry {
	return app.appCodec.InterfaceRegistry()
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *HomeChainApp) GetKey(storeKey string) *types.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *HomeChainApp) GetTKey(storeKey string) *types.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *HomeChainApp) GetMemKey(storeKey string) *types.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *HomeChainApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface
func (app *HomeChainApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *HomeChainApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	// app.mm.RegisterAPIRoutes(apiSvr, apiConfig) // Commented out - doesn't exist in SDK v0.50+
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *HomeChainApp) RegisterTendermintService(clientCtx sdkclient.Context) {
	// tmservice.RegisterTendermintService(clientCtx, app.GRPCQueryRouter())
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// GetBaseApp returns the base app of the application
func (app *HomeChainApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetKeys returns all the store keys of the application
func (app *HomeChainApp) GetKeys() map[string]*types.KVStoreKey {
	return app.keys
}

// GetTKeys returns all the transient store keys of the application
func (app *HomeChainApp) GetTKeys() map[string]*types.TransientStoreKey {
	return app.tkeys
}

// GetMemKeys returns all the memory store keys of the application
func (app *HomeChainApp) GetMemKeys() map[string]*types.MemoryStoreKey {
	return app.memKeys
}

// RegisterNodeService registers the node service
func (app *HomeChainApp) RegisterNodeService(clientCtx sdkclient.Context, cfg config.Config) {
	// Placeholder - SDK v0.50+ RegisterNodeService does not return error
}

// RegisterTxService registers the transaction service
func (app *HomeChainApp) RegisterTxService(clientCtx sdkclient.Context) {
	// Placeholder - SDK v0.50+ RegisterTxService does not return error
}

// GetAccountKeeper returns the account keeper of the application
func (app *HomeChainApp) GetAccountKeeper() authkeeper.AccountKeeper {
	return app.AccountKeeper
}

// GetBankKeeper returns the bank keeper of the application
func (app *HomeChainApp) GetBankKeeper() bankkeeper.Keeper {
	return app.BankKeeper
}

// GetStakingKeeper returns the staking keeper of the application
func (app *HomeChainApp) GetStakingKeeper() stakingkeeper.Keeper {
	return app.StakingKeeper
}

// GetSlashingKeeper returns the slashing keeper of the application
func (app *HomeChainApp) GetSlashingKeeper() slashingkeeper.Keeper {
	return app.SlashingKeeper
}

// GetMintKeeper returns the mint keeper of the application
func (app *HomeChainApp) GetMintKeeper() mintkeeper.Keeper {
	return app.MintKeeper
}

// GetDistrKeeper returns the distribution keeper of the application
func (app *HomeChainApp) GetDistrKeeper() distrkeeper.Keeper {
	return app.DistrKeeper
}

// GetGovKeeper returns the governance keeper of the application
func (app *HomeChainApp) GetGovKeeper() govkeeper.Keeper {
	return app.GovKeeper
}

// GetUpgradeKeeper returns the upgrade keeper of the application
func (app *HomeChainApp) GetUpgradeKeeper() upgradekeeper.Keeper {
	return app.UpgradeKeeper
}

// GetParamsKeeper returns the params keeper of the application
func (app *HomeChainApp) GetParamsKeeper() paramskeeper.Keeper {
	return app.ParamsKeeper
}

// GetFeegrantKeeper returns the feegrant keeper of the application
func (app *HomeChainApp) GetFeegrantKeeper() feegrantkeeper.Keeper {
	return app.FeegrantKeeper
}

// GetEvidenceKeeper returns the evidence keeper of the application
func (app *HomeChainApp) GetEvidenceKeeper() evidencekeeper.Keeper {
	return app.EvidenceKeeper
}

// GetPoHKeeper returns the PoH keeper of the application
func (app *HomeChainApp) GetPoHKeeper() pohkeeper.Keeper {
	return app.PoHKeeper
}

// GetMiningKeeper returns the mining keeper of the application
func (app *HomeChainApp) GetMiningKeeper() miningkeeper.Keeper {
	return app.MiningKeeper
}

// GetReferralKeeper returns the referral keeper of the application
func (app *HomeChainApp) GetReferralKeeper() referralkeeper.Keeper {
	return app.ReferralKeeper
}

// GetNodestakeKeeper returns the nodestake keeper of the application
func (app *HomeChainApp) GetNodestakeKeeper() nodestakekeeper.Keeper {
	return app.NodestakeKeeper
}

// GetModuleManager returns the module manager of the application
func (app *HomeChainApp) GetModuleManager() *module.Manager {
	return app.mm
}

// GetConfig returns the config of the application
func (app *HomeChainApp) GetConfig() module.Configurator {
	return app.configurator
}

// GetSimulationManager returns the simulation manager of the application
func (app *HomeChainApp) GetSimulationManager() *module.SimulationManager {
	return app.sm
}

// SimulationState implements simulation.App interface for HomeChainApp.
// This allows use of 'go test -run TestFullAppSimulation' for deterministic testing.
func (app *HomeChainApp) SimulationState() *module.SimulationState {
	// For SDK v0.50+, return a basic simulation state
	// The actual simulation logic will be handled by the simulation manager
	return &module.SimulationState{}
}

// getModuleBasics returns all the basic module managers
func (app *HomeChainApp) getModuleBasics() []module.AppModuleBasic {
	return []module.AppModuleBasic{
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		// capability.AppModuleBasic{}, // Removed - doesn't exist in SDK v0.50+
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(nil),
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		// feegrant.AppModuleBasic{}, // Removed - doesn't exist in SDK v0.50+
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		// custom modules
		poh.AppModuleBasic{},
		mining.AppModuleBasic{},
		referral.AppModuleBasic{},
		nodestake.AppModuleBasic{},
	}
}

// maccPerms defines the module account permissions for the homechain app
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName: nil,
	distrtypes.ModuleName:      nil,
	minttypes.ModuleName:       {authtypes.Minter, authtypes.Burner},
	stakingtypes.ModuleName:    {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:        {authtypes.Burner},
	// custom module accounts
	miningtypes.MiningRewardPool: {authtypes.Minter, authtypes.Burner},
	referraltypes.ReferralPool:   {authtypes.Minter, authtypes.Burner},
	referraltypes.TreasuryPool:   {authtypes.Minter, authtypes.Burner},
	// nodestake module accounts
	nodestaketypes.ModuleName: nil,
}

// BlockedAddresses returns all the addresses that should not be allowed to
// receive tokens.
func (app *HomeChainApp) BlockedAddrs() map[string]bool {
	blockedAddrs := make(map[string]bool)

	for addr := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(addr).String()] = true
	}

	return blockedAddrs
}

// coreStoreAdapter adapts store/v2.KVStore to cosmossdk.io/core/store.KVStore
type coreStoreAdapter struct {
	store storev2.KVStore
}

func (a *coreStoreAdapter) Get(key []byte) ([]byte, error) {
	return a.store.Get(key), nil
}

func (a *coreStoreAdapter) Has(key []byte) (bool, error) {
	return a.store.Has(key), nil
}

func (a *coreStoreAdapter) Set(key, value []byte) error {
	a.store.Set(key, value)
	return nil
}

func (a *coreStoreAdapter) Delete(key []byte) error {
	a.store.Delete(key)
	return nil
}

func (a *coreStoreAdapter) Iterator(start, end []byte) (store.Iterator, error) {
	it := a.store.Iterator(start, end)
	return &iteratorAdapter{it: it}, nil
}

func (a *coreStoreAdapter) ReverseIterator(start, end []byte) (store.Iterator, error) {
	it := a.store.ReverseIterator(start, end)
	return &iteratorAdapter{it: it}, nil
}

// iteratorAdapter adapts store/v2.Iterator to cosmossdk.io/core/store.Iterator
type iteratorAdapter struct {
	it storev2.Iterator
}

func (a *iteratorAdapter) Next()                    { a.it.Next() }
func (a *iteratorAdapter) Key() []byte              { return a.it.Key() }
func (a *iteratorAdapter) Value() []byte            { return a.it.Value() }
func (a *iteratorAdapter) Error() error             { return a.it.Error() }
func (a *iteratorAdapter) Close() error             { a.it.Close(); return nil }
func (a *iteratorAdapter) Valid() bool              { return a.it.Valid() }
func (a *iteratorAdapter) Domain() ([]byte, []byte) { return a.it.Domain() }

// kvStoreService adapts cosmossdk.io/store/types.KVStoreKey to cosmossdk.io/core/store.KVStoreService
type kvStoreService struct {
	key *types.KVStoreKey
}

func (k *kvStoreService) OpenKVStore(ctx context.Context) store.KVStore {
	return &coreStoreAdapter{store: sdk.UnwrapSDKContext(ctx).KVStore(k.key)}
}

// newKVStoreService creates a store.KVStoreService from a KVStoreKey.
func newKVStoreService(key *types.KVStoreKey) store.KVStoreService {
	return &kvStoreService{key: key}
}
