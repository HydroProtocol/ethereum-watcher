package nights_watch

import (
	"context"
	"fmt"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/HydroProtocol/nights-watch/structs"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestReceiptLogsPlugin(t *testing.T) {
	logrus.SetLevel(logrus.InfoLevel)

	api := "https://kovan.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedEthWatcher(context.Background(), api)

	contract := "0x63bB8a255a8c045122EFf28B3093Cc225B711F6D"
	// Match
	topics := []string{"0x6bf96fcc2cec9e08b082506ebbc10114578a497ff1ea436628ba8996b750677c"}

	w.RegisterReceiptLogPlugin(plugin.NewReceiptLogPlugin(contract, topics, func(receipt *structs.RemovableReceiptLog) {
		if receipt.IsRemoved {
			logrus.Infof("Removed >> %+v", receipt)
		} else {
			logrus.Infof("Adding >> %+v, tx: %s, logIdx: %d", receipt, receipt.IReceiptLog.GetTransactionHash(), receipt.IReceiptLog.GetLogIndex())
		}
	}))

	startBlock := 12101723
	err := w.RunTillExitFromBlock(uint64(startBlock))

	fmt.Println("err:", err)
}
