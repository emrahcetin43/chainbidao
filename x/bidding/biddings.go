package bidding

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	AUCTIONDURATION end = 40000
	DURATION        end = 2000
)

type Auction interface {
	GetID() ID
	SetID(ID)
	setBid(bHeight end, bidder sdk.AccAddress, amount sdk.Coin, bid sdk.Coin) ([]bOut, []bInput, sdk.Error)
	GetEnd() end
	GetResult() bInput
	String() string
}

type BaseAuction struct {
	ID         ID
	Initiator  sdk.AccAddress
	Lot        sdk.Coin
	Bidder     sdk.AccAddress
	Bid        sdk.Coin
	EndTime    end
	MaxEndTime end
}

type ID uint64
type end int64

type bInput struct {
	Address sdk.AccAddress
	Coin    sdk.Coin
}
type bOut struct {
	Address sdk.AccAddress
	Coin    sdk.Coin
}

func (a BaseAuction) GetEnd() end { return a.EndTime }

func (a BaseAuction) GetResult() bInput {
	return bInput{a.Bidder, a.Lot}
}

func (a BaseAuction) GetID() ID { return a.ID }

func (a *BaseAuction) SetID(id ID) { a.ID = id }

func (e end) String() string {
	return string(e)
}

func (a BaseAuction) String() string {
	return fmt.Sprintf(`Auction %d:
  Init:              %s
  Amount:               			%s
  Bidder:            		  %s
  Bid:        						%s
  End:   						%s
  EndMax:      			%s`,
		a.GetID(), a.Initiator, a.Lot,
		a.Bidder, a.Bid, a.GetEnd().String(),
		a.MaxEndTime.String(),
	)
}

type ForwardAuction struct {
	BaseAuction
}

func NewFAuction(seller sdk.AccAddress, end end, amount sdk.Coin, initialBid sdk.Coin) (ForwardAuction, bOut) {
	auction := ForwardAuction{BaseAuction{
		Initiator:  seller,
		Lot:        amount,
		Bidder:     seller,
		Bid:        initialBid,
		EndTime:    end,
		MaxEndTime: end,
	}}
	output := bOut{seller, amount}
	return auction, output
}

func (a *ForwardAuction) setBid(bHeight end, bidder sdk.AccAddress, amount sdk.Coin, bid sdk.Coin) ([]bOut, []bInput, sdk.Error) {
	if bHeight > a.EndTime {
		return []bOut{}, []bInput{}, sdk.ErrInternal("bidding over")
	}
	if !a.Bid.IsLT(bid) {
		return []bOut{}, []bInput{}, sdk.ErrInternal("error")
	}
	outputs := []bOut{{bidder, bid}}
	inputs := []bInput{{a.Bidder, a.Bid}, {a.Initiator, bid.Sub(a.Bid)}}

	a.Bidder = bidder
	a.Bid = bid
	a.EndTime = end(min(int64(bHeight+DURATION), int64(a.MaxEndTime)))

	return outputs, inputs, nil
}

type ReverseAuction struct {
	BaseAuction
}

func NewRAuction(buyer sdk.AccAddress, bid sdk.Coin, initialAmount sdk.Coin, end end) (ReverseAuction, bOut) {
	auction := ReverseAuction{BaseAuction{
		Initiator:  buyer,
		Lot:        initialAmount,
		Bidder:     buyer,
		Bid:        bid,
		EndTime:    end,
		MaxEndTime: end,
	}}
	output := bOut{buyer, initialAmount}
	return auction, output
}

func (a *ReverseAuction) setBid(currentBlockHeight end, bidder sdk.AccAddress, lot sdk.Coin, bid sdk.Coin) ([]bOut, []bInput, sdk.Error) {

	if currentBlockHeight > a.EndTime {
		return []bOut{}, []bInput{}, sdk.ErrInternal("bidding ended")
	}
	if !lot.IsLT(a.Lot) {
		return []bOut{}, []bInput{}, sdk.ErrInternal("error")
	}
	outputs := []bOut{{bidder, a.Bid}}
	inputs := []bInput{{a.Bidder, a.Bid}, {a.Initiator, a.Lot.Sub(lot)}}

	a.Bidder = bidder
	a.Lot = lot
	a.EndTime = end(min(int64(currentBlockHeight+DURATION), int64(a.MaxEndTime)))

	return outputs, inputs, nil
}

type ForwardReverseAuction struct {
	BaseAuction
	MaxBid      sdk.Coin
	OtherPerson sdk.AccAddress
}

func (a ForwardReverseAuction) String() string {
	return fmt.Sprintf(`Auction %d:
  Initiator:              %s
  Amount:               			%s
  Bidder:            		  %s
  Bid:        						%s
  End:   						%s
	End Max:      			%s
	Max Bid									%s
	Other						%s`,
		a.GetID(), a.Initiator, a.Lot,
		a.Bidder, a.Bid, a.GetEnd().String(),
		a.MaxEndTime.String(), a.MaxBid, a.OtherPerson,
	)
}

func NewFRAuction(seller sdk.AccAddress, amount sdk.Coin, initB sdk.Coin, end end, maxBid sdk.Coin, other sdk.AccAddress) (ForwardReverseAuction, bOut) {
	auction := ForwardReverseAuction{
		BaseAuction: BaseAuction{
			Initiator:  seller,
			Lot:        amount,
			Bidder:     seller,
			Bid:        initB,
			EndTime:    end,
			MaxEndTime: end},
		MaxBid:      maxBid,
		OtherPerson: other,
	}
	output := bOut{seller, amount}
	return auction, output
}

func (a *ForwardReverseAuction) setBid(currentBlockHeight end, bidder sdk.AccAddress, lot sdk.Coin, bid sdk.Coin) (outputs []bOut, inputs []bInput, err sdk.Error) {
	if currentBlockHeight > a.EndTime {
		return []bOut{}, []bInput{}, sdk.ErrInternal("bidding over")
	}

	switch {
	case a.Bid.IsLT(a.MaxBid) && bid.IsLT(a.MaxBid):
		if !a.Bid.IsLT(bid) {
			return []bOut{}, []bInput{}, sdk.ErrInternal("error")
		}
		outputs = []bOut{{bidder, bid}}
		inputs = []bInput{{a.Bidder, a.Bid}, {a.Initiator, bid.Sub(a.Bid)}}
	case a.Bid.IsLT(a.MaxBid):
		if !bid.IsEqual(a.MaxBid) {
			return []bOut{}, []bInput{}, sdk.ErrInternal("error")
		}
		outputs = []bOut{{bidder, bid}}
		inputs = []bInput{
			{a.Bidder, a.Bid},
			{a.Initiator, bid.Sub(a.Bid)},
			{a.OtherPerson, a.Lot.Sub(lot)},
		}

	case a.Bid.IsEqual(a.MaxBid):
		if !lot.IsLT(a.Lot) {
			return []bOut{}, []bInput{}, sdk.ErrInternal("error")
		}
		outputs = []bOut{{bidder, a.Bid}}
		inputs = []bInput{{a.Bidder, a.Bid}, {a.OtherPerson, a.Lot.Sub(lot)}}
	default:
		panic("ERROR")
	}

	a.Bidder = bidder
	a.Lot = lot
	a.Bid = bid
	a.EndTime = end(min(int64(currentBlockHeight+DURATION), int64(a.MaxEndTime)))

	return outputs, inputs, nil
}
