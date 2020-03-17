package bidding

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var qPrefix = []byte("queue")
var kDel = []byte(":")

type Keeper struct {
	bankKeeper bankKeeper
	storeKey   sdk.StoreKey
	code       *codec.Codec
}

func NewKeeper(code *codec.Codec, bankKeeper bankKeeper, storeKey sdk.StoreKey) Keeper {
	return Keeper{
		bankKeeper: bankKeeper,
		storeKey:   storeKey,
		code:       code,
	}
}

func (k Keeper) BeginFRAuction(context sdk.Context, amount sdk.Coin, seller sdk.AccAddress, maxbid sdk.Coin, other sdk.AccAddress) (ID, sdk.Error) {
	initB := sdk.NewInt64Coin(maxbid.Denom, 0)
	auct, initO := NewFRAuction(seller, amount, initB, end(context.BlockHeight())+AUCTIONDURATION, maxbid, other)
	aID, error := k.BeginAuctioning(context, &auct, initO)
	if error != nil {
		return 0, error
	}
	return aID, nil
}

func (k Keeper) BeginAuctioning(context sdk.Context, auction Auction, initO bOut) (ID, sdk.Error) {
	newAID, error := k.getNextID(context)
	if error != nil {
		return 0, error
	}
	auction.SetID(newAID)

	_, error = k.bankKeeper.SubtractCoins(context, initO.Address, sdk.NewCoins(initO.Coin))
	if error != nil {
		return 0, error
	}

	k.ASetter(context, auction)
	k.incNextAID(context)
	return newAID, nil
}

func (k Keeper) BeginFAuction(seller sdk.AccAddress, context sdk.Context, amount sdk.Coin, init sdk.Coin) (ID, sdk.Error) {
	auct, initO := NewFAuction(seller, end(context.BlockHeight())+AUCTIONDURATION, amount, init)
	aID, error := k.BeginAuctioning(context, &auct, initO)

	if error != nil {
		return 0, error
	}
	return aID, nil
}

func (k Keeper) BeginRAuction(buyer sdk.AccAddress, context sdk.Context, bid sdk.Coin, initAmount sdk.Coin) (ID, sdk.Error) {
	auct, initO := NewRAuction(buyer, bid, initAmount, end(context.BlockHeight())+AUCTIONDURATION)
	aID, error := k.BeginAuctioning(context, &auct, initO)
	if error != nil {
		return 0, error
	}
	return aID, nil
}

func (k Keeper) BidPlacing(context sdk.Context, aID ID, bid sdk.Coin, bidder sdk.AccAddress, amount sdk.Coin) sdk.Error {

	auct, exists := k.AGetter(context, aID)
	if !exists {
		return sdk.ErrInternal("no bidding")
	}

	tokenO, tokenI, error := auct.setBid(end(context.BlockHeight()), bidder, amount, bid)
	if error != nil {
		return error
	}
	for _, out := range tokenO {
		_, error = k.bankKeeper.SubtractCoins(context, out.Address, sdk.NewCoins(out.Coin))
		if error != nil {
			panic(error)
		}
	}
	for _, input := range tokenI {
		_, error = k.bankKeeper.AddCoins(context, input.Address, sdk.NewCoins(input.Coin))
		if error != nil {
			panic(error)
		}
	}

	k.ASetter(context, auct)

	return nil
}

func (k Keeper) getNextID(context sdk.Context) (ID, sdk.Error) {
	storage := context.KVStore(k.storeKey)
	a := storage.Get(k.nextAIDk())
	if a == nil {
		a = k.code.MustMarshalBinaryLengthPrefixed(ID(0))
		storage.Set(k.nextAIDk(), a)

	}
	var aID ID
	k.code.MustUnmarshalBinaryLengthPrefixed(a, &aID)
	return aID, nil
}

func (k Keeper) incNextAID(context sdk.Context) sdk.Error {
	storage := context.KVStore(k.storeKey)
	a := storage.Get(k.nextAIDk())
	if a == nil {
		panic("no initial ID set")
	}
	var auctionID ID
	k.code.MustUnmarshalBinaryLengthPrefixed(a, &auctionID)

	a = k.code.MustMarshalBinaryLengthPrefixed(auctionID + 1)
	storage.Set(k.nextAIDk(), a)

	return nil
}

func (k Keeper) AClosing(context sdk.Context, aID ID) sdk.Error {

	auct, exists := k.AGetter(context, aID)
	if !exists {
		return sdk.ErrInternal("no bidding")
	}
	if context.BlockHeight() < int64(auct.GetEnd()) {
		return sdk.ErrInternal(fmt.Sprintf("Cannot close bidding %a < %a", context.BlockHeight(), auct.GetEnd()))
	}
	tokenI := auct.GetResult()
	_, error := k.bankKeeper.AddCoins(context, tokenI.Address, sdk.NewCoins(tokenI.Coin))
	if error != nil {
		return error
	}

	k.removeA(context, aID)

	return nil
}
func (k Keeper) ASetter(context sdk.Context, a Auction) {
	eA, exists := k.AGetter(context, a.GetID())
	if exists {
		k.delQueue(context, eA.GetEnd(), eA.GetID())
	}
	storage := context.KVStore(k.storeKey)
	b := k.code.MustMarshalBinaryLengthPrefixed(a)
	storage.Set(k.AKeyGetter(a.GetID()), b)
	k.addQueue(context, a.GetEnd(), a.GetID())
}

func (k Keeper) AGetter(context sdk.Context, aID ID) (Auction, bool) {
	var auct Auction
	storage := context.KVStore(k.storeKey)
	b := storage.Get(k.AKeyGetter(aID))
	if b == nil {
		return auct, false
	}
	k.code.MustUnmarshalBinaryLengthPrefixed(b, &auct)
	return auct, true
}

func (k Keeper) removeA(context sdk.Context, aID ID) {
	auct, exists := k.AGetter(context, aID)
	if exists {
		k.delQueue(context, auct.GetEnd(), aID)
	}
	storage := context.KVStore(k.storeKey)
	storage.Delete(k.AKeyGetter(aID))
}

func (k Keeper) nextAIDk() []byte {
	return []byte("nextID")
}
func (k Keeper) AKeyGetter(aID ID) []byte {
	return []byte(fmt.Sprintf("id:%d", aID))
}

func (k Keeper) addQueue(context sdk.Context, end end, aID ID) {
	storage := context.KVStore(k.storeKey)
	b := k.code.MustMarshalBinaryLengthPrefixed(aID)
	storage.Set(
		QElemKey(end, aID),
		b,
	)
}

func (k Keeper) GetAIter(context sdk.Context) sdk.Iterator {
	storage := context.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(storage, nil)
}

func (k Keeper) delQueue(context sdk.Context, end end, aID ID) {
	storage := context.KVStore(k.storeKey)
	storage.Delete(QElemKey(end, aID))
}

func (k Keeper) getIter(context sdk.Context, end end) sdk.Iterator {
	storage := context.KVStore(k.storeKey)
	return storage.Iterator(
		qPrefix,
		sdk.PrefixEndBytes(QElemKPrefix(end)),
	)
}

func QElemKey(end end, aID ID) []byte {
	return bytes.Join([][]byte{
		qPrefix,
		sdk.Uint64ToBigEndian(uint64(end)),
		sdk.Uint64ToBigEndian(uint64(aID)),
	}, kDel)
}

func QElemKPrefix(end end) []byte {
	return bytes.Join([][]byte{
		qPrefix,
		sdk.Uint64ToBigEndian(uint64(end)),
	}, kDel)
}
