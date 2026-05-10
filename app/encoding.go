package app

import (
	"os"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/upgrade"

	// Custom modules
	"github.com/homechain/homechain/x/mining"
	"github.com/homechain/homechain/x/poh"
	"github.com/homechain/homechain/x/referral"
)

var (
	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration and genesis account
	// initialization.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distribution.AppModuleBasic{},
		gov.NewAppModuleBasic(nil),
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		// feegrant.AppModuleBasic{}, // Commented out - doesn't exist in SDK v0.50+
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		// Custom modules
		poh.AppModuleBasic{},
		mining.AppModuleBasic{},
		referral.AppModuleBasic{},
	)
)

// MakeEncodingConfig creates the encoding config for the application
func MakeEncodingConfig() EncodingConfig {
	amino := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         codec.NewProtoCodec(interfaceRegistry),
		TxConfig:          tx.NewTxConfig(codec.NewProtoCodec(interfaceRegistry), tx.DefaultSignModes),
		Amino:             amino,
	}
}

// EncodingConfig specifies the encoding configuration for the application.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Marshaler         codec.Codec
	TxConfig          sdkclient.TxConfig
	Amino             *codec.LegacyAmino
}

// AppName defines the application name
const AppName = "homechain"

// DefaultNodeHome defines the default home directory for the application
var DefaultNodeHome = os.ExpandEnv("$HOME/.homechain")
