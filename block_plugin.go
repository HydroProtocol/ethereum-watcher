package nights_watch

import "github.com/HydroProtocol/nights-watch/ethrpc"

type IBlockPlugin interface {
	AcceptBlock(ethrpc.Block)
}

type BlockNumPlugin struct {
	callback func(blockNum int, isRemoved bool)
}

func (p BlockNumPlugin) AcceptBlock(b ethrpc.Block) {
	if p.callback != nil {
		p.callback(b.Number, b.IsRemoved)
	}
}

func NewBlockNumPlugin(callback func(int, bool)) BlockNumPlugin {
	return BlockNumPlugin{
		callback: callback,
	}
}

//provide block info with tx basic infos
type SimpleBlockPlugin struct {
	callback func(block ethrpc.Block, isRemoved bool)
}

func (p SimpleBlockPlugin) AcceptBlock(b ethrpc.Block) {
	if p.callback != nil {
		p.callback(b, b.IsRemoved)
	}
}

func NewSimpleBlockPlugin(callback func(block ethrpc.Block, isRemoved bool)) SimpleBlockPlugin {
	return SimpleBlockPlugin{
		callback: callback,
	}
}

//todo, block info with tx full infos
type BlockWithTxInfoPlugin struct {
}
