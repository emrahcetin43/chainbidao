package bidding

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	getAllAuctions = "biddings"
)

type ResultAuctions []string

func queryAllAuctions(request abci.RequestQuery, context sdk.Context, keeper Keeper) (result []byte, error sdk.Error) {
	var List ResultAuctions

	iter := keeper.GetAIter(context)

	for ; iter.Valid(); iter.Next() {

		var auction Auction
		keeper.code.MustUnmarshalBinaryBare(iter.Value(), &auction)
		List = append(List, auction.String())
	}

	b, err := codec.MarshalJSONIndent(keeper.code, List)
	if err != nil {
		panic("cannot marshal" +

			"")
	}

	return b, nil
}

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(context sdk.Context, p []string, request abci.RequestQuery) (result []byte, error sdk.Error) {
		switch p[0] {
		case getAllAuctions:
			return queryAllAuctions(request, context, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("command unknown")
		}
	}
}

func (n ResultAuctions) String() string {
	return strings.Join(n[:], "\n")
}
