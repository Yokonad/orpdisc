package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Yokonad/orpdisc/internal/models"
	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection
type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// Migrate creates the database schema if it doesn't exist
func (db *DB) Migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS models (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		provider TEXT NOT NULL,
		description TEXT,
		context_length INTEGER,
		max_completion_tokens INTEGER,
		pricing_prompt REAL,
		pricing_completion REAL,
		tokenizer TEXT,
		data_hash TEXT NOT NULL,
		first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_updated DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS price_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		model_id TEXT NOT NULL,
		pricing_prompt REAL,
		pricing_completion REAL,
		recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (model_id) REFERENCES models(id)
	);

	CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		model_id TEXT,
		old_value TEXT,
		new_value TEXT,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		success BOOLEAN DEFAULT FALSE
	);

	CREATE INDEX IF NOT EXISTS idx_models_provider ON models(provider);
	CREATE INDEX IF NOT EXISTS idx_price_history_model ON price_history(model_id);
	CREATE INDEX IF NOT EXISTS idx_notifications_sent ON notifications(sent_at);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// GetModel retrieves a single model by ID
func (db *DB) GetModel(id string) (*models.Model, error) {
	query := `
		SELECT id, name, provider, description, context_length, max_completion_tokens,
			   pricing_prompt, pricing_completion, tokenizer, data_hash, first_seen, last_updated
		FROM models WHERE id = ?`

	var model models.Model
	err := db.QueryRow(query, id).Scan(
		&model.ID, &model.Name, &model.Provider, &model.Description,
		&model.ContextLength, &model.MaxCompletionTokens,
		&model.PricingPrompt, &model.PricingCompletion, &model.Tokenizer,
		&model.DataHash, &model.FirstSeen, &model.LastUpdated,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	return &model, nil
}

// GetAllModels retrieves all models from the database
func (db *DB) GetAllModels() ([]models.Model, error) {
	query := `
		SELECT id, name, provider, description, context_length, max_completion_tokens,
			   pricing_prompt, pricing_completion, tokenizer, data_hash, first_seen, last_updated
		FROM models`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query models: %w", err)
	}
	defer rows.Close()

	var modelList []models.Model
	for rows.Next() {
		var model models.Model
		err := rows.Scan(
			&model.ID, &model.Name, &model.Provider, &model.Description,
			&model.ContextLength, &model.MaxCompletionTokens,
			&model.PricingPrompt, &model.PricingCompletion, &model.Tokenizer,
			&model.DataHash, &model.FirstSeen, &model.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		modelList = append(modelList, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating models: %w", err)
	}

	return modelList, nil
}

// SaveModel inserts or updates a model
func (db *DB) SaveModel(model *models.Model) error {
	query := `
		INSERT INTO models (id, name, provider, description, context_length, max_completion_tokens,
							pricing_prompt, pricing_completion, tokenizer, data_hash, first_seen, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			provider = excluded.provider,
			description = excluded.description,
			context_length = excluded.context_length,
			max_completion_tokens = excluded.max_completion_tokens,
			pricing_prompt = excluded.pricing_prompt,
			pricing_completion = excluded.pricing_completion,
			tokenizer = excluded.tokenizer,
			data_hash = excluded.data_hash,
			last_updated = excluded.last_updated`

	now := time.Now().UTC()
	model.LastUpdated = now

	_, err := db.Exec(query,
		model.ID, model.Name, model.Provider, model.Description,
		model.ContextLength, model.MaxCompletionTokens,
		model.PricingPrompt, model.PricingCompletion, model.Tokenizer,
		model.DataHash, model.FirstSeen, model.LastUpdated,
	)
	if err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}

	return nil
}

// SavePriceHistory records a price change for a model
func (db *DB) SavePriceHistory(modelID string, pricingPrompt, pricingCompletion float64) error {
	query := `
		INSERT INTO price_history (model_id, pricing_prompt, pricing_completion)
		VALUES (?, ?, ?)`

	_, err := db.Exec(query, modelID, pricingPrompt, pricingCompletion)
	if err != nil {
		return fmt.Errorf("failed to save price history: %w", err)
	}

	return nil
}

// LogNotification records a notification in the audit log
func (db *DB) LogNotification(notification *models.Notification) error {
	query := `
		INSERT INTO notifications (type, model_id, old_value, new_value, success)
		VALUES (?, ?, ?, ?, ?)`

	result, err := db.Exec(query,
		notification.Type, notification.ModelID,
		notification.OldValue, notification.NewValue, notification.Success,
	)
	if err != nil {
		return fmt.Errorf("failed to log notification: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	notification.ID = id

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Ping checks the database connection is alive
func (db *DB) Ping() error {
	return db.DB.Ping()
}
