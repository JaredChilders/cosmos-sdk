package direct

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/tx/generator"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestDirectModeHandler(t *testing.T) {
	app := simapp.Setup(false)
	//ctx := app.BaseApp.NewContext(false, abci.Header{})

	_, pubkey, addr := authtypes.KeyTestPubAddr()
	cdc := std.DefaultPublicKeyCodec{}

	tx := generator.NewBuilder(app.AppCodec(), std.DefaultPublicKeyCodec{})
	memo := "sometestmemo"
	msgs := []sdk.Msg{authtypes.NewTestMsg(addr)}

	pk, err := cdc.Encode(pubkey)
	require.NoError(t, err)

	var signerInfo []*txtypes.SignerInfo
	signerInfo = append(signerInfo, &txtypes.SignerInfo{
		PublicKey: pk,
		ModeInfo: &txtypes.ModeInfo{
			Sum: &txtypes.ModeInfo_Single_{
				Single: &txtypes.ModeInfo_Single{
					Mode: signingtypes.SignMode_SIGN_MODE_DIRECT,
				},
			},
		},
	})

	sig := signingtypes.SignatureV2{
		PubKey: pubkey,
		Data: &signingtypes.SingleSignatureData{
			SignMode:  signingtypes.SignMode_SIGN_MODE_DIRECT,
			Signature: pubkey.Bytes(),
		},
	}

	fee := txtypes.Fee{Amount: sdk.NewCoins(sdk.NewInt64Coin("atom", 150)), GasLimit: 20000}

	err = tx.SetMsgs(msgs...)
	require.NoError(t, err)
	tx.SetMemo(memo)
	tx.SetFeeAmount(fee.Amount)
	tx.SetGasLimit(fee.GasLimit)

	err = tx.SetSignatures(sig)
	require.NoError(t, err)

	t.Log("verify modes and default-mode")
	var directModeHandler ModeHandler
	require.Equal(t, directModeHandler.DefaultMode(), signingtypes.SignMode_SIGN_MODE_DIRECT)
	require.Len(t, directModeHandler.Modes(), 1)

	signingData := signing.SignerData{
		ChainID:         "test-chain",
		AccountNumber:   1,
		AccountSequence: 1,
	}

	signBytes, err := directModeHandler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_DIRECT, signingData, tx)

	require.NoError(t, err)
	require.NotNil(t, signBytes)

	authInfo := &txtypes.AuthInfo{
		Fee:         &fee,
		SignerInfos: signerInfo,
	}

	authInfoBytes := app.AppCodec().MustMarshalBinaryBare(authInfo)

	anys := make([]*codectypes.Any, len(msgs))

	for i, msg := range msgs {
		var err error
		anys[i], err = codectypes.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}
	}

	txBody := &txtypes.TxBody{
		Memo:     memo,
		Messages: anys,
	}
	bodyBytes := app.AppCodec().MustMarshalBinaryBare(txBody)

	t.Log("verify GetSignBytes with generating sign bytes by marshaling SignDoc")
	signDoc := txtypes.SignDoc{
		AccountNumber:   1,
		AccountSequence: 1,
		AuthInfoBytes:   authInfoBytes,
		BodyBytes:       bodyBytes,
		ChainId:         "test-chain",
	}

	signDocBytes, err := signDoc.Marshal()
	require.NoError(t, err)
	require.Equal(t, signDocBytes, signBytes)

	t.Log("verify GetSignBytes with false txBody data")
	signDoc.BodyBytes = []byte("dfafdasfds")
	signDocBytes, err = signDoc.Marshal()
	require.NoError(t, err)
	require.NotEqual(t, signDocBytes, signBytes)
}
