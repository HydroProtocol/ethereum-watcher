package ethereum_watcher

import (
	"context"
	"fmt"
	"github.com/HydroProtocol/ethereum-watcher/blockchain"
	"github.com/HydroProtocol/ethereum-watcher/plugin"
	"github.com/HydroProtocol/ethereum-watcher/structs"
	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"testing"
)

// todo why some tx index in block is zero?
func TestTxReceiptPlugin(t *testing.T) {
	log.SetLevel(log.DEBUG)

	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedEthWatcher(context.Background(), api)

	w.RegisterTxReceiptPlugin(plugin.NewTxReceiptPlugin(func(txAndReceipt *structs.RemovableTxAndReceipt) {
		if txAndReceipt.IsRemoved {
			fmt.Println("Removed >>", txAndReceipt.Tx.GetHash(), txAndReceipt.Receipt.GetTxIndex())
		} else {
			fmt.Println("Adding >>", txAndReceipt.Tx.GetHash(), txAndReceipt.Receipt.GetTxIndex())
		}
	}))

	w.RunTillExit()
}

func TestErc20TransferPlugin(t *testing.T) {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedEthWatcher(context.Background(), api)

	w.RegisterTxReceiptPlugin(plugin.NewERC20TransferPlugin(
		func(token, from, to string, amount decimal.Decimal, isRemove bool) {

			logrus.Infof("New ERC20 Transfer >> token(%s), %s -> %s, amount: %s, isRemoved: %t",
				token, from, to, amount, isRemove)

		},
	))

	w.RunTillExit()
}

func TestFilterPlugin(t *testing.T) {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedEthWatcher(context.Background(), api)

	callback := func(txAndReceipt *structs.RemovableTxAndReceipt) {
		fmt.Println("tx:", txAndReceipt.Tx.GetHash())
	}

	// only accept txs which end with: f
	filterFunc := func(tx blockchain.Transaction) bool {
		txHash := tx.GetHash()

		return txHash[len(txHash)-1:] == "f"
	}

	w.RegisterTxReceiptPlugin(plugin.NewTxReceiptPluginWithFilter(callback, filterFunc))

	err := w.RunTillExitFromBlock(7840000)
	if err != nil {
		fmt.Println("RunTillExit with err:", err)
	}
}

func TestFilterPluginForDyDxApprove(t *testing.T) {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedEthWatcher(context.Background(), api)

	callback := func(txAndReceipt *structs.RemovableTxAndReceipt) {
		receipt := txAndReceipt.Receipt

		for _, log := range receipt.GetLogs() {
			topics := log.GetTopics()
			if len(topics) == 3 &&
				topics[0] == "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925" &&
				topics[2] == "0x0000000000000000000000001e0447b19bb6ecfdae1e4ae1694b0c3659614e4e" {
				fmt.Printf(">> approving to dydx, tx: %s\n", txAndReceipt.Tx.GetHash())
			}
		}
	}

	// only accept txs which send to DAI
	filterFunc := func(tx blockchain.Transaction) bool {
		return tx.GetTo() == "0x89d24a6b4ccb1b6faa2625fe562bdd9a23260359"
	}

	w.RegisterTxReceiptPlugin(plugin.NewTxReceiptPluginWithFilter(callback, filterFunc))

	err := w.RunTillExitFromBlock(7844853)
	if err != nil {
		fmt.Println("RunTillExit with err:", err)
	}
}
