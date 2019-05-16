package nights_watch

import (
	"context"
	"fmt"
	"github.com/HydroProtocol/nights-watch/ethrpc"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestTxReceiptPlugin(t *testing.T) {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedWatcher(context.Background(), api)

	w.RegisterTxReceiptPlugin(plugin.NewTxReceiptPlugin(func(tx *ethrpc.Transaction, receipt *ethrpc.TransactionReceipt) {
		if tx.IsRemoved {
			fmt.Println("Removed >>", tx.Hash, receipt.TransactionIndex)
		} else {
			fmt.Println("Adding >>", tx.Hash, receipt.TransactionIndex)
		}
	}))

	w.Run()

	time.Sleep(120 * time.Second)
}

func TestErc20TransferPlugin(t *testing.T) {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedWatcher(context.Background(), api)

	w.RegisterTxReceiptPlugin(plugin.NewERC20TransferPlugin(func(token, from, to string, amount decimal.Decimal) {
		fmt.Println("New ERC20 Transfer >>", token, from, "->", to, amount)
	}))

	w.Run()

	time.Sleep(120 * time.Second)
}
