# nights-watch

![](https://github.com/HydroProtocol/nights-watch/workflows/Go/badge.svg)

nights-watch is an event listener for the [Ethereum Blockchain](https://ethereum.org/) written in Golang. With nights-watch you can monitor and track current or historic events that occur on the Ethereum Blockchain.

## Background

Many applications that interact with the Ethereum Blockchain need to know when specific actions occur on the chain, but cannot directly access the on-chain data. nights-watch acts as an interface between application and chain: gathering specified data from the blockchain so that applications can more seamlessly interact with on-chain events.


## Features

1. Plug-in friendly. You can easily add a plugin to nights-watch to listen to any type of on-chain event.
2. Fork Tolerance. If a [fork](https://en.wikipedia.org/wiki/Fork_(blockchain)) occurs, a revert message is sent to the subscriber.

# Installation

Run `go get github.com/HydroProtocol/nights-watch`

## Sample Commands

This project is primarily designed as a library to build upon. However, to help others easily understand how nights-watch works and what it is capable of, we prepared some sample commands for you to try out.

**display basic help info**

```shell
docker run diveinto/nights-watch:master /bin/nights-watch help
```
![Screen Shot 2020-01-03 at 10 31 24 AM](https://user-images.githubusercontent.com/698482/71704263-40d2c580-2e14-11ea-87be-1e3bcfa775a2.png)


**print new block numbers**

```shell
docker run diveinto/nights-watch:master /bin/nights-watch new-block-number
```
![Screen Shot 2020-01-03 at 10 38 58 AM](https://user-images.githubusercontent.com/698482/71704417-44b31780-2e15-11ea-9ff1-178c039cadeb.png)


**see USDT transfer events**

```shell
docker run diveinto/nights-watch:master /bin/nights-watch usdt-transfer
```
![](http://wx4.sinaimg.cn/large/6272aa65ly1gaj59yks3vj214g0antco.jpg)

**see specific events that occur within a smart contract. The example shows Transfer & Approve events from Multi-Collateral-DAI**

```shell
docker run diveinto/nights-watch:master /bin/nights-watch contract-event-listener \
    --contract 0x6b175474e89094c44da98b954eedeac495271d0f \
    --events 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
```
![Screen Shot 2020-01-03 at 10 36 20 AM](https://user-images.githubusercontent.com/698482/71704362-e6863480-2e14-11ea-8daa-78f9b1bfa243.png)

# Usage

To effectively use nights-watch, you will be interacting with two primary structs:

- Watcher
- ReceiptLogWatcher

## Watcher

`Watcher` is an HTTP client which continuously polls newly mined blocks on the Ethereum Blockchain. We can incorporate various kinds of "plugins" into `Watcher`, which will poll for specific types of events and data, such as:

- BlockPlugin
- TransactionPlugin
- TransactionReceiptPlugin
- ReceiptLogPlugin

Once the `Watcher` sees a new block, it will parse the info and feed the data into the registered plugins. You can see some of the code examples [below](#watcher-examples).

The plugins are designed to be easily modifiable so that you can create your own plugin based on the provided ones. For example, the code shown [below](#listen-for-new-erc20-transfer-events) registers an `ERC20TransferPlugin` to show new ERC20 Transfer Events. This plugin simply parses some receipt info from a different `TransactionReceiptPlugin`. So if you want to show more info than what the `ERC20TransferPlugin` shows, like the gas used in the transaction, you can easily create a `BetterERC20TransferPlugin` showing that.

### Watcher Examples

#### Print number of newly mined blocks

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

#### Listen for new ERC20 Transfer Events

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

`Watcher` is polling for blocks one by one, so what if we want to query certain events from the latest 10000 blocks? `Watcher` can do that but fetching blocks one by one can be slow. `ReceiptLogWatcher` to the rescue!

`ReceiptLogWatcher` makes use of the `eth_getLogs` to query for logs in a batch. Check out the code [below](#example-of-receiptlogwatcher) to see how to use it.


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
  
	// ERC20 Transfer Event
	topicsInterestedIn := []string{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}

	handler := func(from, to int, receiptLogs []blockchain.IReceiptLog, isUpToHighestBlock bool) error {
		logrus.Infof("USDT Transfer count: %d, %d -> %d", len(receiptLogs), from, to)
		return nil
	}

	// query for USDT Transfer Events
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
