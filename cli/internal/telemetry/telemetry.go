// Package telemetry provides event publishing for vibe usage tracking.
package telemetry

import (
	"time"
)

// Build-time variables for telemetry credentials (injected via ldflags)
var (
	ZerobusClientID        = ""
	ZerobusClientSecret    = ""
	ZerobusEndpoint        = ""
	ZerobusUnityCatalogURL = ""
	ZerobusTableName       = ""
)

// Event represents a telemetry event to publish
// Note: Payload is stored as a JSON string for compatibility with
// ingestion systems that don't support nested objects.
type Event struct {
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   string    `json:"payload"` // JSON-serialized payload string
	User      string    `json:"user,omitempty"`
	Source    string    `json:"source,omitempty"`
}

// IsConfigured returns true if telemetry credentials are available
func IsConfigured() bool {
	return ZerobusClientID != "" &&
		ZerobusClientSecret != "" &&
		ZerobusEndpoint != "" &&
		ZerobusUnityCatalogURL != "" &&
		ZerobusTableName != ""
}

// GetConfiguration returns the current telemetry configuration (with secret redacted)
func GetConfiguration() map[string]string {
	secret := "(not set)"
	if ZerobusClientSecret != "" {
		secret = "(set)"
	}
	return map[string]string{
		"client_id":         ZerobusClientID,
		"secret":            secret,
		"endpoint":          ZerobusEndpoint,
		"unity_catalog_url": ZerobusUnityCatalogURL,
		"table":             ZerobusTableName,
	}
}
