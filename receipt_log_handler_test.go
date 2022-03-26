package ethereum_watcher

import (
	"context"
	"github.com/rakshasa/ethereum-watcher/rpc"
	"github.com/rakshasa/ethereum-watcher/structs"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestListenForReceiptLogTillExit(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	ctx := context.Background()
	api := "https://kovan.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	startBlock := 12220000
	contract := "0xAc34923B2b8De9d441570376e6c811D5aA5ed72f"
	interestedTopics := []string{
		"0x23b872dd7302113369cda2901243429419bec145408fa8b352b3dd92b66c680b",
		"0x6bf96fcc2cec9e08b082506ebbc10114578a497ff1ea436628ba8996b750677c",
		"0x5a746ce5ce37fc996a6e682f4f84b6f90d3be79fd8ac9a8a11264345f3d29edd",
		"0x9c4e90320be51bb93d854d0ab9ba8aa249dabc21192529efcd76ae7c22c6bc0b",
		"0x0ce31a5f70780bb6770b52a6793403d856441ccb475715e8382a0525d35b0558",
	}

	handler := func(log structs.RemovableReceiptLog) {
		logrus.Infof("log from tx: %s", log.GetTransactionHash())
	}

	stepsForBigLag := 100

	highestProcessed := ListenForReceiptLogTillExit(ctx, api, startBlock, contract, interestedTopics, handler, stepsForBigLag)
	logrus.Infof("highestProcessed: %d", highestProcessed)
}

func TestListenForReceiptLogTillExit2(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	stepsForBigLag := 100

	ctx := context.Background()
	api := "https://kovan.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"

	startBlock, _ := rpc.NewEthRPC(api).GetCurrentBlockNum()
	startBlock = startBlock - uint64(stepsForBigLag*10)

	contract := "0xAc34923B2b8De9d441570376e6c811D5aA5ed72f"
	interestedTopics := []string{
		"0x23b872dd7302113369cda2901243429419bec145408fa8b352b3dd92b66c680b",
	}

	handler := func(log structs.RemovableReceiptLog) {
		logrus.Infof("log from tx: %s", log.GetTransactionHash())
	}

	highestProcessed := ListenForReceiptLogTillExit(ctx, api, int(startBlock), contract, interestedTopics, handler, stepsForBigLag)
	logrus.Infof("highestProcessed: %d", highestProcessed)
}
