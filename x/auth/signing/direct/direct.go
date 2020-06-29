package direct

import (
	"fmt"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type ProtoTx interface {
	GetBodyBytes() []byte
	GetAuthInfoBytes() []byte
}

type ModeHandler struct{}

func (h ModeHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_DIRECT
}

var _ signing.SignModeHandler = ModeHandler{}

func (ModeHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_DIRECT}
}

func (ModeHandler) GetSignBytes(mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx) ([]byte, error) {
	if mode != signingtypes.SignMode_SIGN_MODE_DIRECT {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_DIRECT, mode)
	}

	protoTx, ok := tx.(ProtoTx)
	if !ok {
		return nil, fmt.Errorf("can only get direct sign bytes for a ProtoTx, got %T", tx)
	}

	bodyBz := protoTx.GetBodyBytes()
	authInfoBz := protoTx.GetAuthInfoBytes()

	return SignBytes(bodyBz, authInfoBz, data.ChainID, data.AccountNumber, data.AccountSequence)
}

func SignBytes(bodyBz, authInfoBz []byte, chainID string, accnum, sequence uint64) ([]byte, error) {
	signDoc := types.SignDoc{
		BodyBytes:       bodyBz,
		AuthInfoBytes:   authInfoBz,
		ChainId:         chainID,
		AccountNumber:   accnum,
		AccountSequence: sequence,
	}
	return signDoc.Marshal()
}
