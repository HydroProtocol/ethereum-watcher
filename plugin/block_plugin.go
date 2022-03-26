package plugin

import (
	"github.com/rakshasa/ethereum-watcher/structs"
)

type IBlockPlugin interface {
	AcceptBlock(block *structs.RemovableBlock)
}

type BlockNumPlugin struct {
	callback func(blockNum uint64, isRemoved bool)
}

func (p BlockNumPlugin) AcceptBlock(b *structs.RemovableBlock) {
	if p.callback != nil {
		p.callback(b.Number(), b.IsRemoved)
	}
}

func NewBlockNumPlugin(callback func(uint64, bool)) BlockNumPlugin {
	return BlockNumPlugin{
		callback: callback,
	}
}

//provide block info with tx basic infos
type SimpleBlockPlugin struct {
	callback func(block *structs.RemovableBlock)
}

func (p SimpleBlockPlugin) AcceptBlock(b *structs.RemovableBlock) {
	if p.callback != nil {
		p.callback(b)
	}
}

func NewSimpleBlockPlugin(callback func(block *structs.RemovableBlock)) SimpleBlockPlugin {
	return SimpleBlockPlugin{
		callback: callback,
	}
}

// block info with tx full infos
//type BlockWithTxInfoPlugin struct {
//}
