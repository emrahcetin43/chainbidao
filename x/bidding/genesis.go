package bidding

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GenesisState struct {
}

func NewGenesisState() GenesisState {
	return GenesisState{}
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return NewGenesisState()
}
