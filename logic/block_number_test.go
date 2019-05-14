package logic

import (
	"ddex/watcher/connection"
	"github.com/stretchr/testify/suite"
	"testing"
)

type testSuite struct {
	suite.Suite
}

func (suite *testSuite) SetupSuite() {
	connection.InitRedis()
}

func (suite *testSuite) TearDownSuite() {
	connection.CloseRedis()
}

func (suite *testSuite) SetupTest() {
	connection.Redis.FlushDB()
}

func (suite *testSuite) TestGetLastSyncedBlockNumber() {
	number := getLastSyncedBlockNumber()
	suite.Equal(int64(-1), number)
}

func (suite *testSuite) TestSaveSyncedBlockNumber() {
	saveSyncedBlockNumber(100)
	number := getLastSyncedBlockNumber()
	suite.Equal(int64(100), number)
}

func TestSuit(t *testing.T) {
	suite.Run(t, new(testSuite))
}
