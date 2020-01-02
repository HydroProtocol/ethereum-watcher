package main

import (
	"context"
	"fmt"
	nights_watch "github.com/HydroProtocol/nights-watch"
	"github.com/HydroProtocol/nights-watch/blockchain"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
)

func main() {
	rootCMD.AddCommand(blockNumCMD)
	rootCMD.AddCommand(usdtTransferCMD)

	if err := rootCMD.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCMD = &cobra.Command{
	Use:   "nights-watch",
	Short: "nights-watch makes getting updates from Ethereum easier",
}

var blockNumCMD = &cobra.Command{
	Use:   "new-block-number",
	Short: "Print number of new block",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
		w := nights_watch.NewHttpBasedEthWatcher(ctx, api)

		logrus.Println("waiting for new block...")
		w.RegisterBlockPlugin(plugin.NewBlockNumPlugin(func(i uint64, b bool) {
			logrus.Printf(">> found new block: %d, is removed: %t", i, b)
		}))

		go func() {
			<-c
			cancel()
		}()

		err := w.RunTillExit()
		if err != nil {
			logrus.Printf("exit with err: %s", err)
		} else {
			logrus.Infoln("exit")
		}
	},
}

var usdtTransferCMD = &cobra.Command{
	Use:   "usdt-transfer",
	Short: "Show Transfer Event of USDT",
	Run: func(cmd *cobra.Command, args []string) {
		api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
		usdtContractAdx := "0xdac17f958d2ee523a2206206994597c13d831ec7"
		// Transfer
		topicsInterestedIn := []string{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}

		handler := func(from, to int, receiptLogs []blockchain.IReceiptLog, isUpToHighestBlock bool) error {

			if from != to {
				logrus.Infof("See new USDT Transfer at blockRange: %d -> %d, count: %2d", from, to, len(receiptLogs))
			} else {
				logrus.Infof("See new USDT Transfer at block: %d, count: %2d", from, len(receiptLogs))
			}

			for _, log := range receiptLogs {
				logrus.Infof("  >> tx: https://etherscan.io/tx/%s", log.GetTransactionHash())
			}

			fmt.Println("  ")

			return nil
		}

		receiptLogWatcher := nights_watch.NewReceiptLogWatcher(
			context.TODO(),
			api,
			-1,
			usdtContractAdx,
			topicsInterestedIn,
			handler,
			nights_watch.ReceiptLogWatcherConfig{
				StepSizeForBigLag:               5,
				IntervalForPollingNewBlockInSec: 5,
				RPCMaxRetry:                     3,
				ReturnForBlockWithNoReceiptLog:  true,
			},
		)

		receiptLogWatcher.Run()
	},
}
