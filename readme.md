# nights-watch

![](https://github.com/HydroProtocol/nights-watch/workflows/Go/badge.svg)

nights-watch is an event listener for Ethereum written in Golang.

# Why build this?

Many backend services at Hydro need to know things are changed on the blockchain. They may be some events or transaction results. 

# Features

1. Plug-in friendly. It easy to add a plugin to listen on a sort of events.
2. Fork Tolerance. If a fork occurs, a revert message will bet sent to the subscriber.

# Install

Run `go get github.com/HydroProtocol/nights-watch`

# How to use

the most two important structs we provide are:

- Watcher
- ReceiptLogWatcher

## Watcher

`Watcher` is an HTTP client keeps polling for newly mined blocks on Ethereum, we can register various kinds of plugins into `Watcher`, including:

- BlockPlugin
- TransactionPlugin
- TransactionReceiptPlugin
- ReceiptLogPlugin

Once the `Watcher` sees a new block, it will parse the info and feed the data into the registered plugins, you can see the code examples [below](#example-of-watcher).

One Interesting thing about plugins is that you can create your plugin based on provided ones, for example, as the code is shown [below](#listen-for-new-erc20-transfer-events), we register an `ERC20TransferPlugin` to show new ERC20 Transfer Events, this plugin is based on `TransactionReceiptPlugin`, what we do in it is simply parse the receipts info from `TransactionReceiptPlugin`. So if you want to show more info than `ERC20TransferPlugin`, for example, the gas used for this transfer transaction, you can easily create a `BetterERC20TransferPlugin` showing that.

### Example of Watcher

**Print number of newly mined blocks**

```go
package main

import (
	"context"
	"fmt"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/HydroProtocol/nights-watch/structs"
)

func main() {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedEthWatcher(context.Background(), api)

  // we use BlockPlugin here
	w.RegisterBlockPlugin(plugin.NewBlockNumPlugin(func(i uint64, b bool) {
		fmt.Println(">>", i, b)
	}))

	w.RunTillExit()
}
```

**Listen for new ERC20 Transfer Events** 

```go
package main

import (
	"context"
	"fmt"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/HydroProtocol/nights-watch/structs"
	"github.com/sirupsen/logrus"
)

func main() {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedEthWatcher(context.Background(), api)

  // we use TxReceiptPlugin here
	w.RegisterTxReceiptPlugin(plugin.NewERC20TransferPlugin(
		func(token, from, to string, amount decimal.Decimal, isRemove bool) {

			logrus.Infof("New ERC20 Transfer >> token(%s), %s -> %s, amount: %s, isRemoved: %t",
				token, from, to, amount, isRemove)

		},
	))

	w.RunTillExit()
}
```

## ReceiptLogWatcher

`Watcher` is polling for blocks one by one, so what if we want to check certain events of the last 10000 blocks? `Watcher` can do that but fetching blocks one by one can be slow. `ReceiptLogWatcher` comes to the rescue.

`ReceiptLogWatcher` make use of the `eth_getLogs` to query for logs in batch. check out the code [below](#example-of-receiptlogwatcher) to see how to use it.


### Example of ReceiptLogWatcher

```go
package main

import (
	"context"
	"github.com/HydroProtocol/nights-watch/blockchain"
	"github.com/sirupsen/logrus"
)

func main() {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	usdtContractAdx := "0xdac17f958d2ee523a2206206994597c13d831ec7"
	// Transfer
	topicsInterestedIn := []string{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}

	handler := func(from, to int, receiptLogs []blockchain.IReceiptLog, isUpToHighestBlock bool) error {
		logrus.Infof("USDT Transfer count: %d, %d -> %d", len(receiptLogs), from, to)
		return nil
	}

	receiptLogWatcher := NewReceiptLogWatcher(
		context.TODO(),
		api,
		-1,
		usdtContractAdx,
		topicsInterestedIn,
		handler,
		ReceiptLogWatcherConfig{
			StepSizeForBigLag:               5,
			IntervalForPollingNewBlockInSec: 5,
			RPCMaxRetry:                     3,
			ReturnForBlockWithNoReceiptLog:  true,
		},
	)

	receiptLogWatcher.Run()
}
```



# License

[Apache 2.0 License](LICENSE)