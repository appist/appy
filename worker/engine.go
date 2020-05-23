package worker

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/appist/appy/mock"
	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
	"github.com/go-redis/redis"
	"github.com/hibiken/asynq"
)

// Engine processes the background jobs.
type Engine struct {
	*asynq.Server
	*asynq.Client
	*asynq.ServeMux
	asynq.RedisConnOpt
	asset     *support.Asset
	config    *support.Config
	dbManager *record.Engine
	jobs      []*Job
	logger    *support.Logger
	mu        *sync.Mutex
}

// Handler processes background jobs.
//
// ProcessTask should return nil if the processing of a background job is
// successful.
//
// If ProcessTask return a non-nil error or panics, the background job will
// be retried after delay.
type Handler = asynq.Handler

// HandlerFunc is an adapter to allow the use of ordinary functions as a
// Handler. If f is a function with the appropriate signature, HandlerFunc(f)
// is a Handler that calls f.
type HandlerFunc = asynq.HandlerFunc

// MiddlewareFunc is a function which receives an Handler and returns
// another Handler. Typically, the returned handler is a closure which does
// something with the context and task passed to it, and then calls the handler
// passed as parameter to the MiddlewareFunc.
type MiddlewareFunc = asynq.MiddlewareFunc

// NewEngine initializes a worker to process background jobs.
func NewEngine(asset *support.Asset, config *support.Config, dbManager *record.Engine, l *support.Logger) *Engine {
	workerLogger := &logger{
		l,
		nil,
	}

	workerConfig := asynq.Config{
		Concurrency:     config.WorkerConcurrency,
		Logger:          workerLogger,
		Queues:          config.WorkerQueues,
		StrictPriority:  config.WorkerStrictPriority,
		ShutdownTimeout: config.WorkerGracefulShutdownTimeout,
	}

	redisConnOpt := &asynq.RedisClientOpt{
		Addr:     config.WorkerRedisAddr,
		Password: config.WorkerRedisPassword,
		DB:       config.WorkerRedisDB,
		PoolSize: config.WorkerRedisPoolSize,
	}

	if config.WorkerRedisURL != "" {
		redisParseURL, err := redis.ParseURL(config.WorkerRedisURL)
		if err != nil {
			l.Fatal(err)
		}

		redisConnOpt.Addr = redisParseURL.Addr
		redisConnOpt.Password = redisParseURL.Password
		redisConnOpt.DB = redisParseURL.DB
	}

	worker := &Engine{
		asynq.NewServer(redisConnOpt, workerConfig),
		asynq.NewClient(redisConnOpt),
		asynq.NewServeMux(),
		redisConnOpt,
		asset,
		config,
		dbManager,
		[]*Job{},
		l,
		&sync.Mutex{},
	}

	if len(config.WorkerRedisSentinelAddrs) > 0 {
		redisConnOpt := &asynq.RedisFailoverClientOpt{
			MasterName:       config.WorkerRedisSentinelMasterName,
			SentinelAddrs:    config.WorkerRedisSentinelAddrs,
			SentinelPassword: config.WorkerRedisSentinelPassword,
			DB:               config.WorkerRedisSentinelDB,
			PoolSize:         config.WorkerRedisSentinelPoolSize,
		}

		worker = &Engine{
			asynq.NewServer(redisConnOpt, workerConfig),
			asynq.NewClient(redisConnOpt),
			asynq.NewServeMux(),
			redisConnOpt,
			asset,
			config,
			dbManager,
			[]*Job{},
			l,
			&sync.Mutex{},
		}
	}

	workerLogger.worker = worker
	worker.ServeMux.Use(func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			payload := strings.ReplaceAll(fmt.Sprintf("%+v", task.Payload), "{data:map[", "")
			payload = strings.ReplaceAll(fmt.Sprintf("%+v", payload), "]}", "")
			l.Infof(`[WORKER] job: %s, payload: (%s) start`, task.Type, payload)

			start := time.Now()
			err := next.ProcessTask(ctx, task)
			l.Infof(`[WORKER] job: %s, payload: (%s) done in %s`, task.Type, payload, time.Since(start))

			return err
		})
	})

	return worker
}

