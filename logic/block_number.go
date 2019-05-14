package logic

import (
	"ddex/watcher/connection"
	"github.com/go-redis/redis"
	"strconv"
)

func getLastSyncedBlockNumber() int64 {
	val, err := connection.Redis.Get("WATCHER_LAST_SYNCED_BLOCK_NUMBER").Result()

	if err == redis.Nil {
		return -1
	} else if err != nil {
		panic(err)
	}

	intVal, err := strconv.ParseInt(val, 10, 64)

	if err != nil {
		panic(err)
	}

	return intVal
}

func saveSyncedBlockNumber(number int64) {
	err := connection.Redis.Set("WATCHER_LAST_SYNCED_BLOCK_NUMBER", number, 0).Err()

	if err != nil {
		panic(err)
	}
}
