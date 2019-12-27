# nights-watch

![](https://github.com/HydroProtocol/nights-watch/workflows/Go/badge.svg)

nights-watch is an event listener for Ethereum written in Golang.

## Why we build this?

Many backend services at Hydro need to know things are changed on blockchain. They may be some events or transactions results. 

## Features

1. Plug-in friendly. It easy to add a plugin to listen on a sort of events.
2. Fork Tolerance. If a fork occurs, a revert message will bet sent to the subscriber.

## Install

Run `go get github.com/HydroProtocol/nights-watch`

## Example

```golang
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

	w.RegisterBlockPlugin(plugin.NewBlockNumPlugin(func(i uint64, b bool) {
		fmt.Println(">>", i, b)
	}))

	w.RunTillExit()
}
```








