package nights_watch

import (
	"context"
	"github.com/HydroProtocol/ethereum-watcher/rpc"
	"github.com/HydroProtocol/ethereum-watcher/structs"
	"github.com/sirupsen/logrus"
	"time"
)

const DefaultStepSizeForBigLag = 10

//deprecated, please use receipt_log_watcher instead.
func ListenForReceiptLogTillExit(
	ctx context.Context,
	api string,
	startBlock int,
	contract string,
	interestedTopics []string,
	handler func(receiptLog structs.RemovableReceiptLog),
	steps ...int,
) int {
	var stepSizeForBigLag int
	if len(steps) > 0 && steps[0] > 0 {
		stepSizeForBigLag = steps[0]
	} else {
		stepSizeForBigLag = DefaultStepSizeForBigLag
	}

	rpc := rpc.NewEthRPCWithRetry(api, 5)

	var blockNumToBeProcessedNext = startBlock

	for {
		select {
		case <-ctx.Done():
			return blockNumToBeProcessedNext - 1
		default:
			highestBlock, err := rpc.GetCurrentBlockNum()
			if err != nil {
				return blockNumToBeProcessedNext - 1
			}

			if blockNumToBeProcessedNext < 0 {
				blockNumToBeProcessedNext = int(highestBlock)
			}

			numOfBlocksToProcess := int(highestBlock) - blockNumToBeProcessedNext + 1
			if numOfBlocksToProcess <= 0 {
				logrus.Debugf("no ready block after %d, sleep 3 seconds", highestBlock)
				time.Sleep(3 * time.Second)
				continue
			}

			var to int
			if numOfBlocksToProcess > stepSizeForBigLag {
				// quick mode
				to = blockNumToBeProcessedNext + stepSizeForBigLag - 1
			} else {
				// normal mode, 1block each time
				to = blockNumToBeProcessedNext
			}

			logs, err := rpc.GetLogs(uint64(blockNumToBeProcessedNext), uint64(to), contract, interestedTopics)
			if err != nil {
				return blockNumToBeProcessedNext - 1
			}

			for i := 0; i < len(logs); i++ {
				handler(structs.RemovableReceiptLog{
					IReceiptLog: logs[i],
				})
			}

			blockNumToBeProcessedNext = to + 1
		}
	}
}
