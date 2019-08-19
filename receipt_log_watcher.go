package nights_watch

import (
	"context"
	"fmt"
	"github.com/HydroProtocol/nights-watch/rpc"
	"github.com/HydroProtocol/nights-watch/structs"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type ReceiptLogWatcher struct {
	ctx                   context.Context
	api                   string
	startBlockNum         int
	contract              string
	interestedTopics      []string
	handler               func(from, to int, receiptLog *structs.RemovableReceiptLog) error
	highestSyncedBlockNum int
	highestSyncedLogIndex int
	config                ReceiptLogWatcherConfig
}

func NewReceiptLogWatcher(
	ctx context.Context,
	api string,
	startBlockNum int,
	contract string,
	interestedTopics []string,
	handler func(from, to int, receiptLog *structs.RemovableReceiptLog) error,
	configs ...ReceiptLogWatcherConfig,
) *ReceiptLogWatcher {

	config := decideConfig(configs...)

	return &ReceiptLogWatcher{
		ctx:                   ctx,
		api:                   api,
		startBlockNum:         startBlockNum,
		contract:              contract,
		interestedTopics:      interestedTopics,
		handler:               handler,
		highestSyncedBlockNum: startBlockNum - 1,
		config:                config,
	}
}

func decideConfig(configs ...ReceiptLogWatcherConfig) ReceiptLogWatcherConfig {
	var config ReceiptLogWatcherConfig
	if len(configs) == 0 {
		config = defaultConfig
	} else {
		config = configs[0]

		if config.IntervalForPollingNewBlockInSec <= 0 {
			config.IntervalForPollingNewBlockInSec = defaultConfig.IntervalForPollingNewBlockInSec
		}

		if config.StepSizeForBigLag <= 0 {
			config.StepSizeForBigLag = defaultConfig.StepSizeForBigLag
		}

		if config.RPCMaxRetry <= 0 {
			config.RPCMaxRetry = defaultConfig.RPCMaxRetry
		}
	}

	return config
}

type ReceiptLogWatcherConfig struct {
	StepSizeForBigLag               int
	ReturnForBlockWithNoReceiptLog  bool
	IntervalForPollingNewBlockInSec int
	RPCMaxRetry                     int
	LagToHighestBlock               int
	StartSyncAfterLogIndex          int
}

var defaultConfig = ReceiptLogWatcherConfig{
	StepSizeForBigLag:               50,
	ReturnForBlockWithNoReceiptLog:  false,
	IntervalForPollingNewBlockInSec: 15,
	RPCMaxRetry:                     5,
	LagToHighestBlock:               0,
	StartSyncAfterLogIndex:          0,
}

func (w *ReceiptLogWatcher) Run() error {

	var blockNumToBeProcessedNext = w.startBlockNum

	rpc := rpc.NewEthRPCWithRetry(w.api, w.config.RPCMaxRetry)

	for {
		select {
		case <-w.ctx.Done():
			return nil
		default:
			highestBlock, err := rpc.GetCurrentBlockNum()
			if err != nil {
				return err
			}

			if blockNumToBeProcessedNext < 0 {
				blockNumToBeProcessedNext = int(highestBlock)
			}

			numOfBlocksToProcess := (int(highestBlock) - w.config.LagToHighestBlock) - blockNumToBeProcessedNext + 1
			if numOfBlocksToProcess <= 0 {
				sleepSec := w.config.IntervalForPollingNewBlockInSec

				logrus.Debugf("no new block after %d, sleep %d seconds", highestBlock, sleepSec)
				time.Sleep(time.Duration(sleepSec) * time.Second)
				continue
			}

			var to int
			if numOfBlocksToProcess > w.config.StepSizeForBigLag {
				// quick mode
				to = blockNumToBeProcessedNext + w.config.StepSizeForBigLag - 1
			} else {
				// normal mode, 1block each time
				to = blockNumToBeProcessedNext
			}

			logs, err := rpc.GetLogs(uint64(blockNumToBeProcessedNext), uint64(to), w.contract, w.interestedTopics)
			if err != nil {
				return err
			}

			if len(logs) == 0 && w.config.ReturnForBlockWithNoReceiptLog {
				err := w.handler(blockNumToBeProcessedNext, to, nil)
				if err != nil {
					logrus.Infof("err when handling nil receipt log, block range: %d - %d", blockNumToBeProcessedNext, to)
					return fmt.Errorf("NIGHTS_WATCH handler(nil) returns error: %s", err)
				}

				w.updateHighestSyncedBlockNumAndLogIndex(to, -1)
			} else {
				for i := 0; i < len(logs); i++ {
					log := logs[i]

					if log.GetBlockNum() == w.startBlockNum && log.GetLogIndex() <= w.highestSyncedLogIndex {
						logrus.Debugf("log(%d-%d) < %d-%d, skipped",
							log.GetBlockNum(), log.GetLogIndex(), w.startBlockNum, w.highestSyncedLogIndex,
						)

						continue
					}

					err := w.handler(blockNumToBeProcessedNext, to, &structs.RemovableReceiptLog{
						IReceiptLog: log,
					})

					if err != nil {
						logrus.Infof("err when handling receipt log, block range: %d - %d, receipt: %+v",
							blockNumToBeProcessedNext, to, log,
						)

						return fmt.Errorf("NIGHTS_WATCH handler returns error: %s", err)
					}

					w.updateHighestSyncedBlockNumAndLogIndex(log.GetBlockNum(), log.GetLogIndex())
				}
			}

			blockNumToBeProcessedNext = to + 1
		}
	}
}

var progressLock = sync.Mutex{}

func (w *ReceiptLogWatcher) updateHighestSyncedBlockNumAndLogIndex(block int, logIndex int) {
	progressLock.Lock()
	defer progressLock.Unlock()

	w.highestSyncedBlockNum = block
	w.highestSyncedLogIndex = logIndex
}

func (w *ReceiptLogWatcher) GetHighestSyncedBlockNum() int {
	return w.highestSyncedBlockNum
}

func (w *ReceiptLogWatcher) GetHighestSyncedBlockNumAndLogIndex() (int, int) {
	progressLock.Lock()
	defer progressLock.Unlock()

	return w.highestSyncedBlockNum, w.highestSyncedLogIndex
}
