# nights-watch

![](https://github.com/HydroProtocol/nights-watch/workflows/Go/badge.svg)

nights-watch is an event listener for the [Ethereum Blockchain](https://ethereum.org/) written in Golang. With nights-watch you can monitor and track current or historic events that occur on the Ethereum Blockchain.

## Background

Many applications that interact with the Ethereum Blockchain need to know when specific actions occur on the chain, but cannot directly access the on-chain data. nights-watch acts as an interface between application and chain: gathering specified data from the blockchain so that applications can more seamlessly interact with on-chain events.


## Features

1. Plug-in friendly. You can easily add a plugin to nights-watch to listen to any type of on-chain event.
2. Fork Tolerance. If a [fork](https://en.wikipedia.org/wiki/Fork_(blockchain)) occurs, a revert message is sent to the subscriber.

## Example Use Cases

- [DefiWatch](https://defiwatch.io/) monitors the state of Ethereum addresses on several DeFi platforms including DDEX, Compound, DyDx, and Maker. It tracks things like loan ROI, current borrows, liquidations, etc. To track all of this on multiple platforms, DefiWatch has to continuously receive updates from their associated smart contracts. This is done using nights-watch instead of spending time dealing with serialization/deserialization messages from the Ethereum Node, so they can focus on their core logic.   
- Profit & Loss calculations on [DDEX](https://ddex.io). DDEX provides their margin trading users with estimated Profit and Loss (P&L) calculations for their margin positions. To update the P&L as timely and accurately as possible, DDEX uses nights-watch to listen to updates from the Ethereum Blockchain. These updates include: onchain price updates, trading actions from users, and more.
- DDEX also uses an "Eth-Transaction-Watcher" to monitor the on-chain status of trading transactions. DDEX needs to know the latest states of these transactions once they are included in newly mined blocks, so that the platform properly updates trading balances and histories. This is done using the `TxReceiptPlugin` of nights-watch.


# Installation

Run `go get github.com/HydroProtocol/nights-watch`

## Sample Commands

This project is primarily designed as a library to build upon. However, to help others easily understand how nights-watch works and what it is capable of, we prepared some sample commands for you to try out.

**display basic help info**

```shell
docker run hydroprotocolio/nights-watch:master /bin/nights-watch help

nights-watch makes getting updates from Ethereum easier

Usage:
  nights-watch [command]

Available Commands:
  contract-event-listener listen and print events from contract
  help                    Help about any command
  new-block-number        Print number of new block
  usdt-transfer           Show Transfer Event of USDT

Flags:
  -h, --help   help for nights-watch

Use "nights-watch [command] --help" for more information about a command.
```



**print new block numbers**

```shell
docker run hydroprotocolio/nights-watch:master /bin/nights-watch new-block-number

time="2020-01-07T07:33:17Z" level=info msg="waiting for new block..."
time="2020-01-07T07:33:19Z" level=info msg=">> found new block: 9232152, is removed: false"
time="2020-01-07T07:33:44Z" level=info msg=">> found new block: 9232153, is removed: false"
time="2020-01-07T07:34:03Z" level=info msg=">> found new block: 9232154, is removed: false"
time="2020-01-07T07:34:04Z" level=info msg=">> found new block: 9232155, is removed: false"
time="2020-01-07T07:34:05Z" level=info msg=">> found new block: 9232156, is removed: false"
...
```



**see USDT transfer events**

```shell
docker run hydroprotocolio/nights-watch:master /bin/nights-watch usdt-transfer

time="2020-01-07T07:34:32Z" level=info msg="See new USDT Transfer at block: 9232158, count:  9"
time="2020-01-07T07:34:32Z" level=info msg="  >> tx: https://etherscan.io/tx/0x1072efee913229a5e9a3013af6b580099f03ac9c75bfc60013cfa7efac726067"
time="2020-01-07T07:34:32Z" level=info msg="  >> tx: https://etherscan.io/tx/0xae191608df7688ad6d83ebe8151fc5519ff0e29515b557e15ed3fbcbfadca698"
time="2020-01-07T07:34:32Z" level=info msg="  >> tx: https://etherscan.io/tx/0xc3bcfcffa4d3318c27b923eccbd6bbac179f9f982d0282bed164c8e5216130cc"
time="2020-01-07T07:34:32Z" level=info msg="  >> tx: https://etherscan.io/tx/0xe15f2b029c248d1560b10f1879af46b2c8fb2362db9287a8a257042ec2dbb46c"
time="2020-01-07T07:34:32Z" level=info msg="  >> tx: https://etherscan.io/tx/0x860156ebe4c1f00cc479d4f75b7665e2c6ef66d3b085ba1630e803fa44ce2836"
time="2020-01-07T07:34:32Z" level=info msg="  >> tx: https://etherscan.io/tx/0x5b2f4b07332096f55c6ca69e3b349fc19e2c6f54f3a7062f4365697ffb6c5d51"
time="2020-01-07T07:34:32Z" level=info msg="  >> tx: https://etherscan.io/tx/0xb9325afae978b509feee8bc092a090e7a4df750e9e94d10a2bce29e543406c26"
time="2020-01-07T07:34:32Z" level=info msg="  >> tx: https://etherscan.io/tx/0xe379b9e776cf0920a5cc284e9a0f7d141894078827b28f43905dd45457f886e3"
time="2020-01-07T07:34:32Z" level=info msg="  >> tx: https://etherscan.io/tx/0x0f89d867117a6107de44da523b1ad8fbd630d42d9f809c16b82c6f582a8c97da"

time="2020-01-07T07:34:50Z" level=info msg="See new USDT Transfer at block: 9232159, count: 18"
time="2020-01-07T07:34:50Z" level=info msg="  >> tx: https://etherscan.io/tx/0x506ff84b205f0802f037eefe988c9c431d5cae40d3c2c9354616f99f1751bad9"
time="2020-01-07T07:34:50Z" level=info msg="  >> tx: https://etherscan.io/tx/0xd55a1175801eb9af3be65db23d3a9a32f4494f31473ad92dbd95e1d9f9b57fc3"
time="2020-01-07T07:34:50Z" level=info msg="  >> tx: https://etherscan.io/tx/0x93bba744986c75645590b2c930b43e34e9ab219feb38875b87b327854c37e7eb"
time="2020-01-07T07:34:50Z" level=info msg="  >> tx: https://etherscan.io/tx/0xae22eaa9aef7079ff9bb23fa25f314fcf22e9688ef7b585462d3a2afe33628ef"
...
```
**see specific events that occur within a smart contract. The example shows Transfer & Approve events from Multi-Collateral-DAI**

```shell
docker run hydroprotocolio/nights-watch:master /bin/nights-watch contract-event-listener \
    --block-backoff 100 \
    --contract 0x6b175474e89094c44da98b954eedeac495271d0f \
    --events 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
    
INFO[2020-01-07T18:05:26+08:00] --block-backoff activated, we start from block: 9232741 (= 9232841 - 100)

INFO[2020-01-07T18:05:27+08:00] # of interested events at block(9232741->9232745): 1
INFO[2020-01-07T18:05:27+08:00]   >> tx: https://etherscan.io/tx/0x4e05d731ca1b8bf0e2a85d825751312c722bfa3c9210a8b574c06e1e6c992a75

INFO[2020-01-07T18:05:28+08:00] # of interested events at block(9232746->9232750): 3
INFO[2020-01-07T18:05:28+08:00]   >> tx: https://etherscan.io/tx/0x8645784c881a01d326469ddbf45a151f9eab8bedbfc02f6cbe1ab2f03c03c022
INFO[2020-01-07T18:05:28+08:00]   >> tx: https://etherscan.io/tx/0xac5e9fa166a2e7f13aa104812b9278cc65247beff68a4aca7b382e9547d84ec5
INFO[2020-01-07T18:05:28+08:00]   >> tx: https://etherscan.io/tx/0xac5e9fa166a2e7f13aa104812b9278cc65247beff68a4aca7b382e9547d84ec5

INFO[2020-01-07T18:05:29+08:00] # of interested events at block(9232751->9232755): 2
INFO[2020-01-07T18:05:29+08:00]   >> tx: https://etherscan.io/tx/0x280eb9556f6286bf707e859fc6c08bdd1e2ed8f2bdd360883c207f0da0122b01
INFO[2020-01-07T18:05:29+08:00]   >> tx: https://etherscan.io/tx/0xcc5ee10d5ac8f55f51f74b23186052c042ebc5dda61578c1a1038d0b30b6fd91
...
```
Here the flag `--block-backoff` signals for nights-watch to use historic tracking from 100 blocks ago.

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
