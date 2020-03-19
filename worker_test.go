package appy

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

type WorkerSuite struct {
	TestSuite
	asset  *Asset
	config *Config
	logger *Logger
	buffer *bytes.Buffer
	writer *bufio.Writer
}

func (s *WorkerSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, s.buffer, s.writer = NewFakeLogger()
	s.asset = NewAsset(nil, map[string]string{
		"docker": "testdata/server/.docker",
		"config": "testdata/server/configs",
		"locale": "testdata/server/pkg/locales",
		"view":   "testdata/server/pkg/views",
		"web":    "testdata/server/web",
	}, "")
	s.config = NewConfig(s.asset, s.logger, &Support{})
}

func (s *WorkerSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *WorkerSuite) TestNewWorkerWithWorkerRedisAddr() {
	s.config.WorkerRedisAddr = "0.0.0.0:6379"
	s.config.WorkerRedisPassword = "foobar"
	s.config.WorkerRedisDB = 1
	s.config.WorkerRedisPoolSize = 15

	worker := NewWorker(s.asset, s.config, s.logger)
	s.Equal("0.0.0.0:6379", worker.RedisConnOpt.(*asynq.RedisClientOpt).Addr)
	s.Equal("foobar", worker.RedisConnOpt.(*asynq.RedisClientOpt).Password)
	s.Equal(1, worker.RedisConnOpt.(*asynq.RedisClientOpt).DB)
	s.Equal(15, worker.RedisConnOpt.(*asynq.RedisClientOpt).PoolSize)
}

func (s *WorkerSuite) TestNewWorkerWithWorkerRedisURL() {
	s.config.WorkerRedisURL = "redis://:barfoo@0.0.0.0:26379/3"
	s.config.WorkerRedisPoolSize = 25

	worker := NewWorker(s.asset, s.config, s.logger)
	s.Equal("0.0.0.0:26379", worker.RedisConnOpt.(*asynq.RedisClientOpt).Addr)
	s.Equal("barfoo", worker.RedisConnOpt.(*asynq.RedisClientOpt).Password)
	s.Equal(3, worker.RedisConnOpt.(*asynq.RedisClientOpt).DB)
	s.Equal(25, worker.RedisConnOpt.(*asynq.RedisClientOpt).PoolSize)
}

func (s *WorkerSuite) TestNewWorkerWithWorkerRedisSentinelAddrs() {
	s.config.WorkerRedisSentinelAddrs = []string{"localhost:6379", "localhost:6380", "localhost:6381"}
	s.config.WorkerRedisSentinelMasterName = "sentinel-master"
	s.config.WorkerRedisSentinelDB = 5
	s.config.WorkerRedisSentinelPassword = "foobar"
	s.config.WorkerRedisSentinelPoolSize = 20

	worker := NewWorker(s.asset, s.config, s.logger)
	s.Equal([]string{"localhost:6379", "localhost:6380", "localhost:6381"}, worker.RedisConnOpt.(*asynq.RedisFailoverClientOpt).SentinelAddrs)
	s.Equal("foobar", worker.RedisConnOpt.(*asynq.RedisFailoverClientOpt).SentinelPassword)
	s.Equal("sentinel-master", worker.RedisConnOpt.(*asynq.RedisFailoverClientOpt).MasterName)
	s.Equal(5, worker.RedisConnOpt.(*asynq.RedisFailoverClientOpt).DB)
	s.Equal(20, worker.RedisConnOpt.(*asynq.RedisFailoverClientOpt).PoolSize)
}

func (s *WorkerSuite) TestWorkerGlobalMiddleware() {
	s.config.WorkerRedisAddr = "0.0.0.0:6379"
	s.config.WorkerRedisPassword = "foobar"
	s.config.WorkerRedisDB = 1
	s.config.WorkerRedisPoolSize = 15

	worker := NewWorker(s.asset, s.config, s.logger)
	worker.ProcessTask(context.Background(), NewJob("test", map[string]interface{}{"name": "barfoo"}))
	s.writer.Flush()
	s.Contains(s.buffer.String(), "INFO")
	s.Contains(s.buffer.String(), "[WORKER] job: test, payload: (name:barfoo) start")
	s.Contains(s.buffer.String(), "[WORKER] job: test, payload: (name:barfoo) done in")
}

func (s *WorkerSuite) TestWorkerTestUtils() {
	os.Setenv("APPY_ENV", "test")
	defer os.Unsetenv("APPY_ENV")

	s.config = NewConfig(s.asset, s.logger, &Support{})
	worker := NewWorker(s.asset, s.config, s.logger)

	err := worker.Enqueue(NewJob("foo", map[string]interface{}{}), nil)
	s.Nil(err)
	s.Equal(len(worker.Jobs()), 1)

	err = worker.EnqueueAt(time.Now(), NewJob("foo", map[string]interface{}{}), nil)
	s.Nil(err)
	s.Equal(len(worker.Jobs()), 2)

	worker.Drain()
	s.Equal(len(worker.Jobs()), 0)

	err = worker.EnqueueIn(5*time.Minute, NewJob("foo", map[string]interface{}{}), nil)
	s.Nil(err)
	s.Equal(len(worker.Jobs()), 1)
}

func TestWorkerSuite(t *testing.T) {
	RunTestSuite(t, new(WorkerSuite))
}
