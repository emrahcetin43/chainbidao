package main

import (
	"encoding/json"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/x/auth/genaccounts"
	genaccscli "github.com/cosmos/cosmos-sdk/x/auth/genaccounts/client/cli"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	"bidao/app"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	flagOverwrite = "overwrite"
)

var invCheckPeriod uint

func main() {

	cdc := app.CodecMaker()
	app.AddBidToAddr()

	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "bidaoD",
		Short:             "Bidao Chain Service",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(genutilcli.InitCmd(ctx, cdc, app.BasicManager, app.ValidatorDir))
	rootCmd.AddCommand(genutilcli.CollectGenTxsCmd(ctx, cdc, genaccounts.AppModuleBasic{}, app.ValidatorDir))
	rootCmd.AddCommand(genutilcli.GenTxCmd(ctx, cdc, app.BasicManager, genaccounts.AppModuleBasic{}, app.ValidatorDir, app.CommandDir))
	rootCmd.AddCommand(genutilcli.ValidateGenesisCmd(ctx, cdc, app.BasicManager))
	rootCmd.AddCommand(genaccscli.AddGenesisAccountCmd(ctx, cdc, app.ValidatorDir, app.CommandDir))
	rootCmd.AddCommand(client.NewCompletionCmd(rootCmd, true))

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	executor := cli.PrepareBaseCmd(rootCmd, "Bidao", app.ValidatorDir)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.BidaoAppN(
		db, traceStore, true, logger, invCheckPeriod,
		baseapp.SetPruning(store.NewPruningOptionsFromString(viper.GetString("pruning"))),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
		baseapp.SetHaltHeight(uint64(viper.GetInt(server.FlagHaltHeight))),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		gApp := app.BidaoAppN(db, traceStore, false, logger, uint(1))
		err := gApp.LoadingChainHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return gApp.StateExport(forZeroHeight, jailWhiteList)
	}
	gApp := app.BidaoAppN(db, traceStore, true, logger, uint(1))
	return gApp.StateExport(forZeroHeight, jailWhiteList)
}
