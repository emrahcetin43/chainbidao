package bidding

import sdk "github.com/cosmos/cosmos-sdk/types"

type MsgPlaceBid struct {
	AuctionID ID
	Bidder    sdk.AccAddress
	Bid       sdk.Coin
	Lot       sdk.Coin
}

func (msg MsgPlaceBid) Route() string { return "bidding" }

func (msg MsgPlaceBid) Type() string { return "place_bid" }

func (msg MsgPlaceBid) ValidateBasic() sdk.Error {
	if msg.Bidder.Empty() {
		return sdk.ErrInternal("address cannot be empty")
	}
	if msg.Bid.Amount.LT(sdk.ZeroInt()) {
		return sdk.ErrInternal("amount cannot be negavite")
	}
	if msg.Lot.Amount.LT(sdk.ZeroInt()) {
		return sdk.ErrInternal("lot cannot be negative")
	}
	return nil
}

func (msg MsgPlaceBid) GetSignBytes() []byte {
	bz := modcodec.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgPlaceBid) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Bidder}
}
