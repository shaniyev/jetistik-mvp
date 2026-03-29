package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"jetistik/internal/platform/config"
	"jetistik/internal/sqlcdb"
	"jetistik/internal/storage"
)

const (
	TaskGenerateBatch = "task:generate_batch"
)

// GenerateBatchPayload is the payload for the generate_batch task.
type GenerateBatchPayload struct {
	BatchID int64 `json:"batch_id"`
}

// Client wraps the Asynq client for enqueuing tasks.
type Client struct {
	c *asynq.Client
}

// NewClient creates a new Asynq client.
func NewClient(redisURL string) (*Client, error) {
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return &Client{c: asynq.NewClient(opt)}, nil
}

// Close closes the Asynq client.
func (c *Client) Close() error {
	return c.c.Close()
}

// EnqueueGenerateBatch enqueues a batch generation task.
func (c *Client) EnqueueGenerateBatch(batchID int64) error {
	payload, _ := json.Marshal(GenerateBatchPayload{BatchID: batchID})
	task := asynq.NewTask(TaskGenerateBatch, payload)
	_, err := c.c.Enqueue(task)
	if err != nil {
		return fmt.Errorf("enqueue generate_batch: %w", err)
	}
	return nil
}

// Server represents the Asynq worker server.
type Server struct {
	srv     *asynq.Server
	mux     *asynq.ServeMux
	pool    *pgxpool.Pool
	storage *storage.Client
	cfg     *config.Config
	rdb     *redis.Client
}

// NewServer creates a new Asynq worker server.
func NewServer(pool *pgxpool.Pool, storageClient *storage.Client, cfg *config.Config) (*Server, error) {
	opt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	rdbOpt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url for pubsub: %w", err)
	}
	rdb := redis.NewClient(rdbOpt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	srv := asynq.NewServer(opt, asynq.Config{
		Concurrency: 5,
		Logger:      newAsynqLogger(),
	})

	s := &Server{
		srv:     srv,
		mux:     asynq.NewServeMux(),
		pool:    pool,
		storage: storageClient,
		cfg:     cfg,
		rdb:     rdb,
	}

	q := sqlcdb.New(pool)
	handler := &generateHandler{
		q:       q,
		storage: storageClient,
		cfg:     cfg,
		rdb:     rdb,
	}

	s.mux.HandleFunc(TaskGenerateBatch, handler.HandleGenerateBatch)

	return s, nil
}

// Run starts the Asynq worker server. Blocks until shutdown.
func (s *Server) Run() error {
	slog.Info("starting asynq worker")
	return s.srv.Run(s.mux)
}

// Shutdown gracefully stops the Asynq worker.
func (s *Server) Shutdown() {
	s.srv.Shutdown()
	s.rdb.Close()
}

// asynqLogger adapts slog for asynq.
type asynqLogger struct{}

func newAsynqLogger() *asynqLogger { return &asynqLogger{} }

func (l *asynqLogger) Debug(args ...interface{})                    { slog.Debug(fmt.Sprint(args...)) }
func (l *asynqLogger) Info(args ...interface{})                     { slog.Info(fmt.Sprint(args...)) }
func (l *asynqLogger) Warn(args ...interface{})                     { slog.Warn(fmt.Sprint(args...)) }
func (l *asynqLogger) Error(args ...interface{})                    { slog.Error(fmt.Sprint(args...)) }
func (l *asynqLogger) Fatal(args ...interface{})                    { slog.Error(fmt.Sprint(args...)) }
