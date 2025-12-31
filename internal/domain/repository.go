package domain

import "context"

// SensorDataRepository defines the interface for sensor data persistence (PostgreSQL)
type SensorDataRepository interface {
	// Create saves a new sensor data record
	Create(ctx context.Context, data *SensorData) error
	// GetHistory retrieves sensor data history for a device
	GetHistory(ctx context.Context, deviceID string, limit, offset int) ([]*SensorData, error)
	// Count returns the total number of records for a device
	Count(ctx context.Context, deviceID string) (int64, error)
}

// LatestValueRepository defines the interface for latest value cache (Redis)
type LatestValueRepository interface {
	// SetLatest updates the latest value for a device
	SetLatest(ctx context.Context, deviceID string, value *LatestValue) error
	// GetLatest retrieves the latest value for a device
	GetLatest(ctx context.Context, deviceID string) (*LatestValue, error)
}

