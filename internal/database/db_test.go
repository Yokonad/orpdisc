package database

import (
	"os"
	"testing"
	"time"

	"github.com/Yokonad/orpdisc/internal/models"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	// Use a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_db_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.Migrate(); err != nil {
		db.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to migrate database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return db, cleanup
}

func TestNew(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_new_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	if db.DB == nil {
		t.Error("New() returned nil database connection")
	}
}

func TestNewInvalidPath(t *testing.T) {
	_, err := New("/nonexistent/path/to/file.db")
	if err == nil {
		t.Error("New() expected error for invalid path, got nil")
	}
}

func TestMigrate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Verify tables exist by querying them
	var tableCount int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='models'").Scan(&tableCount)
	if err != nil {
		t.Fatalf("Failed to check tables: %v", err)
	}

	if tableCount != 1 {
		t.Errorf("models table count = %d, want 1", tableCount)
	}
}

func TestGetModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Test with non-existent model
	model, err := db.GetModel("nonexistent")
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}
	if model != nil {
		t.Error("GetModel() returned model for non-existent ID")
	}

	// Insert a model and retrieve it
	testModel := &models.Model{
		ID:                   "test/model",
		Name:                 "Test Model",
		Provider:             "test",
		Description:          "A test model",
		ContextLength:        1000,
		MaxCompletionTokens:  500,
		PricingPrompt:        0.01,
		PricingCompletion:    0.02,
		Tokenizer:            "test-tokenizer",
		DataHash:             "abc123",
		FirstSeen:            time.Now(),
		LastUpdated:          time.Now(),
	}

	if err := db.SaveModel(testModel); err != nil {
		t.Fatalf("SaveModel() error = %v", err)
	}

	model, err = db.GetModel("test/model")
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}
	if model == nil {
		t.Fatal("GetModel() returned nil for existing model")
	}

	if model.ID != testModel.ID {
		t.Errorf("Model ID = %s, want %s", model.ID, testModel.ID)
	}
	if model.Name != testModel.Name {
		t.Errorf("Model Name = %s, want %s", model.Name, testModel.Name)
	}
}

func TestGetAllModels(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Initially should be empty
	modelsList, err := db.GetAllModels()
	if err != nil {
		t.Fatalf("GetAllModels() error = %v", err)
	}
	if len(modelsList) != 0 {
		t.Errorf("GetAllModels() returned %d models, want 0", len(modelsList))
	}

	// Add some models
	for i := 0; i < 3; i++ {
		model := &models.Model{
			ID:       "test/model" + string(rune('a'+i)),
			Name:     "Test Model" + string(rune('A'+i)),
			Provider: "test",
			DataHash: "hash" + string(rune('0'+i)),
		}
		if err := db.SaveModel(model); err != nil {
			t.Fatalf("SaveModel() error = %v", err)
		}
	}

	modelsList, err = db.GetAllModels()
	if err != nil {
		t.Fatalf("GetAllModels() error = %v", err)
	}
	if len(modelsList) != 3 {
		t.Errorf("GetAllModels() returned %d models, want 3", len(modelsList))
	}
}

func TestSaveModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	model := &models.Model{
		ID:                   "test/model",
		Name:                 "Test Model",
		Provider:             "test",
		Description:          "A test model",
		ContextLength:        1000,
		MaxCompletionTokens:  500,
		PricingPrompt:        0.01,
		PricingCompletion:    0.02,
		Tokenizer:            "test-tokenizer",
		DataHash:             "abc123",
	}

	if err := db.SaveModel(model); err != nil {
		t.Fatalf("SaveModel() error = %v", err)
	}

	// Verify it was saved
	retrieved, err := db.GetModel("test/model")
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}
	if retrieved.Name != model.Name {
		t.Errorf("Model Name = %s, want %s", retrieved.Name, model.Name)
	}
}

func TestSaveModelUpdate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert initial model
	model := &models.Model{
		ID:                "test/model",
		Name:              "Original Name",
		Provider:          "test",
		PricingPrompt:     0.01,
		PricingCompletion: 0.02,
		DataHash:          "original-hash",
	}

	if err := db.SaveModel(model); err != nil {
		t.Fatalf("SaveModel() error = %v", err)
	}

	// Update the model
	model.Name = "Updated Name"
	model.PricingPrompt = 0.015
	model.DataHash = "updated-hash"

	if err := db.SaveModel(model); err != nil {
		t.Fatalf("SaveModel() update error = %v", err)
	}

	// Verify update
	retrieved, err := db.GetModel("test/model")
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}
	if retrieved.Name != "Updated Name" {
		t.Errorf("Model Name = %s, want Updated Name", retrieved.Name)
	}
	if retrieved.PricingPrompt != 0.015 {
		t.Errorf("Model PricingPrompt = %f, want 0.015", retrieved.PricingPrompt)
	}
}

func TestSavePriceHistory(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// First save a model
	model := &models.Model{
		ID:       "test/model",
		Name:     "Test Model",
		Provider: "test",
		DataHash: "hash",
	}
	if err := db.SaveModel(model); err != nil {
		t.Fatalf("SaveModel() error = %v", err)
	}

	// Save price history
	err := db.SavePriceHistory("test/model", 0.01, 0.02)
	if err != nil {
		t.Fatalf("SavePriceHistory() error = %v", err)
	}

	// We can't easily verify the price history without adding a getter,
	// but we can verify the insert didn't error
}

func TestLogNotification(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	notification := &models.Notification{
		Type:     models.NotificationTypeNew,
		ModelID:  "test/model",
		OldValue: "",
		NewValue: "",
		Success:  true,
	}

	err := db.LogNotification(notification)
	if err != nil {
		t.Fatalf("LogNotification() error = %v", err)
	}

	if notification.ID == 0 {
		t.Error("LogNotification() did not set notification ID")
	}
}

func TestClose(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_close_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestPing(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	err := db.Ping()
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}
