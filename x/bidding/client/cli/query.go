package cli

import (
	"fmt"

	"bidao/x/bidding"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

func GetCmdGetAuctions(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "auctions",
		Short: "return currently active auctions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getauctions", queryRoute), nil)
			if err != nil {
				fmt.Printf("cannot return auctions because of %s", err)
				return nil
			}
			var out bidding.ResultAuctions
			cdc.MustUnmarshalJSON(res, &out)
			if len(out) == 0 {
				out = append(out, "No active auctions")
			}
			return cliCtx.PrintOutput(out)
		},
	}
}
