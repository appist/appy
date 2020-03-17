package appy

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/hibiken/asynq"
)

var rawWorkerDataRegex = regexp.MustCompile("(?i)({data:map[|]})")

// Worker processes the background jobs.
type Worker struct {
	*asynq.Background
	*asynq.Client
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

	mux := asynq.NewServeMux()
	mux.Use(func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			payload := strings.ReplaceAll(fmt.Sprintf("%+v", t.Payload), "{data:map[", "")
			payload = strings.ReplaceAll(fmt.Sprintf("%+v", payload), "]}", "")
			logger.Infof(`[WORKER] job: %s, payload: (%s) start`, t.Type, payload)

			start := time.Now()
			err := h.ProcessTask(ctx, t)
			logger.Infof(`[WORKER] job: %s, payload: (%s) done in %s`, t.Type, payload, time.Since(start))

			return err
		})
	})

	worker := &Worker{
		asynq.NewBackground(redisConnOpt, workerConfig),
		asynq.NewClient(redisConnOpt),
		mux,
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
			asynq.NewClient(redisConnOpt),
			mux,
			asset,
			config,
		}
	}

	workerLogger.worker = worker
	return worker
}

// Enqueue enqueues job to be processed immediately.
//
// Enqueue returns nil if the job is enqueued successfully, otherwise returns an error.
func (w *Worker) Enqueue(job *Job, opts *JobOptions) error {
	return w.Client.Enqueue(job.Task, parseJobOptions(opts))
}

// EnqueueAt schedules job to be enqueued at the specified time.
//
// It returns nil if the job is scheduled successfully, otherwise returns an error.
func (w *Worker) EnqueueAt(t time.Time, job *Job, opts *JobOptions) error {
	return w.Client.EnqueueAt(t, job.Task, parseJobOptions(opts))
}

// EnqueueIn schedules job to be enqueued after the specified delay.
//
// It returns nil if the job is scheduled successfully, otherwise returns an error.
func (w *Worker) EnqueueIn(d time.Duration, job *Job, opts *JobOptions) error {
	return w.Client.EnqueueIn(d, job.Task, parseJobOptions(opts))
}

// HandleFunc registers the handler function for the given pattern.
func (w *Worker) HandleFunc(pattern string, handler func(context.Context, *Job) error) {
	w.ServeMux.HandleFunc(pattern, func(ctx context.Context, task *asynq.Task) error {
		return handler(ctx, &Job{
			task,
		})
	})
}

// ProcessTask dispatches the job to the handler whose pattern most closely matches the job type.
func (w *Worker) ProcessTask(ctx context.Context, job *Job) {
	w.ServeMux.ProcessTask(ctx, job.Task)
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
