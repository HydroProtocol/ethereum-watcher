package nights_watch

import (
	"context"
	"fmt"
	"github.com/HydroProtocol/nights-watch/ethrpc"
	"testing"
	"time"
)

func TestNewBlockNumPlugin(t *testing.T) {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedWatcher(context.Background(), api)

	w.RegisterBlockPlugin(NewBlockNumPlugin(func(i int, b bool) {
		fmt.Println(">>", i, b)
	}))

	w.Run()

	time.Sleep(60 * time.Second)
}

func TestSimpleBlockPlugin(t *testing.T) {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedWatcher(context.Background(), api)

	w.RegisterBlockPlugin(NewSimpleBlockPlugin(func(block ethrpc.Block, isRemoved bool) {
		fmt.Println(">>", block, isRemoved)
	}))

	w.Run()

	time.Sleep(60 * time.Second)
}
