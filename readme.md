# Link

https://quip.com/ThpkAJ1V5WbK/Block-Chain-watchers

# MVP

- 支持scaffold-dex的watcher
- 支持ddex的watcher
- 给ddex前端推余额变化

# Interface

```go
watcher.RegisterBlockPlugin(blockPlugin, callback)
interface blockPlugin {
  
}

watcher.RegisterTransactionPlugin(txPlugin, callback)
interface txPlugin {
  
}
```



# Deal with Removed Block & Tx



