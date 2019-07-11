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
	// Deposit
	topics := []string{"0x137e5dc9e374ee2e2ceb51a34d187b03b8a1d348da6f35c8519c4f97b720df5f"}

	w.RegisterReceiptLogPlugin(plugin.NewReceiptLogPlugin(contract, topics, func(receipt *structs.RemovableReceiptLog) {
		if receipt.IsRemoved {
			logrus.Infof("Removed >> %+v", receipt)
		} else {
			logrus.Infof("Adding >> %+v", receipt)
		}
	}))

	err := w.RunTillExitFromBlock(12147113)

	fmt.Println("err:", err)
}
