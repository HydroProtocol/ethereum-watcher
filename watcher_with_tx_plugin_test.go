package nights_watch

import (
	"context"
	"fmt"
	"github.com/HydroProtocol/nights-watch/plugin"
	"testing"
	"time"
)

func TestTxPlugin(t *testing.T) {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedWatcher(context.Background(), api)

	w.RegisterTxPlugin(plugin.NewTxHashPlugin(func(txHash string, isRemoved bool) {
		fmt.Println(">>", txHash, isRemoved)
	}))

	w.Run()

	time.Sleep(60 * time.Second)
}
