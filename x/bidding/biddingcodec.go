package bidding

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var modcodec *codec.Codec

func init() {
	code := codec.New()
	CodecRegistration(code)
	codec.RegisterCrypto(code)
	modcodec = code.Seal()
}

func CodecRegistration(code *codec.Codec) {
	code.RegisterConcrete(MsgPlaceBid{}, "bidding/MsgPlaceBid", nil)
	code.RegisterInterface((*Auction)(nil), nil)
	code.RegisterConcrete(&ForwardAuction{}, "bidding/ForwardAuction", nil)
	code.RegisterConcrete(&ReverseAuction{}, "bidding/ReverseAuction", nil)
	code.RegisterConcrete(&ForwardReverseAuction{}, "bidding/ForwardReverseAuction", nil)
}
