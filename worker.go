package appy

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/go-redis/redis/v7"
	"github.com/hibiken/asynq"
)

// Worker processes the background jobs.
type Worker struct {
	*asynq.Background
	*asynq.ServeMux
	asset  *Asset
	config *Config
}

// NewWorker initializes a worker to process background jobs.
func NewWorker(asset *Asset, config *Config, logger *Logger) *Worker {
	workerLogger := &WorkerLogger{
		logger,
		nil,
	}

	workerConfig := &asynq.Config{
		Concurrency:    config.WorkerConcurrency,
		Logger:         workerLogger,
		Queues:         config.WorkerQueues,
		StrictPriority: config.WorkerStrictPriority,
	}

	redisConnOpt := &asynq.RedisClientOpt{
		Addr:     config.WorkerRedisAddr,
		DB:       config.WorkerRedisDB,
		Password: config.WorkerRedisPassword,
		PoolSize: config.WorkerRedisPoolSize,
	}

	if config.WorkerRedisURL != "" {
		redisParseURL, err := redis.ParseURL(config.WorkerRedisURL)
		if err != nil {
			logger.Fatal(err)
		}

		redisConnOpt.Addr = redisParseURL.Addr
		redisConnOpt.DB = redisParseURL.DB
		redisConnOpt.Password = redisParseURL.Password
	}

	worker := &Worker{
		asynq.NewBackground(redisConnOpt, workerConfig),
		asynq.NewServeMux(),
		asset,
		config,
	}

	if len(config.WorkerRedisSentinelAddrs) > 0 {
		redisConnOpt := &asynq.RedisFailoverClientOpt{
			MasterName:       config.WorkerRedisSentinelMasterName,
			SentinelAddrs:    config.WorkerRedisSentinelAddrs,
			SentinelPassword: config.WorkerRedisSentinelPassword,
			DB:               config.WorkerRedisSentinelDB,
			PoolSize:         config.WorkerRedisSentinelPoolSize,
		}

		worker = &Worker{
			asynq.NewBackground(redisConnOpt, workerConfig),
			asynq.NewServeMux(),
			asset,
			config,
		}
	}

	workerLogger.worker = worker
	return worker
}

// Info returns the worker info.
func (w *Worker) Info() []string {
	configPath := w.config.Path()

	if w.asset.moduleRoot != "" && strings.Contains(configPath, w.asset.moduleRoot+"/") {
		configPath = strings.ReplaceAll(configPath, w.asset.moduleRoot+"/", "")
	}

	lines := []string{}
	lines = append(lines,
		fmt.Sprintf("* appy %s (%s), build: %s, environment: %s, config: %s",
			VERSION, runtime.Version(), Build, w.config.AppyEnv, configPath,
		),
	)

	queues := []string{}
	for key, value := range w.config.WorkerQueues {
		queues = append(queues, fmt.Sprintf("%s=%d", key, value))
	}

	lines = append(lines, fmt.Sprintf("* Concurrency: %d, queues: %s", w.config.WorkerConcurrency, strings.Join(queues, ", ")))
	return append(lines, "* Worker is now ready to process jobs...")
}

// Run starts running the worker to process background jobs.
func (w *Worker) Run() {
	w.Background.Run(w.ServeMux)
}
