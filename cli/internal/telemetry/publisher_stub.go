//go:build !zerobus

package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Publisher handles publishing events to zerobus
// This is a stub implementation used when CGO is disabled.
type Publisher struct {
	clientID        string
	clientSecret    string
	endpoint        string
	unityCatalogURL string
	tableName       string
}

// NewPublisher creates a new telemetry publisher using build-time credentials
func NewPublisher() (*Publisher, error) {
	if ZerobusClientID == "" || ZerobusClientSecret == "" {
		return nil, fmt.Errorf("zerobus credentials not configured (binary was not built with credentials)")
	}
	if ZerobusEndpoint == "" {
		return nil, fmt.Errorf("zerobus endpoint not configured")
	}
	if ZerobusUnityCatalogURL == "" {
		return nil, fmt.Errorf("zerobus Unity Catalog URL not configured")
	}
	if ZerobusTableName == "" {
		return nil, fmt.Errorf("zerobus table name not configured")
	}

	return &Publisher{
		clientID:        ZerobusClientID,
		clientSecret:    ZerobusClientSecret,
		endpoint:        ZerobusEndpoint,
		unityCatalogURL: ZerobusUnityCatalogURL,
		tableName:       ZerobusTableName,
	}, nil
}

// Publish sends an event to zerobus
// This is a stub implementation that returns an error when CGO is disabled.
func (p *Publisher) Publish(ctx context.Context, event *Event) error {
	// Set defaults
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if event.User == "" {
		user := os.Getenv("USER")
		if user != "" {
			// Use VIBE_USER_EMAIL if set, otherwise fall back to just the system username
			if email := os.Getenv("VIBE_USER_EMAIL"); email != "" {
				event.User = email
			} else {
				event.User = user
			}
		}
	}

	// In stub mode, we just validate we can serialize
	_, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return fmt.Errorf("zerobus publishing not available: binary was built without CGO support")
}

// PublishJSON publishes a raw JSON payload with the given event type
func (p *Publisher) PublishJSON(ctx context.Context, eventType string, jsonPayload []byte) error {
	// Validate that the payload is valid JSON
	var payload map[string]interface{}
	if err := json.Unmarshal(jsonPayload, &payload); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}

	event := &Event{
		EventType: eventType,
		Payload:   string(jsonPayload), // Store as JSON string
	}

	return p.Publish(ctx, event)
}
