package nights_watch

import (
	"context"
	"fmt"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/HydroProtocol/nights-watch/structs"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

// todo why some tx index in block is zero?
func TestTxReceiptPlugin(t *testing.T) {
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

	time.Sleep(120 * time.Second)
}

func TestErc20TransferPlugin(t *testing.T) {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedEthWatcher(context.Background(), api)

	w.RegisterTxReceiptPlugin(plugin.NewERC20TransferPlugin(func(token, from, to string, amount decimal.Decimal, isRemove bool) {
		fmt.Println("New ERC20 Transfer >>", token, from, "->", to, amount, isRemove)
	}))

	w.RunTillExit()

	time.Sleep(120 * time.Second)
}
