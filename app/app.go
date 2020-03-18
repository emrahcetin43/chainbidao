package app

import (
	"io"
	"os"


	"bidao/x/bidding"
	"bidao/x/positions"
	"bidao/x/sellcollateral"
	"bidao/x/oracle"

	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/cosmos/cosmos-sdk/x/auth"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/auth/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

)

const (
	appName = "bidao"
)

var (
	CommandDir   = os.ExpandEnv("$HOME/.bidao")
	ValidatorDir = os.ExpandEnv("$HOME/.bidaoD")
	BasicManager sdk.ModuleBasicManager
)

func init() {
	BasicManager = sdk.NewModuleBasicManager(
		bidding.AppModuleBasic{},
		positions.AppModuleBasic{},
		sellcollateral.AppModuleBasic{},
		oracle.AppModule{},
		genaccounts.AppModuleBasic{},
		genutil.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.AppModuleBasic{},
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
	)
}


func CodecMaker() *codec.Codec {
	var cdc = codec.New()
	BasicManager.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

func AddBidToAddr() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("bid", "bid"+"pub")
	config.SetBech32PrefixForValidator("bid"+"val"+"oper", "bid"+"val"+"oper"+"pub")
	config.SetBech32PrefixForConsensusNode("bid"+"val"+"cons", "bid"+"val"+"cons"+"pub")
	config.Seal()
}

type BidaoApp struct {
	*bam.BaseApp
	codec *codec.Codec
	checkPeriod uint
	transDistriKey  *sdk.TransientStoreKey
	governanceKey   *sdk.KVStoreKey
	mainKey         *sdk.KVStoreKey
	accKey          *sdk.KVStoreKey
	stakeKey        *sdk.KVStoreKey
	transKeyStaking *sdk.TransientStoreKey
	slashKey        *sdk.KVStoreKey
	mintingKey      *sdk.KVStoreKey
	distriKey       *sdk.KVStoreKey
	oracleKey       *sdk.KVStoreKey
	feeCollectorKey *sdk.KVStoreKey
	parameterKey    *sdk.KVStoreKey
	transParamKey   *sdk.TransientStoreKey
	biddingKey      *sdk.KVStoreKey
	posKey          *sdk.KVStoreKey
	sellKey         *sdk.KVStoreKey
	crisisKeeper        crisis.Keeper
	paramsKeeper        params.Keeper
	feeCollectionKeeper auth.FeeCollectionKeeper
	bankKeeper          bank.Keeper
	accountKeeper       auth.AccountKeeper
	distrKeeper         distr.Keeper
	govKeeper           gov.Keeper
	stakingKeeper       staking.Keeper
	slashingKeeper      slashing.Keeper
	mintKeeper          mint.Keeper
	posKeeper      positions.Keeper
	sellcollKeeper sellcollateral.Keeper
	biddingKeeper  bidding.Keeper
	oracleKeeper   oracle.Keeper

	moduleManager *sdk.ModuleManager
}

