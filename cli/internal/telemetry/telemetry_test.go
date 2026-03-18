package telemetry

import (
	"testing"
)

func TestIsConfigured(t *testing.T) {
	// Save original values
	origClientID := ZerobusClientID
	origClientSecret := ZerobusClientSecret
	origEndpoint := ZerobusEndpoint
	origUnityCatalogURL := ZerobusUnityCatalogURL
	origTableName := ZerobusTableName

	// Restore after test
	defer func() {
		ZerobusClientID = origClientID
		ZerobusClientSecret = origClientSecret
		ZerobusEndpoint = origEndpoint
		ZerobusUnityCatalogURL = origUnityCatalogURL
		ZerobusTableName = origTableName
	}()

	// Test with no configuration
	ZerobusClientID = ""
	ZerobusClientSecret = ""
	ZerobusEndpoint = ""
	ZerobusUnityCatalogURL = ""
	ZerobusTableName = ""

	if IsConfigured() {
		t.Error("IsConfigured() should return false when no credentials are set")
	}

	// Test with partial configuration
	ZerobusClientID = "test-client-id"
	if IsConfigured() {
		t.Error("IsConfigured() should return false when only client ID is set")
	}

	// Test with full configuration
	ZerobusClientID = "test-client-id"
	ZerobusClientSecret = "test-secret"
	ZerobusEndpoint = "https://zerobus.test.com"
	ZerobusUnityCatalogURL = "https://workspace.test.com"
	ZerobusTableName = "catalog.schema.table"

	if !IsConfigured() {
		t.Error("IsConfigured() should return true when all credentials are set")
	}
}

func TestGetConfiguration(t *testing.T) {
	// Save original values
	origClientID := ZerobusClientID
	origClientSecret := ZerobusClientSecret
	origEndpoint := ZerobusEndpoint
	origUnityCatalogURL := ZerobusUnityCatalogURL
	origTableName := ZerobusTableName

	// Restore after test
	defer func() {
		ZerobusClientID = origClientID
		ZerobusClientSecret = origClientSecret
		ZerobusEndpoint = origEndpoint
		ZerobusUnityCatalogURL = origUnityCatalogURL
		ZerobusTableName = origTableName
	}()

	// Test with configuration
	ZerobusClientID = "test-client-id"
	ZerobusClientSecret = "test-secret"
	ZerobusEndpoint = "https://zerobus.test.com"
	ZerobusUnityCatalogURL = "https://workspace.test.com"
	ZerobusTableName = "catalog.schema.table"

	cfg := GetConfiguration()

	if cfg["client_id"] != "test-client-id" {
		t.Errorf("Expected client_id 'test-client-id', got '%s'", cfg["client_id"])
	}

	// Secret should be redacted
	if cfg["secret"] != "(set)" {
		t.Errorf("Expected secret '(set)', got '%s'", cfg["secret"])
	}

	if cfg["endpoint"] != "https://zerobus.test.com" {
		t.Errorf("Expected endpoint 'https://zerobus.test.com', got '%s'", cfg["endpoint"])
	}

	if cfg["table"] != "catalog.schema.table" {
		t.Errorf("Expected table 'catalog.schema.table', got '%s'", cfg["table"])
	}

	// Test with no secret
	ZerobusClientSecret = ""
	cfg = GetConfiguration()
	if cfg["secret"] != "(not set)" {
		t.Errorf("Expected secret '(not set)', got '%s'", cfg["secret"])
	}
}

func TestNewPublisher(t *testing.T) {
	// Save original values
	origClientID := ZerobusClientID
	origClientSecret := ZerobusClientSecret
	origEndpoint := ZerobusEndpoint
	origUnityCatalogURL := ZerobusUnityCatalogURL
	origTableName := ZerobusTableName

	// Restore after test
	defer func() {
		ZerobusClientID = origClientID
		ZerobusClientSecret = origClientSecret
		ZerobusEndpoint = origEndpoint
		ZerobusUnityCatalogURL = origUnityCatalogURL
		ZerobusTableName = origTableName
	}()

	// Test with no configuration
	ZerobusClientID = ""
	ZerobusClientSecret = ""
	ZerobusEndpoint = ""
	ZerobusUnityCatalogURL = ""
	ZerobusTableName = ""

	_, err := NewPublisher()
	if err == nil {
		t.Error("NewPublisher() should return error when credentials not configured")
	}

	// Test with full configuration
	ZerobusClientID = "test-client-id"
	ZerobusClientSecret = "test-secret"
	ZerobusEndpoint = "https://zerobus.test.com"
	ZerobusUnityCatalogURL = "https://workspace.test.com"
	ZerobusTableName = "catalog.schema.table"

	publisher, err := NewPublisher()
	if err != nil {
		t.Errorf("NewPublisher() should not return error with full config: %v", err)
	}
	if publisher == nil {
		t.Error("NewPublisher() should return a publisher")
	}
}
