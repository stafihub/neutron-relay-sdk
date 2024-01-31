package client

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/stafihub/neutron-relay-sdk/common/log"

	xWasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
)

var c *Client

func initClient() {
	var err error
	logrus.SetLevel(logrus.TraceLevel)

	kr, err := getKeyring()
	if err != nil {
		logrus.Fatal(err)
	}
	endpoints := []string{"https://rpc-palvus.pion-1.ntrn.tech:443"}
	// endpoints := []string{"http://127.0.0.1:26657"}
	c, err = NewClient(kr, "", "0.005untrn", "neutron", endpoints, log.NewLog("client", "neutron-sdk-test"))
	// c, err = NewClient(kr, "demowallet1", "0.005untrn", "neutron", endpoints, log.NewLog("client", "neutron-relay"))
	if err != nil {
		logrus.Fatal(err)
	}
}

func getKeyring() (keyring.Keyring, error) {
	kr, err := keyring.New("test", keyring.BackendTest, "neutron-testing-data/test-1", os.Stdin, MakeEncodingConfig().Marshaler)
	if err != nil {
		return nil, err
	}

	return kr, nil
}

func TestQueryContract(t *testing.T) {
	initClient()

	res, err := c.queryClient.AllContractState(context.Background(), &xWasmTypes.QueryAllContractStateRequest{
		Address: "neutron1jarq7kgdyd7dcfu2ezeqvg4w4hqdt3m5lv364d8mztnp9pzmwwwqjw7fvg",
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
	res2, err := c.queryClient.SmartContractState(context.Background(), &xWasmTypes.QuerySmartContractStateRequest{
		Address:   "neutron1jarq7kgdyd7dcfu2ezeqvg4w4hqdt3m5lv364d8mztnp9pzmwwwqjw7fvg",
		QueryData: []byte(`{"balance":{"ica_addr":"cosmos15ver270ujn0hy43tr362xnsas5r7pemcm0g9nsyeadlt035eu9nq4u445u"}}`),
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(res2)
}

func TestQueryContractWithHeight(t *testing.T) {
	initClient()

	res, err := c.QuerySmartContractState(
		"neutron1lynzj6ym0y4ne69cm4m3zhmdyuql9clapcdx6k8kd3grxl494g0quy4xjq",
		[]byte(`{"balance":{"address":"neutron1hs0ejaa8g83098zszskhuckvwxypyuhacl0aj2"}}`))
	if err != nil {
		t.Error(err)
	}
	t.Log("latest:", res)

	res, err = c.QuerySmartContractStateWithHeight(
		"neutron1lynzj6ym0y4ne69cm4m3zhmdyuql9clapcdx6k8kd3grxl494g0quy4xjq",
		[]byte(`{"balance":{"address":"neutron1hs0ejaa8g83098zszskhuckvwxypyuhacl0aj2"}}`),
		10595590)
	if err != nil {
		t.Error(err)
	}
	t.Log("10595590:", res)

	res, err = c.QuerySmartContractStateWithHeight(
		"neutron1lynzj6ym0y4ne69cm4m3zhmdyuql9clapcdx6k8kd3grxl494g0quy4xjq",
		[]byte(`{"balance":{"address":"neutron1hs0ejaa8g83098zszskhuckvwxypyuhacl0aj2"}}`),
		10595591)
	if err != nil {
		t.Error(err)
	}
	t.Log("10595591:", res)

	res, err = c.QuerySmartContractStateWithHeight(
		"neutron1lynzj6ym0y4ne69cm4m3zhmdyuql9clapcdx6k8kd3grxl494g0quy4xjq",
		[]byte(`{"balance":{"address":"neutron1lwfzf8h0dkecltvz5wz6jn00t4te9qrhdvcwfa"}}`),
		10595592)
	if err != nil {
		t.Error(err)
	}
	t.Log("10595592:", res)
}

func TestEx(t *testing.T) {
	initClient()

	type Msg struct {
		EraUpdate struct {
			PoolAddr string `json:"pool_addr"`
		} `json:"era_update"`
	}

	bMsg, err := json.Marshal(&Msg{
		EraUpdate: struct {
			PoolAddr string "json:\"pool_addr\""
		}{
			PoolAddr: "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2",
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	msgs := []sdk.Msg{
		&xWasmTypes.MsgExecuteContract{
			Sender:   "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2",
			Contract: "neutron1jarq7kgdyd7dcfu2ezeqvg4w4hqdt3m5lv364d8mztnp9pzmwwwqjw7fvg",
			Msg:      bMsg,
			Funds:    nil,
		},
	}

	txbts, err := c.ConstructAndSignTx(msgs...)
	if err != nil {
		t.Fatal(err)
	}

	txHash, err := c.BroadcastTx(txbts)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(txHash)

	t.Log(txHash)
}
