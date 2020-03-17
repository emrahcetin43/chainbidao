package cli

import (
	"fmt"

	"bidao/x/bidding"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/spf13/cobra"
)

func GetCmdPlaceBid(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "participate [AuctionID] [Bidder] [Bid] [Lot]",
		Short: "participate in bidding",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}
			id, err := bidding.NewIDFromString(args[0])
			if err != nil {
				fmt.Printf("bidding id unknown -> %s \n", string(args[0]))
				return err
			}

			bid, err := sdk.ParseCoin(args[2])
			if err != nil {
				fmt.Printf("bid amount invalid -> %s \n", string(args[2]))
				return err
			}

			lot, err := sdk.ParseCoin(args[3])
			if err != nil {
				fmt.Printf("invalid lot -> %s \n", string(args[3]))
				return err
			}
			msg := bidding.NewMsgPlaceBid(id, cliCtx.GetFromAddress(), bid, lot)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			cliCtx.PrintResponse = true
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
