package nights_watch

import (
	"container/list"
	"context"
	"fmt"
	"github.com/HydroProtocol/nights-watch/ethrpc"
	"sync"
	"time"
)

type Watcher interface {
	RegisterBlockPlugin(IBlockPlugin)
	RegisterTxPlugin(ITxPlugin)
	Run()
}

type HttpWatcher struct {
	Ctx  context.Context
	lock sync.RWMutex
	API  string
	//httpClient *http.Client
	rpc *ethrpc.RPC

	NewBlockChan chan ethrpc.Block
	//LatestSyncedBlockNum uint64
	SyncedBlocks *list.List

	BlockPlugins []IBlockPlugin
	TxPlugins    []ITxPlugin
}

func NewHttpBasedWatcher(ctx context.Context, api string) *HttpWatcher {
	return &HttpWatcher{
		Ctx:          ctx,
		API:          api,
		rpc:          ethrpc.New(api),
		NewBlockChan: make(chan ethrpc.Block, 32),
		SyncedBlocks: list.New(),
	}
}

func (watcher *HttpWatcher) RegisterBlockPlugin(plugin IBlockPlugin) {
	watcher.BlockPlugins = append(watcher.BlockPlugins, plugin)
}

func (watcher *HttpWatcher) RegisterTxPlugin(plugin ITxPlugin) {
	watcher.TxPlugins = append(watcher.TxPlugins, plugin)
}

func (watcher *HttpWatcher) LatestSyncedBlockNum() int {
	watcher.lock.RLock()
	defer watcher.lock.RUnlock()

	if watcher.SyncedBlocks.Len() <= 0 {
		return -1
	}

	b := watcher.SyncedBlocks.Back().Value.(*ethrpc.Block)

	return (*b).Number
}

func (watcher *HttpWatcher) Run() {
	go func() {
		for {
			select {
			case <-watcher.Ctx.Done():
				return
			default:
				latestBlockNum, err := watcher.getCurrentBlockNum()
				fmt.Println("latestBlockNum", latestBlockNum)
				if err != nil {
					panic(err)
				}

				noNewBlockForSync := watcher.LatestSyncedBlockNum() >= latestBlockNum

				fmt.Println("watcher.LatestSyncedBlockNum()", watcher.LatestSyncedBlockNum())
				for watcher.LatestSyncedBlockNum() < latestBlockNum {
					var newBlockNumToSync int
					if watcher.LatestSyncedBlockNum() <= 0 {
						newBlockNumToSync = latestBlockNum
					} else {
						newBlockNumToSync = watcher.LatestSyncedBlockNum() + 1
					}

					newBlock, err := watcher.getBlockByNum(newBlockNumToSync)
					if err != nil {
						panic(err)
					}

					if watcher.FoundFork(newBlock) {
						fmt.Println("found fork, popping")
						watcher.popBlocksUntilReachMainChain()
					} else {
						fmt.Println("adding new block")
						watcher.addNewBlock(*newBlock)
					}
				}

				if noNewBlockForSync {
					fmt.Println("no new block to sync, sleep for 3 secs")

					// sleep for 3 secs
					timer := time.NewTimer(3 * time.Second)
					<-timer.C
				}
			}
		}
	}()

	go func() {
		for {
			block := <-watcher.NewBlockChan
			// run thru block plugins
			for i := 0; i < len(watcher.BlockPlugins); i++ {
				blockPlugin := watcher.BlockPlugins[i]

				blockPlugin.AcceptBlock(block)
			}

			// run thru tx plugins
			txPlugins := watcher.TxPlugins
			for i := 0; i < len(txPlugins); i++ {
				txPlugin := txPlugins[i]

				for j := 0; j < len(block.Transactions); j++ {
					txPlugin.AcceptTx(block.Transactions[j])
				}
			}
		}
	}()
}

func (watcher *HttpWatcher) addNewBlock(block ethrpc.Block) {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()

	watcher.SyncedBlocks.PushBack(block)
	watcher.NewBlockChan <- block
}

func (watcher *HttpWatcher) popBlocksUntilReachMainChain() {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()

	for {
		if watcher.SyncedBlocks.Back() == nil {
			return
		}

		block, err := watcher.getBlockByNum(watcher.LatestSyncedBlockNum())
		if err != nil {
			continue
		}

		lastSyncedBlock := watcher.SyncedBlocks.Back().Value.(*ethrpc.Block)

		if block.Hash != (*lastSyncedBlock).Hash {
			removedBlock := watcher.SyncedBlocks.Remove(watcher.SyncedBlocks.Back()).(*ethrpc.Block)
			removedBlock.IsRemoved = true

			watcher.NewBlockChan <- *removedBlock
		} else {
			return
		}
	}
}

func (watcher *HttpWatcher) FoundFork(newBlock *ethrpc.Block) bool {
	for e := watcher.SyncedBlocks.Back(); e != nil; e = e.Prev() {
		syncedBlock := e.Value.(*ethrpc.Block)

		if (*syncedBlock).Number+1 == newBlock.Number {
			return (*syncedBlock).Hash != newBlock.ParentHash
		}
	}

	return false
}

func (watcher *HttpWatcher) getBlockByNum(blockNum int) (*ethrpc.Block, error) {
	return watcher.rpc.EthGetBlockByNumber(int(blockNum), true)
}

func (watcher *HttpWatcher) getCurrentBlockNum() (int, error) {
	i, err := watcher.rpc.EthBlockNumber()
	if err != nil {
		return -1, err
	} else {
		return i, nil
	}
}

// types and interfaces
type Hash [32]byte

type BlockWithTxInfos struct {
	ethrpc.Block
	Transactions []Transaction
}

type Transaction interface {
	Hash() Hash
}

type TransactionReceipt interface {
	Hash() Hash
	Status() bool
}
