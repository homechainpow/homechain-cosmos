package cmd

import (
	"os"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	log "cosmossdk.io/log/v2"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	"github.com/homechain/homechain/app"
)

// NewRootCmd creates a new root command for homechaind
func NewRootCmd() *cobra.Command {
	// Initialize SDK configuration
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("home", "homepub")
	cfg.SetBech32PrefixForValidator("homevaloper", "homevaloperpub")
	cfg.SetBech32PrefixForConsensusNode("homevalcons", "homevalconspub")
	cfg.Seal()

	// Initialize encoding configuration
	encodingConfig := app.MakeEncodingConfig()
	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("")

	rootCmd := &cobra.Command{
		Use:   "homechaind",
		Short: "HomeChain Daemon - Bitcoin-Grade Stability with EVM Compatibility",
		Long: `HomeChain is a Layer-1 blockchain with:
- Proof of Hashrate (PoH) consensus using Argon2id
- EVM compatibility via Ethermint
- Bitcoin-grade stability with adaptive difficulty
- Dual staking (PoS + PoH) and referral rewards`,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			// Initialize the client context
			initClientCtx, err := config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			// Set the client context
			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			return nil
		},
	}

	// Add flags
	rootCmd.PersistentFlags().StringP(flags.FlagHome, "", app.DefaultNodeHome, "directory for config and data")
	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")
	rootCmd.PersistentFlags().String(flags.FlagKeyringBackend, "test", "Select keyring's backend (os|file|kwallet|pass|test|memory)")
	rootCmd.PersistentFlags().Int64(flags.FlagHeight, 0, "Use a specific height to query state at")

	// Add commands
	// TODO: Re-enable genutil commands when v0.50 API is fully supported
	rootCmd.AddCommand(
		genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		AddGenesisAccountCmd(app.DefaultNodeHome),
		NewTestnetCmd(app.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
		NewVersionCmd(),
	)

	// Add server commands
	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, exportAppStateAndTMValidators, addModuleInitFlags)

	// Add telemetry command
	rootCmd.AddCommand(NewTelemetryCmd())

	// Add PoH commands
	rootCmd.AddCommand(NewPoHCmd())

	// Add mining commands
	rootCmd.AddCommand(NewMiningCmd())

	return rootCmd
}

// newApp creates the application
func newApp(logger log.Logger, db dbm.DB, appOpts servertypes.AppOptions) servertypes.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)
	return app.NewHomeChainApp(
		logger, db, nil, true,
		skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		appOpts,
		[]string{}, // enabledProposals
		baseappOptions...,
	)
}

// exportAppStateAndTMValidators exports the application state and validators
func exportAppStateAndTMValidators(
	logger log.Logger,
	db dbm.DB,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	var homechainApp *app.HomeChainApp
	if height != -1 {
		homechainApp = app.NewHomeChainApp(logger, db, nil, false, skipUpgradeHeights, "", 0, appOpts, []string{})
		if err := homechainApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		homechainApp = app.NewHomeChainApp(logger, db, nil, true, skipUpgradeHeights, "", 0, appOpts, []string{})
	}

	return homechainApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// addModuleInitFlags adds module specific init flags
func addModuleInitFlags(startCmd *cobra.Command) {
	// Add any module-specific flags here
}

// skipUpgradeHeights is a map of heights to skip for upgrades
var skipUpgradeHeights = make(map[int64]bool)

// flags package aliases to avoid import cycle
const (
	flagHome           = "home"
	flagChainID        = "chain-id"
	flagKeyringBackend = "keyring-backend"
	flagHeight         = "height"
)
