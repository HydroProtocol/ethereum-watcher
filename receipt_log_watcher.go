package nights_watch

import (
	"context"
	"github.com/HydroProtocol/nights-watch/rpc"
	"github.com/HydroProtocol/nights-watch/structs"
	"github.com/sirupsen/logrus"
	"time"
)

type ReceiptLogWatcher struct {
	ctx                   context.Context
	api                   string
	startBlockNum         int
	contract              string
	interestedTopics      []string
	handler               func(receiptLog structs.RemovableReceiptLog)
	steps                 int
	highestSyncedBlockNum int
}

func NewReceiptLogWatcher(
	ctx context.Context,
	api string,
	startBlockNum int,
	contract string,
	interestedTopics []string,
	handler func(receiptLog structs.RemovableReceiptLog),
	steps int,
) *ReceiptLogWatcher {
	return &ReceiptLogWatcher{
		ctx:                   ctx,
		api:                   api,
		startBlockNum:         startBlockNum,
		contract:              contract,
		interestedTopics:      interestedTopics,
		handler:               handler,
		steps:                 steps,
		highestSyncedBlockNum: startBlockNum - 1,
	}
}

func (w *ReceiptLogWatcher) Run() error {

	var blockNumToBeProcessedNext = w.startBlockNum
	defer func() {
		w.updateHighestSyncedBlockNum(blockNumToBeProcessedNext - 1)
	}()

	rpc := rpc.NewEthRPCWithRetry(w.api, 5)

	for {
		select {
		case <-w.ctx.Done():
			//w.updateHighestSyncedBlockNum(blockNumToBeProcessedNext - 1)
			return nil
		default:
			highestBlock, err := rpc.GetCurrentBlockNum()
			if err != nil {
				//w.updateHighestSyncedBlockNum(blockNumToBeProcessedNext - 1)
				return err
			}

			if blockNumToBeProcessedNext < 0 {
				blockNumToBeProcessedNext = int(highestBlock)
			}

			numOfBlocksToProcess := int(highestBlock) - blockNumToBeProcessedNext + 1
			if numOfBlocksToProcess <= 0 {
				logrus.Debugf("no new block after %d, sleep 3 seconds", highestBlock)
				time.Sleep(3 * time.Second)
				continue
			}

			var to int
			if numOfBlocksToProcess > w.steps {
				// quick mode
				to = blockNumToBeProcessedNext + w.steps - 1
			} else {
				// normal mode, 1block each time
				to = blockNumToBeProcessedNext
			}

			logs, err := rpc.GetLogs(uint64(blockNumToBeProcessedNext), uint64(to), w.contract, w.interestedTopics)
			if err != nil {
				//w.updateHighestSyncedBlockNum(blockNumToBeProcessedNext - 1)
				return err
			}

			for i := 0; i < len(logs); i++ {
				w.handler(structs.RemovableReceiptLog{
					IReceiptLog: logs[i],
				})
			}

			w.updateHighestSyncedBlockNum(to)
			blockNumToBeProcessedNext = to + 1
		}
	}
}

func (w *ReceiptLogWatcher) updateHighestSyncedBlockNum(num int) {
	w.highestSyncedBlockNum = num
}

func (w *ReceiptLogWatcher) GetHighestSyncedBlockNum() int {
	return w.highestSyncedBlockNum
}
