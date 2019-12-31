package nights_watch

import (
	"context"
	"github.com/HydroProtocol/nights-watch/blockchain"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestReceiptLogWatcher_Run(t *testing.T) {
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
