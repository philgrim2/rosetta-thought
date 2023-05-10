// Copyright 2020 Coinbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/philgrim2/rosetta-thought/configuration"
	mocks "github.com/philgrim2/rosetta-thought/mocks/services"
	"github.com/philgrim2/rosetta-thought/thought"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
)

func forceHexDecode(t *testing.T, s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("could not decode hex %s", s)
	}

	return b
}

func forceMarshalMap(t *testing.T, i interface{}) map[string]interface{} {
	m, err := types.MarshalMap(i)
	if err != nil {
		t.Fatalf("could not marshal map %s", types.PrintStruct(i))
	}

	return m
}

func TestConstructionService(t *testing.T) {
	networkIdentifier = &types.NetworkIdentifier{
		Network:    thought.TestnetNetwork,
		Blockchain: thought.Blockchain,
	}

	cfg := &configuration.Configuration{
		Mode:     configuration.Online,
		Network:  networkIdentifier,
		Params:   thought.TestnetParams,
		Currency: thought.TestnetCurrency,
	}

	mockIndexer := &mocks.Indexer{}
	mockClient := &mocks.Client{}
	servicer := NewConstructionAPIService(cfg, mockClient, mockIndexer)
	ctx := context.Background()

	// Test Derive
	publicKey := &types.PublicKey{
		Bytes: forceHexDecode(
			t,
			"039ec9a2265b552b81b0552e6e0d58925cc38c1264ab9828e8c5f071b7dc3d262d",
		),
		CurveType: types.Secp256k1,
	}
	deriveResponse, err := servicer.ConstructionDerive(ctx, &types.ConstructionDeriveRequest{
		NetworkIdentifier: networkIdentifier,
		PublicKey:         publicKey,
	})
	assert.Nil(t, err)
	assert.Equal(t, &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: "kvdPDVw6T6ws8N2fAZiaFMHsJLXWDXtHiq",
		},
	}, deriveResponse)

	// Test Preprocess
	ops := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 0, //Might be possible to remove this later
			},
			Type: thought.InputOpType,
			Account: &types.AccountIdentifier{
				Address: "kyw8MaocLYCniZ3NnJqNST3qtZNygLSiCC",
			},
			Amount: &types.Amount{
				Value:    "-40000",
				Currency: thought.TestnetCurrency,
			},
			CoinChange: &types.CoinChange{
				CoinIdentifier: &types.CoinIdentifier{
					Identifier: "5d7ffb8cf555d87a9524d26d5b2f49570ad1b62fd58bcc391ebe8a469ce1da7f:0",
				},
				CoinAction: types.CoinSpent,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 0,
			},
			Type: thought.OutputOpType,
			Account: &types.AccountIdentifier{
				Address: "m92udt8YzZ3B2WZ4uzjuL5sdaQuNnLM8KU",
			},
			Amount: &types.Amount{
				Value:    "38000",
				Currency: thought.TestnetCurrency,
			},
		},
	}
	feeMultiplier := float64(0.75)
	preprocessResponse, err := servicer.ConstructionPreprocess(
		ctx,
		&types.ConstructionPreprocessRequest{
			NetworkIdentifier:      networkIdentifier,
			Operations:             ops,
			SuggestedFeeMultiplier: &feeMultiplier,
		},
	)
	assert.Nil(t, err)
	options := &preprocessOptions{
		Coins: []*types.Coin{
			{
				CoinIdentifier: &types.CoinIdentifier{
					Identifier: "5d7ffb8cf555d87a9524d26d5b2f49570ad1b62fd58bcc391ebe8a469ce1da7f:0",
				},
				Amount: &types.Amount{
					Value:    "-40000",
					Currency: thought.TestnetCurrency,
				},
			},
		},
		EstimatedSize: 114, // Change this later
		FeeMultiplier: &feeMultiplier,
	}
	assert.Equal(t, &types.ConstructionPreprocessResponse{
		Options: forceMarshalMap(t, options),
	}, preprocessResponse)

	// Test Metadata
	metadata := &constructionMetadata{
		ScriptPubKeys: []*thought.ScriptPubKey{
			{
				ASM:          "OP_DUP OP_HASH160 b19e5c5433afbf7aca8a73949a48fa6b41a1089d OP_EQUALVERIFY OP_CHECKSIG",
				Hex:          "76a914b19e5c5433afbf7aca8a73949a48fa6b41a1089d88ac",
				RequiredSigs: 1,
				Type:         "pubkeyhash",
				Addresses: []string{
					"m92udt8YzZ3B2WZ4uzjuL5sdaQuNnLM8KU",
				},
			},
		},
	}

	// Normal Fee
	mockIndexer.On(
		"GetScriptPubKeys",
		ctx,
		options.Coins,
	).Return(
		metadata.ScriptPubKeys,
		nil,
	).Once()
	mockClient.On(
		"SuggestedFeeRate",
		ctx,
		defaultConfirmationTarget,
	).Return(
		thought.MinFeeRate*10,
		nil,
	).Once()
	metadataResponse, err := servicer.ConstructionMetadata(ctx, &types.ConstructionMetadataRequest{
		NetworkIdentifier: networkIdentifier,
		Options:           forceMarshalMap(t, options),
	})
	assert.Nil(t, err)
	assert.Equal(t, &types.ConstructionMetadataResponse{
		Metadata: forceMarshalMap(t, metadata),
		SuggestedFee: []*types.Amount{
			{
				Value:    "855", // Describe how fee is calculated in notions
				Currency: thought.TestnetCurrency,
			},
		},
	}, metadataResponse)

	// Low Fee
	mockIndexer.On(
		"GetScriptPubKeys",
		ctx,
		options.Coins,
	).Return(
		metadata.ScriptPubKeys,
		nil,
	).Once()
	mockClient.On(
		"SuggestedFeeRate",
		ctx,
		defaultConfirmationTarget,
	).Return(
		thought.MinFeeRate,
		nil,
	).Once()
	metadataResponse, err = servicer.ConstructionMetadata(ctx, &types.ConstructionMetadataRequest{
		NetworkIdentifier: networkIdentifier,
		Options:           forceMarshalMap(t, options),
	})
	assert.Nil(t, err)
	assert.Equal(t, &types.ConstructionMetadataResponse{
		Metadata: forceMarshalMap(t, metadata),
		SuggestedFee: []*types.Amount{
			{
				Value:    "114", // we don't go below minimum fee rate change this later
				Currency: thought.TestnetCurrency,
			},
		},
	}, metadataResponse)

	// Test Payloads
	unsignedRaw := "7b227472616e73616374696f6e223a223032303030303030303137666461653139633436386162653165333963633862643532666236643130613537343932663562366464323234393537616438353566353863666237663564303030303030303030306666666666666666303137303934303030303030303030303030313937366139313462313965356335343333616662663761636138613733393439613438666136623431613130383964383861633030303030303030222c227363726970745075624b657973223a5b7b2261736d223a224f505f445550204f505f484153483136302062313965356335343333616662663761636138613733393439613438666136623431613130383964204f505f455155414c564552494659204f505f434845434b534947222c22686578223a223736613931346231396535633534333361666266376163613861373339343961343866613662343161313038396438386163222c2272657153696773223a312c2274797065223a227075626b657968617368222c22616464726573736573223a5b226d393275647438597a5a334232575a34757a6a754c3573646151754e6e4c4d384b55225d7d5d2c22696e7075745f616d6f756e7473223a5b222d3430303030225d2c22696e7075745f616464726573736573223a5b226b7977384d616f634c59436e695a334e6e4a714e53543371745a4e79674c53694343225d7d" // nolint
	payloadsResponse, err := servicer.ConstructionPayloads(ctx, &types.ConstructionPayloadsRequest{
		NetworkIdentifier: networkIdentifier,
		Operations:        ops,
		Metadata:          forceMarshalMap(t, metadata),
	})

	parseOps := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 0,
			},
			Type: thought.InputOpType,
			Account: &types.AccountIdentifier{
				Address: "kyw8MaocLYCniZ3NnJqNST3qtZNygLSiCC",
			},
			Amount: &types.Amount{
				Value:    "-40000",
				Currency: thought.TestnetCurrency,
			},
			CoinChange: &types.CoinChange{
				CoinIdentifier: &types.CoinIdentifier{
					Identifier: "5d7ffb8cf555d87a9524d26d5b2f49570ad1b62fd58bcc391ebe8a469ce1da7f:0",
				},
				CoinAction: types.CoinSpent,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 0,
			},
			Type: thought.OutputOpType,
			Account: &types.AccountIdentifier{
				Address: "m92udt8YzZ3B2WZ4uzjuL5sdaQuNnLM8KU",
			},
			Amount: &types.Amount{
				Value:    "38000",
				Currency: thought.TestnetCurrency,
			},
		},
	}

	assert.Nil(t, err)
	signingPayload := &types.SigningPayload{
		Bytes: forceHexDecode(
			t,
			"b6aa747c4dbe4e0397da142c28aabd326e08ce9b0ce8fd5afc3c5840f3f41b05",
		),
		AccountIdentifier: &types.AccountIdentifier{
			Address: "kyw8MaocLYCniZ3NnJqNST3qtZNygLSiCC",
		},
		SignatureType: types.Ecdsa,
	}
	assert.Equal(t, &types.ConstructionPayloadsResponse{
		UnsignedTransaction: unsignedRaw,
		Payloads:            []*types.SigningPayload{signingPayload},
	}, payloadsResponse)

	// Test Parse Unsigned
	parseUnsignedResponse, err := servicer.ConstructionParse(ctx, &types.ConstructionParseRequest{
		NetworkIdentifier: networkIdentifier,
		Signed:            false,
		Transaction:       unsignedRaw,
	})
	assert.Nil(t, err)
	assert.Equal(t, &types.ConstructionParseResponse{
		Operations:               parseOps,
		AccountIdentifierSigners: []*types.AccountIdentifier{},
	}, parseUnsignedResponse)

	// Test Combine
	signedRaw := "7b227472616e73616374696f6e223a22303230303030303030313766646165313963343638616265316533396363386264353266623664313061353734393266356236646432323439353761643835356635386366623766356430303030303030303662343833303435303232313030393132376132663731633332356534376234313139653239386335633438366131626266303833336334346663343732636138323961663636316566316531333032323036346166646331633464353534343637373232656661623130356263633466376661663061393463663530323466646664643630616639336266383863636336303132313032636661333538356261353934303839393838303839326663353037643233616232633739626438663561653430303339643130653734356464363035303862666666666666666666303137303934303030303030303030303030313937366139313462313965356335343333616662663761636138613733393439613438666136623431613130383964383861633030303030303030222c22696e7075745f616d6f756e7473223a5b222d3430303030225d7d" // nolint
	combineResponse, err := servicer.ConstructionCombine(ctx, &types.ConstructionCombineRequest{
		NetworkIdentifier:   networkIdentifier,
		UnsignedTransaction: unsignedRaw,
		Signatures: []*types.Signature{
			{
				Bytes: forceHexDecode(
					t,
					"30450221009127a2f71c325e47b4119e298c5c486a1bbf0833c44fc472ca829af661ef1e13022064afdc1c4d554467722efab105bcc4f7faf0a94cf5024fdfdd60af93bf88ccc602cfa3585ba5940899880892fc507d23ab2c79bd8f5ae40039d10e745dd60508bf", // nolint
				),
				SigningPayload: signingPayload,
				PublicKey:      publicKey,
				SignatureType:  types.Ecdsa,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, &types.ConstructionCombineResponse{
		SignedTransaction: signedRaw,
	}, combineResponse)

	// Test Parse Signed
	parseSignedResponse, err := servicer.ConstructionParse(ctx, &types.ConstructionParseRequest{
		NetworkIdentifier: networkIdentifier,
		Signed:            true,
		Transaction:       signedRaw,
	})
	assert.Nil(t, err)
	assert.Equal(t, &types.ConstructionParseResponse{
		Operations: parseOps,
		AccountIdentifierSigners: []*types.AccountIdentifier{
			{Address: "kyw8MaocLYCniZ3NnJqNST3qtZNygLSiCC"},
		},
	}, parseSignedResponse)

	// Test Hash
	transactionIdentifier := &types.TransactionIdentifier{
		Hash: "11cabe81d421dd4f97c11e79850e66c90df75130195ff836c5f372452801390e",
	}
	hashResponse, err := servicer.ConstructionHash(ctx, &types.ConstructionHashRequest{
		NetworkIdentifier: networkIdentifier,
		SignedTransaction: signedRaw,
	})
	assert.Nil(t, err)
	assert.Equal(t, &types.TransactionIdentifierResponse{
		TransactionIdentifier: transactionIdentifier,
	}, hashResponse)

	// Test Submit
	thoughtTransaction := "02000000017fdae19c468abe1e39cc8bd52fb6d10a57492f5b6dd224957ad855f58cfb7f5d000000006b4830450221009127a2f71c325e47b4119e298c5c486a1bbf0833c44fc472ca829af661ef1e13022064afdc1c4d554467722efab105bcc4f7faf0a94cf5024fdfdd60af93bf88ccc6012102cfa3585ba5940899880892fc507d23ab2c79bd8f5ae40039d10e745dd60508bfffffffff0170940000000000001976a914b19e5c5433afbf7aca8a73949a48fa6b41a1089d88ac00000000" // nolint
	mockClient.On(
		"SendRawTransaction",
		ctx,
		thoughtTransaction,
	).Return(
		transactionIdentifier.Hash,
		nil,
	)
	submitResponse, err := servicer.ConstructionSubmit(ctx, &types.ConstructionSubmitRequest{
		NetworkIdentifier: networkIdentifier,
		SignedTransaction: signedRaw,
	})
	assert.Nil(t, err)
	assert.Equal(t, &types.TransactionIdentifierResponse{
		TransactionIdentifier: transactionIdentifier,
	}, submitResponse)

	mockClient.AssertExpectations(t)
	mockIndexer.AssertExpectations(t)
}
