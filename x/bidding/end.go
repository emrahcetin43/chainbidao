package bidding

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, k Keeper) sdk.Tags {

	expiredAuctions := k.getIter(ctx, end(ctx.BlockHeight()))
	defer expiredAuctions.Close()

	for ; expiredAuctions.Valid(); expiredAuctions.Next() {
		var auctionID ID
		k.code.MustUnmarshalBinaryLengthPrefixed(expiredAuctions.Value(), &auctionID)

		err := k.AClosing(ctx, auctionID)
		if err != nil {
			panic(err)
		}
	}

	return sdk.Tags{}
}
