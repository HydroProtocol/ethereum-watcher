package main

import (
	"context"
	"fmt"
	nights_watch "github.com/HydroProtocol/nights-watch"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
)

func main() {
	rootCmd.AddCommand(blockNumCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "nights-watch",
	Short: "nights-watch makes getting updates from Ethereum easier",
}

var blockNumCmd = &cobra.Command{
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
