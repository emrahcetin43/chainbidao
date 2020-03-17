package main

import (
	"os"
	"path"

	"bidao/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	at "github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	auth "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	bank "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	crisisclient "github.com/cosmos/cosmos-sdk/x/crisis/client"
	distcmd "github.com/cosmos/cosmos-sdk/x/distribution"
	distClient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrcli "github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	dist "github.com/cosmos/cosmos-sdk/x/distribution/client/rest"
	gv "github.com/cosmos/cosmos-sdk/x/gov"
	govClient "github.com/cosmos/cosmos-sdk/x/gov/client"
	gov "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintclient "github.com/cosmos/cosmos-sdk/x/mint/client"
	mintrest "github.com/cosmos/cosmos-sdk/x/mint/client/rest"
	paramcli "github.com/cosmos/cosmos-sdk/x/params/client/cli"
	paramsrest "github.com/cosmos/cosmos-sdk/x/params/client/rest"
	sl "github.com/cosmos/cosmos-sdk/x/slashing"
	slashingclient "github.com/cosmos/cosmos-sdk/x/slashing/client"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/client/rest"
	st "github.com/cosmos/cosmos-sdk/x/staking"
	stakingclient "github.com/cosmos/cosmos-sdk/x/staking/client"
	staking "github.com/cosmos/cosmos-sdk/x/staking/client/rest"

	auctionclient "bidao/x/bidding/client"
	auctionrest "bidao/x/bidding/client/rest"
	priceclient "bidao/x/oracle/client"
	pricerest "bidao/x/oracle/client/rest"
	cdpclient "bidao/x/positions/client"
	cdprest "bidao/x/positions/client/rest"
	liquidatorclient "bidao/x/sellcollateral/client"
	liquidatorrest "bidao/x/sellcollateral/client/rest"

	_ "github.com/cosmos/gaia/cmd/gaiacli/statik"
)

func main() {
	cobra.EnableCommandSorting = false

	cdc := app.CodecMaker()

	app.AddBidToAddr()

	mc := []sdk.ModuleClient{
		govClient.NewModuleClient(gv.StoreKey, cdc, paramcli.GetCmdSubmitProposal(cdc), distrcli.GetCmdSubmitProposal(cdc)),
		distClient.NewModuleClient(distcmd.StoreKey, cdc),
		stakingclient.NewModuleClient(st.StoreKey, cdc),
		mintclient.NewModuleClient(mint.StoreKey, cdc),
		slashingclient.NewModuleClient(sl.StoreKey, cdc),
		crisisclient.NewModuleClient(sl.StoreKey, cdc),
		priceclient.NewModuleClient("oracle", cdc),
		cdpclient.NewModuleClient("positions", cdc),
		auctionclient.NewModuleClient("bidding", cdc),
		liquidatorclient.NewModuleClient("sellcollateral", cdc),
	}

	rootCmd := &cobra.Command{
		Use:   "bidao",
		Short: "Bidao commands",
	}

	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	rootCmd.AddCommand(
		rpc.StatusCommand(),
		client.ConfigCmd(app.CommandDir),
		queryCmd(cdc, mc),
		txCmd(cdc, mc),
		client.LineBreak,
		lcd.ServeCommand(cdc, registerRoutes),
		client.LineBreak,
		keys.Commands(),
		client.LineBreak,
		version.Cmd,
		client.NewCompletionCmd(rootCmd, true),
	)

	executor := cli.PrepareMainCmd(rootCmd, "bidao", app.CommandDir)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func queryCmd(cdc *amino.Codec, mc []sdk.ModuleClient) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "queries",
		Aliases: []string{"q"},
		Short:   "get subcommands",
	}

	queryCmd.AddCommand(
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		tx.SearchTxCmd(cdc),
		tx.QueryTxCmd(cdc),
		client.LineBreak,
		authcmd.GetAccountCmd(at.StoreKey, cdc),
	)

	for _, m := range mc {
		mQueryCmd := m.GetQueryCmd()
		if mQueryCmd != nil {
			queryCmd.AddCommand(mQueryCmd)
		}
	}

	return queryCmd
}

func txCmd(cdc *amino.Codec, mc []sdk.ModuleClient) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "transactions",
		Short: "transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		client.LineBreak,
		authcmd.GetSignCommand(cdc),
		authcmd.GetMultiSignCommand(cdc),
		tx.GetBroadcastCommand(cdc),
		tx.GetEncodeCommand(cdc),
		client.LineBreak,
	)

	for _, m := range mc {
		txCmd.AddCommand(m.GetTxCmd())
	}

	return txCmd
}

func registerRoutes(rs *lcd.RestServer) {
	rpc.RegisterRoutes(rs.CliCtx, rs.Mux)
	tx.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	auth.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, at.StoreKey)
	bank.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	dist.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, distcmd.StoreKey)
	staking.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	slashing.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	gov.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, paramsrest.ProposalRESTHandler(rs.CliCtx, rs.Cdc), dist.ProposalRESTHandler(rs.CliCtx, rs.Cdc))
	mintrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	pricerest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, "oracle")
	auctionrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	cdprest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	liquidatorrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
