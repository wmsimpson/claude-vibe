//go:build zerobus

package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	zerobus "github.com/databricks/zerobus-sdk-go"
)

// Publisher handles publishing events to zerobus
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

	// Serialize event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create zerobus SDK
	sdk, err := zerobus.NewZerobusSdk(p.endpoint, p.unityCatalogURL)
	if err != nil {
		return fmt.Errorf("failed to create zerobus SDK: %w", err)
	}
	defer sdk.Free()

	// Create table properties
	tableProps := zerobus.TableProperties{
		TableName: p.tableName,
	}

	// Create stream options for JSON records
	options := zerobus.DefaultStreamConfigurationOptions()
	options.RecordType = zerobus.RecordTypeJson

	// Create stream
	stream, err := sdk.CreateStream(tableProps, p.clientID, p.clientSecret, options)
	if err != nil {
		return fmt.Errorf("failed to create zerobus stream: %w", err)
	}
	defer stream.Close()

	// Write the event using IngestRecordOffset (recommended API)
	_, err = stream.IngestRecordOffset(string(data))
	if err != nil {
		return fmt.Errorf("failed to write event to zerobus: %w", err)
	}

	return nil
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
