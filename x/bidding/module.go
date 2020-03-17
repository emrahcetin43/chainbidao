package bidding

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ sdk.AppModule      = AppModule{}
	_ sdk.AppModuleBasic = AppModuleBasic{}
)

const ModuleName = "bidding"

type AppModuleBasic struct{}

var _ sdk.AppModuleBasic = AppModuleBasic{}

func (AppModuleBasic) Name() string {
	return ModuleName
}

func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	CodecRegistration(cdc)
}

func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return modcodec.MustMarshalJSON(DefaultGenesisState())
}

func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	return nil
}

type AppModule struct {
	AppModuleBasic
	keeper Keeper
}

func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

func (AppModule) Name() string {
	return ModuleName
}

func (AppModule) RegisterInvariants(_ sdk.InvariantRouter) {}

func (AppModule) Route() string {
	return ModuleName
}

func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

func (AppModule) QuerierRoute() string {
	return ModuleName
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return modcodec.MustMarshalJSON(gs)
}

func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) sdk.Tags {
	return sdk.EmptyTags()
}

func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) ([]abci.ValidatorUpdate, sdk.Tags) {
	tags := EndBlocker(ctx, am.keeper)
	return []abci.ValidatorUpdate{}, tags
}