func BidaoAppN(database dbm.DB, trace io.Writer, latestLoad bool, log log.Logger, checkPeriod uint, options ...func(*bam.BaseApp)) *BidaoApp {

	codec := CodecMaker()

	baseApp := bam.NewBaseApp(appName, log, database, auth.DefaultTxDecoder(codec), options...)
	baseApp.SetAppVersion(version.Version)
	baseApp.SetCommitMultiStoreTracer(trace)

	var bidaoApp = &BidaoApp{
		BaseApp:         baseApp,
		codec:           codec,
		checkPeriod:     checkPeriod,
		biddingKey:      sdk.NewKVStoreKey("bidding"),
		slashKey:        sdk.NewKVStoreKey(slashing.StoreKey),
		governanceKey:   sdk.NewKVStoreKey(gov.StoreKey),
		accKey:          sdk.NewKVStoreKey(auth.StoreKey),
		stakeKey:        sdk.NewKVStoreKey(staking.StoreKey),
		transKeyStaking: sdk.NewTransientStoreKey(staking.TStoreKey),
		feeCollectorKey: sdk.NewKVStoreKey(auth.FeeStoreKey),
		parameterKey:    sdk.NewKVStoreKey(params.StoreKey),
		transParamKey:   sdk.NewTransientStoreKey(params.TStoreKey),
		mintingKey:      sdk.NewKVStoreKey(mint.StoreKey),
		distriKey:       sdk.NewKVStoreKey(distr.StoreKey),
		transDistriKey:  sdk.NewTransientStoreKey(distr.TStoreKey),
		mainKey:         sdk.NewKVStoreKey(bam.MainStoreKey),

	}

	bidaoApp.paramsKeeper = params.NewKeeper(bidaoApp.codec, bidaoApp.parameterKey, bidaoApp.transParamKey, params.DefaultCodespace)
	authSubspace := bidaoApp.paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := bidaoApp.paramsKeeper.Subspace(bank.DefaultParamspace)
	stakingSubspace := bidaoApp.paramsKeeper.Subspace(staking.DefaultParamspace)
	mintSubspace := bidaoApp.paramsKeeper.Subspace(mint.DefaultParamspace)
	distrSubspace := bidaoApp.paramsKeeper.Subspace(distr.DefaultParamspace)
	slashingSubspace := bidaoApp.paramsKeeper.Subspace(slashing.DefaultParamspace)
	govSubspace := bidaoApp.paramsKeeper.Subspace(gov.DefaultParamspace)
	crisisSubspace := bidaoApp.paramsKeeper.Subspace(crisis.DefaultParamspace)

	bidaoApp.accountKeeper = auth.NewAccountKeeper(bidaoApp.codec, bidaoApp.accKey, authSubspace, auth.ProtoBaseAccount)
	bidaoApp.bankKeeper = bank.NewBaseKeeper(bidaoApp.accountKeeper, bankSubspace, bank.DefaultCodespace)
	bidaoApp.feeCollectionKeeper = auth.NewFeeCollectionKeeper(bidaoApp.codec, bidaoApp.feeCollectorKey)
	stakingKeeper := staking.NewKeeper(bidaoApp.codec, bidaoApp.stakeKey, bidaoApp.transKeyStaking, bidaoApp.bankKeeper,
		stakingSubspace, staking.DefaultCodespace)
	bidaoApp.mintKeeper = mint.NewKeeper(bidaoApp.codec, bidaoApp.mintingKey, mintSubspace, &stakingKeeper, bidaoApp.feeCollectionKeeper)
	bidaoApp.distrKeeper = distr.NewKeeper(bidaoApp.codec, bidaoApp.distriKey, distrSubspace, bidaoApp.bankKeeper, &stakingKeeper,
		bidaoApp.feeCollectionKeeper, distr.DefaultCodespace)
	bidaoApp.slashingKeeper = slashing.NewKeeper(bidaoApp.codec, bidaoApp.slashKey, &stakingKeeper,
		slashingSubspace, slashing.DefaultCodespace)
	bidaoApp.crisisKeeper = crisis.NewKeeper(crisisSubspace, checkPeriod, bidaoApp.distrKeeper,
		bidaoApp.bankKeeper, bidaoApp.feeCollectionKeeper)

	bidaoApp.oracleKeeper = oracle.NewKeeper(bidaoApp.oracleKey, bidaoApp.codec, oracle.DefaultCodespace)
	bidaoApp.posKeeper = positions.NewKeeper(
		bidaoApp.codec,
		bidaoApp.posKey,
		cdpSubspace,
		bidaoApp.oracleKeeper,
		bidaoApp.bankKeeper,
	)
	bidaoApp.biddingKeeper = bidding.NewKeeper(
		bidaoApp.codec,
		bidaoApp.posKeeper,
		bidaoApp.biddingKey,
	)
	bidaoApp.sellcollKeeper = sellcollateral.NewKeeper(
		bidaoApp.codec,
		bidaoApp.sellKey,
		liquidatorSubspace,
		bidaoApp.posKeeper,
		bidaoApp.biddingKeeper,
		bidaoApp.posKeeper,
	)


	router := gov.NewRouter()
	router.AddRoute(gov.RouterKey, gov.ProposalHandler).
		AddRoute(params.RouterKey, params.NewParamChangeProposalHandler(bidaoApp.paramsKeeper)).
		AddRoute(distr.RouterKey, distr.NewCommunityPoolSpendProposalHandler(bidaoApp.distrKeeper))
	bidaoApp.govKeeper = gov.NewKeeper(bidaoApp.codec, bidaoApp.governanceKey, bidaoApp.paramsKeeper, govSubspace,
		bidaoApp.bankKeeper, &stakingKeeper, gov.DefaultCodespace, router)


	bidaoApp.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(bidaoApp.distrKeeper.Hooks(), bidaoApp.slashingKeeper.Hooks()))

	bidaoApp.moduleManager = sdk.NewModuleManager(
		genaccounts.NewAppModule(bidaoApp.accountKeeper),
		genutil.NewAppModule(bidaoApp.accountKeeper, bidaoApp.stakingKeeper, bidaoApp.BaseApp.DeliverTx),
		auth.NewAppModule(bidaoApp.accountKeeper, bidaoApp.feeCollectionKeeper),
		bank.NewAppModule(bidaoApp.bankKeeper, bidaoApp.accountKeeper),
		crisis.NewAppModule(bidaoApp.crisisKeeper, bidaoApp.Logger()),
		distr.NewAppModule(bidaoApp.distrKeeper),
		gov.NewAppModule(bidaoApp.govKeeper),
		mint.NewAppModule(bidaoApp.mintKeeper),
		slashing.NewAppModule(bidaoApp.slashingKeeper, bidaoApp.stakingKeeper),
		staking.NewAppModule(bidaoApp.stakingKeeper, bidaoApp.feeCollectionKeeper, bidaoApp.distrKeeper, bidaoApp.accountKeeper),
		bidding.NewAppModule(bidaoApp.biddingKeeper),
		positions.NewAppModule(bidaoApp.posKeeper),
		sellcollateral.NewAppModule(bidaoApp.sellcollKeeper),
		oracle.NewAppModule(bidaoApp.oracleKeeper),
	)

	bidaoApp.moduleManager.SetOrderEndBlockers(gov.ModuleName, staking.ModuleName, oracle.ModuleName)
	bidaoApp.moduleManager.SetOrderBeginBlockers(mint.ModuleName, distr.ModuleName, slashing.ModuleName)


	bidaoApp.moduleManager.SetOrderInitGenesis(genaccounts.ModuleName, distr.ModuleName,
		staking.ModuleName, auth.ModuleName, bank.ModuleName, slashing.ModuleName,
		gov.ModuleName, mint.ModuleName, crisis.ModuleName, genutil.ModuleName,
		bidding.ModuleName, positions.ModuleName, sellcollateral.ModuleName, oracle.ModuleName)

	bidaoApp.moduleManager.RegisterInvariants(&bidaoApp.crisisKeeper)
	bidaoApp.moduleManager.RegisterRoutes(bidaoApp.Router(), bidaoApp.QueryRouter())

	bidaoApp.MountStores(
		bidaoApp.mainKey,
		bidaoApp.accKey,
		bidaoApp.stakeKey,
		bidaoApp.mintingKey,
		bidaoApp.distriKey,
		bidaoApp.slashKey,
		bidaoApp.governanceKey,
		bidaoApp.feeCollectorKey,
		bidaoApp.parameterKey,
		bidaoApp.transParamKey,
		bidaoApp.transKeyStaking,
		bidaoApp.transDistriKey,
		bidaoApp.oracleKey,
		bidaoApp.biddingKey,
		bidaoApp.posKey,
		bidaoApp.sellKey,
	)

	bidaoApp.SetInitChainer(bidaoApp.BlockchainInitialize)
	bidaoApp.SetBeginBlocker(bidaoApp.ChainBegin)
	bidaoApp.SetAnteHandler(auth.NewAnteHandler(bidaoApp.accountKeeper, bidaoApp.feeCollectionKeeper, auth.DefaultSigVerificationGasConsumer))
	bidaoApp.SetEndBlocker(bidaoApp.ChainEnd)

	checkLatestLoad(latestLoad, bidaoApp)
	return bidaoApp
}




func checkLatestLoad(latestLoad bool, bidaoApp *BidaoApp) {
	if latestLoad {
		err := bidaoApp.LoadLatestVersion(bidaoApp.mainKey)
		if err != nil {
			cmn.Exit(err.Error())
		}
	}
}

func (app *BidaoApp) LoadingChainHeight(h int64) error {
	return app.LoadVersion(h, app.mainKey)
}

func (app *BidaoApp) ChainEnd(context sdk.Context, request abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.moduleManager.EndBlock(context, request)
}


func (app *BidaoApp) ChainBegin(context sdk.Context, request abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.moduleManager.BeginBlock(context, request)
}


func (app *BidaoApp) BlockchainInitialize(context sdk.Context, request abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenS
	app.codec.MustUnmarshalJSON(request.AppStateBytes, &genesisState)
	return app.moduleManager.InitGenesis(context, genesisState)
}

