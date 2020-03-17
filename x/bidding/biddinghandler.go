package bidding

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BidPlacingHandle(context sdk.Context, msg MsgPlaceBid, keeper Keeper) sdk.Result {

	err := keeper.BidPlacing(context, msg.AuctionID, msg.Bid, msg.Bidder, msg.Lot)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{}
}

func NewHandler(keeper Keeper) sdk.Handler {
	return func(context sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgPlaceBid:
			return BidPlacingHandle(context, msg, keeper)
		default:
			error := fmt.Sprintf("unknown Type: %T", msg)
			return sdk.ErrUnknownRequest(error).Result()
		}
	}
}
