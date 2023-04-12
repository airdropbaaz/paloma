package evm_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/palomachain/paloma/app"
	keepertest "github.com/palomachain/paloma/testutil/keeper"
	"github.com/palomachain/paloma/testutil/nullify"
	"github.com/palomachain/paloma/x/evm"
	"github.com/palomachain/paloma/x/evm/keeper"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keepertest.EvmKeeper(t)
	evm.InitGenesis(ctx, *k, genesisState)
	got := evm.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}

func TestGenesisGinkgo(t *testing.T) {
	RegisterFailHandler(g.Fail)

	g.RunSpecs(t, "EVM genssis")
}

var _ = g.Describe("genesis", func() {
	var genesisState types.GenesisState
	var k *keeper.Keeper
	var ctx sdk.Context
	var a app.TestApp

	g.BeforeEach(func() {
		t := g.GinkgoT()
		a = app.NewTestApp(t, false)
		ctx = a.NewContext(false, tmproto.Header{
			Height: 5,
		})
		k = &a.EvmKeeper
	})

	g.BeforeEach(func() {
		genesisState = types.GenesisState{
			Params: types.DefaultParams(),
		}
	})

	g.Context("init and export", func() {
		g.DescribeTable(
			"init genesis",
			func(chainInfo []*types.GenesisChainInfo, smartContract *types.GenesisSmartContract) {
				genesisState.Chains = chainInfo
				genesisState.SmartContract = smartContract
				evm.InitGenesis(ctx, *k, genesisState)

				if smartContract != nil {
					for _, ci := range chainInfo {
						sc, err := k.GetLastSmartContract(ctx)
						Expect(err).To(BeNil())
						err = k.ActivateChainReferenceID(
							ctx,
							ci.ChainReferenceID,
							sc,
							"0x1234",
							[]byte("abc"),
						)
						Expect(err).To(BeNil())
					}
				}
				got := evm.ExportGenesis(ctx, *k)

				Expect(chainInfo).To(Equal(got.Chains))
				Expect(smartContract).To(Equal(got.SmartContract))
			},
			g.Entry(
				"it returns an empty genesis",
				nil,
				nil,
			),
			g.Entry(
				"with chains and smart contract it exports it back",
				[]*types.GenesisChainInfo{
					{
						ChainReferenceID:  "eth-main",
						ChainID:           1,
						BlockHeight:       123,
						BlockHashAtHeight: "0x1234",
						MinOnChainBalance: "555",
					},
					{
						ChainReferenceID:  "ropsten",
						ChainID:           3,
						BlockHeight:       124,
						BlockHashAtHeight: "0x5555",
						MinOnChainBalance: "555",
					},
				},
				&types.GenesisSmartContract{
					AbiJson:     "[]",
					BytecodeHex: "0x1234",
				},
			),
		)
	})

	g.Context("invalid minOnChainBalance", func() {
		g.It("panics if the balance is invalid", func() {
			genesisState.Chains = []*types.GenesisChainInfo{
				{
					ChainReferenceID:  "eth-main",
					ChainID:           1,
					BlockHeight:       123,
					BlockHashAtHeight: "0x1234",
					MinOnChainBalance: "123invalid",
				},
			}
			Expect(func() {
				evm.InitGenesis(ctx, *k, genesisState)
			}).Should(Panic())
		})
	})

	g.When("chain is not active", func() {
		g.It("does not include it the export", func() {
			genesisState.Chains = []*types.GenesisChainInfo{
				{
					ChainReferenceID:  "eth-main",
					ChainID:           1,
					BlockHeight:       123,
					BlockHashAtHeight: "0x1234",
					MinOnChainBalance: "555",
				},
			}

			evm.InitGenesis(ctx, *k, genesisState)
			got := evm.ExportGenesis(ctx, *k)
			Expect(got.Chains).To(BeEmpty())
		})
	})

	g.When("there are no chains, but smart contract exists", func() {
		g.It("returns a smart contract anyway", func() {
			genesisState.SmartContract = &types.GenesisSmartContract{
				AbiJson:     "[]",
				BytecodeHex: "0x1234",
			}

			evm.InitGenesis(ctx, *k, genesisState)
			got := evm.ExportGenesis(ctx, *k)
			Expect(got.Chains).To(BeEmpty())
			Expect(got.SmartContract).To(Equal(genesisState.SmartContract))
		})
	})
})

func TestGenesisChainInfo(t *testing.T) {
}
