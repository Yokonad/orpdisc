package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Yokonad/orpdisc/internal/config"
	"github.com/Yokonad/orpdisc/internal/database"
	"github.com/Yokonad/orpdisc/internal/discord"
	"github.com/Yokonad/orpdisc/internal/models"
	"github.com/Yokonad/orpdisc/internal/openrouter"
	"github.com/Yokonad/orpdisc/internal/processor"
)

// Logger defines the logging interface used by the service
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// StdLogger implements Logger using the standard log package
type StdLogger struct {
	level  string
 Debugf func(format string, args ...interface{})
	Infof  func(format string, args ...interface{})
	Warnf  func(format string, args ...interface{})
	Errorf func(format string, args ...interface{})
}

// NewStdLogger creates a new standard logger
func NewStdLogger(level string) *StdLogger {
	return &StdLogger{
		level: level,
		Debugf: func(format string, args ...interface{}) {
			if level == "debug" {
				fmt.Printf(format+"\n", args...)
			}
		},
		Infof: func(format string, args ...interface{}) {
			if level == "debug" || level == "info" {
				fmt.Printf(format+"\n", args...)
			}
		},
		Warnf: func(format string, args ...interface{}) {
			if level == "debug" || level == "info" || level == "warn" {
				fmt.Printf(format+"\n", args...)
			}
		},
		Errorf: func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		},
	}
}

// Debug logs a debug message
func (l *StdLogger) Debug(msg string, args ...interface{}) {
	l.Debugf("[DEBUG] "+msg, args...)
}

// Info logs an info message
func (l *StdLogger) Info(msg string, args ...interface{}) {
	l.Infof("[INFO] "+msg, args...)
}

// Warn logs a warning message
func (l *StdLogger) Warn(msg string, args ...interface{}) {
	l.Warnf("[WARN] "+msg, args...)
}

// Error logs an error message
func (l *StdLogger) Error(msg string, args ...interface{}) {
	l.Errorf("[ERROR] "+msg, args...)
}

// Service is the main application service that orchestrates all components
type Service struct {
	cfg      *config.Config
	db       *database.DB
	client   *openrouter.Client
	proc     *processor.Processor
	webhook  *discord.WebhookClient
	logger   Logger

	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	stopChan  chan struct{}
	stopped   bool
	stopMutex sync.Mutex
}

// NewService creates a new Service instance with provided dependencies
func NewService(cfg *config.Config, db *database.DB, client *openrouter.Client, proc *processor.Processor, webhook *discord.WebhookClient, logger Logger) (*Service, error) {
	if logger == nil {
		logger = NewStdLogger(cfg.LogLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		cfg:      cfg,
		db:       db,
		client:   client,
		proc:     proc,
		webhook:  webhook,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
		stopChan: make(chan struct{}),
	}, nil
}

// Start begins the service's polling loop
func (s *Service) Start() error {
	s.logger.Info("Starting OpenRouter Discord Monitor service")
	s.logger.Info("Poll interval: %s", s.cfg.PollInterval)
	s.logger.Info("Discord webhook: %s", s.cfg.RedactedWebhookURL())

	// Create ticker for polling
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()

	// Run initial poll immediately
	s.poll()

	// Main polling loop
	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Service context cancelled, stopping...")
			return nil
		case <-ticker.C:
			s.poll()
		case <-s.stopChan:
			s.logger.Info("Service stop signal received")
			return nil
		}
	}
}

// poll performs a single poll cycle: fetch -> process -> notify
func (s *Service) poll() {
	s.logger.Debug("Starting poll cycle")

	// Fetch models from OpenRouter
	apiModels, err := s.client.FetchModels(s.ctx)
	if err != nil {
		s.logger.Error("Failed to fetch models: %v", err)
		return
	}

	s.logger.Info("Fetched %d text models from OpenRouter", len(apiModels))

	// Process models and detect changes
	changeset, err := s.proc.ProcessModels(s.ctx, apiModels)
	if err != nil {
		s.logger.Error("Failed to process models: %v", err)
		return
	}

	if !changeset.HasChanges() {
		s.logger.Info("No changes detected in this poll cycle")
		return
	}

	s.logger.Info("Detected %d changes: %d new, %d updated, %d removed",
		changeset.TotalChanges(),
		len(changeset.NewModels),
		len(changeset.UpdatedModels),
		len(changeset.RemovedModels))

	// Send Discord notification
	if err := s.webhook.SendNotification(s.ctx, changeset); err != nil {
		s.logger.Error("Failed to send Discord notification: %v", err)
		return
	}

	s.logger.Info("Discord notification sent successfully")
}

// Stop gracefully stops the service
func (s *Service) Stop() error {
	s.stopMutex.Lock()
	if s.stopped {
		s.stopMutex.Unlock()
		return nil
	}
	s.stopped = true
	s.stopMutex.Unlock()

	s.logger.Info("Stopping service...")

	// Cancel context to stop in-flight operations
	s.cancel()

	// Close database connection
	if err := s.db.Close(); err != nil {
		s.logger.Error("Failed to close database: %v", err)
	}

	s.logger.Info("Service stopped gracefully")
	return nil
}

// WaitForSignal waits for termination signals and triggers graceful shutdown
func (s *Service) WaitForSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigChan
	s.logger.Info("Received signal: %s", sig)

	// Trigger stop
	close(s.stopChan)

	// Give the service time to stop gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		s.Stop()
		close(done)
	}()

	select {
	case <-shutdownCtx.Done():
		s.logger.Warn("Shutdown timeout exceeded, forcing exit")
	case <-done:
		s.logger.Info("Graceful shutdown completed")
	}
}

// HealthCheck returns the health status of the service
func (s *Service) HealthCheck() error {
	// Check database connectivity
	if err := s.db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// HealthCheckServer starts an optional HTTP server for health checks
func (s *Service) HealthCheckServer(addr string) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := s.HealthCheck(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "unhealthy: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "healthy")
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return server
}

// SendDigest sends a daily digest with top models
func (s *Service) SendDigest(ctx context.Context) error {
	allModels, err := s.db.GetAllModels()
	if err != nil {
		return fmt.Errorf("failed to get models for digest: %w", err)
	}

	if len(allModels) == 0 {
		return nil
	}

	// Get top 1 by cost
	topByCost := processor.TopByCostPer1K(allModels, 1)

	// Get top 1 by context/cost ratio
	topByRatio := processor.TopByContextCostRatio(allModels, 1)

	changeset := &models.Changeset{
		NewModels:     topByCost,
		UpdatedModels: topByRatio,
		IsDigest:      true,
	}

	// Log digest info
	s.logger.Info("Sending digest: %d model by cost, %d by context/cost ratio",
		len(topByCost), len(topByRatio))

	return s.webhook.SendNotification(ctx, changeset)
}
