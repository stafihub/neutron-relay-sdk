package client

import (
	"fmt"

	xWasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	clientTx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xAuthClient "github.com/cosmos/cosmos-sdk/x/auth/client"
	xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/spf13/cobra"
	"github.com/stafihub/neutron-relay-sdk/common/core"
)

func (c *Client) SingleTransferTo(toAddr types.AccAddress, amount types.Coins) error {
	done := core.UseSdkConfigContext(c.GetAccountPrefix())
	defer done()

	msg := xBankTypes.NewMsgSend(c.Ctx().GetFromAddress(), toAddr, amount)
	cmd := cobra.Command{}
	return clientTx.GenerateOrBroadcastTxCLI(c.Ctx(), cmd.Flags(), msg)
}

func (c *Client) SendContractExecuteMsg(contract string, msg []byte, amount types.Coins) (string, error) {
	msgs := []types.Msg{
		&xWasmTypes.MsgExecuteContract{
			Sender:   c.clientCtx.FromAddress.String(),
			Contract: contract,
			Msg:      msg,
			Funds:    amount,
		},
	}

	txbts, err := c.ConstructAndSignTx(msgs...)
	if err != nil {
		return "", err
	}

	txHash, err := c.BroadcastTx(txbts)
	if err != nil {
		return "", err
	}

	return txHash, nil
}

func (c *Client) BroadcastTx(tx []byte) (string, error) {
	done := core.UseSdkConfigContext(c.GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		return c.Ctx().BroadcastTx(tx)
	})
	if err != nil {
		return "", fmt.Errorf("retry broadcastTx err: %s", err)
	}
	res := cc.(*types.TxResponse)
	if res.Code != 0 {
		return res.TxHash, fmt.Errorf("broadcast err with res.code: %d, res.Codespace: %s", res.Code, res.Codespace)
	}
	return res.TxHash, nil
}

func (c *Client) ConstructAndSignTx(msgs ...types.Msg) ([]byte, error) {
	account, err := c.GetAccount()
	if err != nil {
		return nil, err
	}
	done := core.UseSdkConfigContext(c.GetAccountPrefix())
	defer done()

	clientCtx := c.Ctx()
	cmd := cobra.Command{}
	txf, err := clientTx.NewFactoryCLI(clientCtx, cmd.Flags())
	if err != nil {
		return nil, err
	}
	txf = txf.WithSequence(account.GetSequence()).
		WithAccountNumber(account.GetAccountNumber()).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT). // multi sig need this mod
		WithGasAdjustment(1.5).
		WithGas(0).
		WithGasPrices(c.gasPrice).
		WithSimulateAndExecute(true)

	// auto cal gas with retry
	adjusted, err := c.CalculateGas(txf, msgs...)
	if err != nil {
		return nil, err
	}
	txf = txf.WithGas(adjusted)

	txBuilderRaw, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	err = xAuthClient.SignTx(txf, c.Ctx(), clientCtx.GetFromName(), txBuilderRaw, true, true)
	if err != nil {
		return nil, err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilderRaw.GetTx())
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}

func (c *Client) CalculateGas(txf clientTx.Factory, msgs ...types.Msg) (uint64, error) {
	cc, err := c.retry(func() (interface{}, error) {
		_, adjustGas, err := clientTx.CalculateGas(c.Ctx(), txf, msgs...)
		return adjustGas, err
	})
	if err != nil {
		return 0, err
	}

	return cc.(uint64), err
}
