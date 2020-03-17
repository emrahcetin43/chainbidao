package app

import (
	"encoding/json"
	"log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"


)


func (app *BidaoApp) noHeightPrepare(context sdk.Context, wl []string) {
	shouldWL := false

	if len(wl) > 0 {
		shouldWL = true
	}

	wlmap := make(map[string]bool)

	for _, addr := range wl {
		_, err := sdk.ValAddressFromBech32(addr)
		if err != nil {
			log.Fatal(err)
		}
		wlmap[addr] = true
	}

	app.crisisKeeper.AssertInvariants(context, app.Logger())

	app.stakingKeeper.IterateValidators(context, func(_ int64, val sdk.Validator) (stop bool) {
		_, _ = app.distrKeeper.WithdrawValidatorCommission(context, val.GetOperator())
		return false
	})

	delegates := app.stakingKeeper.GetAllDelegations(context)
	for _, delegation := range delegates {
		_, _ = app.distrKeeper.WithdrawDelegationRewards(context, delegation.DelegatorAddress, delegation.ValidatorAddress)
	}

	app.distrKeeper.DeleteAllValidatorSlashEvents(context)

	app.distrKeeper.DeleteAllValidatorHistoricalRewards(context)

	h := context.BlockHeight()

	context = context.WithBlockHeight(0)

	app.stakingKeeper.IterateValidators(context, func(_ int64, val sdk.Validator) (stop bool) {

		scraps := app.distrKeeper.GetValidatorOutstandingRewards(context, val.GetOperator())
		feePool := app.distrKeeper.GetFeePool(context)
		feePool.CommunityPool = feePool.CommunityPool.Add(scraps)
		app.distrKeeper.SetFeePool(context, feePool)

		app.distrKeeper.Hooks().AfterValidatorCreated(context, val.GetOperator())
		return false
	})

	for _, del := range delegates {
		app.distrKeeper.Hooks().BeforeDelegationCreated(context, del.DelegatorAddress, del.ValidatorAddress)
		app.distrKeeper.Hooks().AfterDelegationModified(context, del.DelegatorAddress, del.ValidatorAddress)
	}

	context = context.WithBlockHeight(h)

	app.stakingKeeper.IterateRedelegations(context, func(_ int64, red staking.Redelegation) (stop bool) {
		for i := range red.Entries {
			red.Entries[i].CreationHeight = 0
		}
		app.stakingKeeper.SetRedelegation(context, red)
		return false
	})

	app.stakingKeeper.IterateUnbondingDelegations(context, func(_ int64, ubd staking.UnbondingDelegation) (stop bool) {
		for i := range ubd.Entries {
			ubd.Entries[i].CreationHeight = 0
		}
		app.stakingKeeper.SetUnbondingDelegation(context, ubd)
		return false
	})

	store := context.KVStore(app.stakeKey)
	iterator := sdk.KVStoreReversePrefixIterator(store, staking.ValidatorsKey)
	c := int16(0)

	var valConsAddrs []sdk.ConsAddress
	for ; iterator.Valid(); iterator.Next() {
		addr := sdk.ValAddress(iterator.Key()[1:])
		validator, found := app.stakingKeeper.GetValidator(context, addr)
		if !found {
			panic("expected validator, not found")
		}

		validator.UnbondingHeight = 0
		valConsAddrs = append(valConsAddrs, validator.ConsAddress())
		if shouldWL && !wlmap[addr.String()] {
			validator.Jailed = true
		}

		app.stakingKeeper.SetValidator(context, validator)
		c++
	}

	iterator.Close()

	_ = app.stakingKeeper.ApplyAndReturnValidatorSetUpdates(context)


	app.slashingKeeper.IterateValidatorSigningInfos(
		context,
		func(addr sdk.ConsAddress, info slashing.ValidatorSigningInfo) (stop bool) {
			info.StartHeight = 0
			app.slashingKeeper.SetValidatorSigningInfo(context, addr, info)
			return false
		},
	)
}

func (app *BidaoApp) StateExport(iszeroh bool, wl []string, ) (vals []tmtypes.GenesisValidator, state json.RawMessage, error error) {

	context := app.NewContext(true, abci.Header{Height: app.LastBlockHeight()})

	if iszeroh {
		app.noHeightPrepare(context, wl)
	}

	genesisS := app.moduleManager.ExportGenesis(context)
	state, error = codec.MarshalJSONIndent(app.codec, genesisS)
	if error != nil {
		return nil, nil, error
	}
	vals = staking.WriteValidators(context, app.stakingKeeper)
	return vals, state, nil
}

