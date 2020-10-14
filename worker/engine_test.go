package worker

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	"github.com/hibiken/asynq"
)

type engineSuite struct {
	test.Suite
	asset     *support.Asset
	config    *support.Config
	dbManager *record.Engine
	logger    *support.Logger
	buffer    *bytes.Buffer
	writer    *bufio.Writer
}

func (s *engineSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("DB_URI_PRIMARY", "postgres://postgres:whatever@0.0.0.0:15432/postgres?sslmode=disable&connect_timeout=5")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("WORKER_REDIS_ADDR", "0.0.0.0:16379")

	s.logger, s.buffer, s.writer = support.NewTestLogger()
	s.asset = support.NewAsset(nil, "")
	s.config = support.NewConfig(s.asset, s.logger)
	s.dbManager = record.NewEngine(s.logger, nil)
}

func (s *engineSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("DB_URI_PRIMARY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
	os.Unsetenv("WORKER_REDIS_ADDR")
}

func (s *engineSuite) TestNewEngineWithWorkerRedisAddr() {
	s.config.WorkerRedisAddr = "0.0.0.0:6379"
	s.config.WorkerRedisPassword = "foobar"
	s.config.WorkerRedisDB = 1
	s.config.WorkerRedisPoolSize = 15

	worker := NewEngine(s.asset, s.config, s.dbManager, s.logger)
	s.Equal("0.0.0.0:6379", worker.RedisConnOpt.(*asynq.RedisClientOpt).Addr)
	s.Equal("foobar", worker.RedisConnOpt.(*asynq.RedisClientOpt).Password)
	s.Equal(1, worker.RedisConnOpt.(*asynq.RedisClientOpt).DB)
	s.Equal(15, worker.RedisConnOpt.(*asynq.RedisClientOpt).PoolSize)
}

func (s *engineSuite) TestNewEngineWithWorkerRedisURL() {
	s.config.WorkerRedisURL = "redis://:barfoo@0.0.0.0:26379/3"
	s.config.WorkerRedisPoolSize = 25

	worker := NewEngine(s.asset, s.config, s.dbManager, s.logger)
	s.Equal("0.0.0.0:26379", worker.RedisConnOpt.(*asynq.RedisClientOpt).Addr)
	s.Equal("barfoo", worker.RedisConnOpt.(*asynq.RedisClientOpt).Password)
	s.Equal(3, worker.RedisConnOpt.(*asynq.RedisClientOpt).DB)
	s.Equal(25, worker.RedisConnOpt.(*asynq.RedisClientOpt).PoolSize)
}

func (s *engineSuite) TestNewEngineWithWorkerRedisSentinelAddrs() {
	s.config.WorkerRedisSentinelAddrs = []string{"localhost:6379", "localhost:6380", "localhost:6381"}
	s.config.WorkerRedisSentinelMasterName = "sentinel-master"
	s.config.WorkerRedisSentinelDB = 5
	s.config.WorkerRedisSentinelPassword = "foobar"
	s.config.WorkerRedisSentinelPoolSize = 20

	worker := NewEngine(s.asset, s.config, s.dbManager, s.logger)
	s.Equal([]string{"localhost:6379", "localhost:6380", "localhost:6381"}, worker.RedisConnOpt.(*asynq.RedisFailoverClientOpt).SentinelAddrs)
	s.Equal("foobar", worker.RedisConnOpt.(*asynq.RedisFailoverClientOpt).SentinelPassword)
	s.Equal("sentinel-master", worker.RedisConnOpt.(*asynq.RedisFailoverClientOpt).MasterName)
	s.Equal(5, worker.RedisConnOpt.(*asynq.RedisFailoverClientOpt).DB)
	s.Equal(20, worker.RedisConnOpt.(*asynq.RedisFailoverClientOpt).PoolSize)
}

func (s *engineSuite) TestGlobalMiddleware() {
	s.config.WorkerRedisAddr = "0.0.0.0:6379"
	s.config.WorkerRedisPassword = "foobar"
	s.config.WorkerRedisDB = 1
	s.config.WorkerRedisPoolSize = 15

	count := 0
	worker := NewEngine(s.asset, s.config, s.dbManager, s.logger)
	worker.Use(func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, task *Job) error {
			count += 10
			return next.ProcessTask(ctx, task)
		})
	})
	worker.ProcessTask(context.Background(), NewJob("test", map[string]interface{}{"name": "barfoo"}))

	s.writer.Flush()
	s.Contains(s.buffer.String(), "INFO")
	s.Contains(s.buffer.String(), "[WORKER] job: test, payload: (map[name:barfoo]) start")
	s.Contains(s.buffer.String(), "[WORKER] job: test, payload: (map[name:barfoo]) done in")
	s.Equal(10, count)
}

func (s *engineSuite) TestEnqueue() {
	s.config = support.NewConfig(s.asset, s.logger)
	worker := NewEngine(s.asset, s.config, s.dbManager, s.logger)

	_, err := worker.Enqueue(NewJob("foo", map[string]interface{}{}), nil)
	s.Nil(err)

	_, err = worker.Enqueue(NewJob("bar", map[string]interface{}{}), &JobOptions{ProcessAt: time.Now()})
	s.Nil(err)

	_, err = worker.Enqueue(NewJob("foobar", map[string]interface{}{}), &JobOptions{ProcessIn: 100 * time.Millisecond})
	s.Nil(err)
}

func (s *engineSuite) TestOpsWithTestEnv() {
	os.Setenv("APPY_ENV", "test")
	defer os.Unsetenv("APPY_ENV")

	s.config = support.NewConfig(s.asset, s.logger)
	worker := NewEngine(s.asset, s.config, s.dbManager, s.logger)

	_, err := worker.Enqueue(NewJob("foo", map[string]interface{}{}), nil)
	s.Nil(err)
	s.Equal(len(worker.Jobs()), 1)

	_, err = worker.Enqueue(NewJob("foo", map[string]interface{}{}), &JobOptions{ProcessAt: time.Now()})
	s.Nil(err)
	s.Equal(len(worker.Jobs()), 2)

	worker.Drain()
	s.Equal(len(worker.Jobs()), 0)

	_, err = worker.Enqueue(NewJob("foo", map[string]interface{}{}), &JobOptions{ProcessIn: 5 * time.Minute})
	s.Nil(err)
	s.Equal(len(worker.Jobs()), 1)
}

func (s *engineSuite) TestMockedHandler() {
	ctx := context.Background()
	job := NewJob("test", nil)

	mockedHandler := NewMockedHandler()
	mockedHandler.On("ProcessTask", ctx, job).Return(nil)

	s.Nil(mockedHandler.ProcessTask(ctx, job))
	mockedHandler.AssertExpectations(s.T())
}

func TestEngineSuite(t *testing.T) {
	test.Run(t, new(engineSuite))
}
