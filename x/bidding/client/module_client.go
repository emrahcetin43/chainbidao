package client

import (
	auctioncmd "bidao/x/bidding/client/cli"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
)

type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	auctionQueryCmd := &cobra.Command{
		Use:   "bid",
		Short: "commands for bidding",
	}

	auctionQueryCmd.AddCommand(client.GetCommands(
		auctioncmd.GetCmdGetAuctions(mc.storeKey, mc.cdc),
	)...)

	return auctionQueryCmd
}

func (mc ModuleClient) GetTxCmd() *cobra.Command {
	auctionTxCmd := &cobra.Command{
		Use:   "bid",
		Short: "tx subcommands for bids",
	}

	auctionTxCmd.AddCommand(client.PostCommands(
		auctioncmd.GetCmdPlaceBid(mc.cdc),
	)...)

	return auctionTxCmd
}