// Drain simulates job processing by removing them, only available for unit
// test with APPY_ENV=test.
func (w *Engine) Drain() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.jobs = []*Job{}
}

// Enqueue enqueues job to be processed immediately.
//
// Enqueue returns nil if the job is enqueued successfully, otherwise returns
// an error.
func (w *Engine) Enqueue(job *Job, opts *JobOptions) error {
	if w.config.AppyEnv == "test" {
		w.mu.Lock()
		defer w.mu.Unlock()

		w.jobs = append(w.jobs, job)
		return nil
	}

	return w.Client.Enqueue(job, parseJobOptions(opts))
}

// EnqueueAt schedules job to be enqueued at the specified time.
//
// It returns nil if the job is scheduled successfully, otherwise returns an
// error.
func (w *Engine) EnqueueAt(t time.Time, job *Job, opts *JobOptions) error {
	if w.config.AppyEnv == "test" {
		w.mu.Lock()
		defer w.mu.Unlock()

		w.jobs = append(w.jobs, job)
		return nil
	}

	return w.Client.EnqueueAt(t, job, parseJobOptions(opts))
}

// EnqueueIn schedules job to be enqueued after the specified delay.
//
// It returns nil if the job is scheduled successfully, otherwise returns an
// error.
func (w *Engine) EnqueueIn(d time.Duration, job *Job, opts *JobOptions) error {
	if w.config.AppyEnv == "test" {
		w.mu.Lock()
		defer w.mu.Unlock()

		w.jobs = append(w.jobs, job)
		return nil
	}

	return w.Client.EnqueueIn(d, job, parseJobOptions(opts))
}

// Jobs returns the enqueued jobs, only available for unit test with
// APPY_ENV=test.
func (w *Engine) Jobs() []*Job {
	return w.jobs
}

// ProcessTask dispatches the job to the handler whose pattern most closely
// matches the job type.
func (w *Engine) ProcessTask(ctx context.Context, job *Job) {
	w.ServeMux.ProcessTask(ctx, job)
}

// Use appends a MiddlewareFunc to the chain. Middlewares are executed
// in the order that they are applied to the ServeMux.
func (w *Engine) Use(handlers ...MiddlewareFunc) {
	w.ServeMux.Use(handlers...)
}

// Info returns the worker info.
func (w *Engine) Info() []string {
	configPath := w.config.Path()

	lines := []string{}
	lines = append(lines,
		fmt.Sprintf("* appy %s (%s), build: %s, environment: %s, config: %s",
			support.VERSION, runtime.Version(), support.Build, w.config.AppyEnv, configPath,
		),
	)

	queues := []string{}
	for key, value := range w.config.WorkerQueues {
		queues = append(queues, fmt.Sprintf("%s=%d", key, value))
	}

	lines = append(lines, w.dbManager.Info())
	lines = append(lines, fmt.Sprintf("* Concurrency: %d, queues: %s", w.config.WorkerConcurrency, strings.Join(queues, ", ")))
	return append(lines, "* Worker is now ready to process jobs...")
}

// Run starts running the worker to process background jobs.
func (w *Engine) Run() {
	w.Server.Run(w.ServeMux)
}

// MockedHandler is used for mocking in unit test.
type MockedHandler struct {
	mock.Mock
}

// NewMockedHandler initializes a mocked Handler instance that is useful for unit test.
func NewMockedHandler() *MockedHandler {
	return new(MockedHandler)
}

// ProcessTask mocks the job processing functionality that is useful for unit
// test.
func (h *MockedHandler) ProcessTask(ctx context.Context, job *Job) error {
	args := h.Called(ctx, job)

	return args.Error(0)
}
