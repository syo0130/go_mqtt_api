package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"go_mqtt_api/internal/domain"
)

// SensorDataService handles business logic for sensor data
type SensorDataService struct {
	sensorRepo domain.SensorDataRepository
	latestRepo domain.LatestValueRepository
}

// NewSensorDataService creates a new sensor data service
func NewSensorDataService(
	sensorRepo domain.SensorDataRepository,
	latestRepo domain.LatestValueRepository,
) *SensorDataService {
	return &SensorDataService{
		sensorRepo: sensorRepo,
		latestRepo: latestRepo,
	}
}

// ProcessSensorData processes incoming sensor data from MQTT
func (s *SensorDataService) ProcessSensorData(ctx context.Context, deviceID string, value float64, timestamp time.Time) error {
	// Create sensor data entity
	sensorData := &domain.SensorData{
		DeviceID:  deviceID,
		Value:     value,
		Timestamp: timestamp,
		CreatedAt: time.Now(),
	}

	// Save to PostgreSQL (history)
	if err := s.sensorRepo.Create(ctx, sensorData); err != nil {
		return fmt.Errorf("failed to save sensor data to PostgreSQL: %w", err)
	}

	// Update Redis (latest value)
	latestValue := &domain.LatestValue{
		DeviceID:  deviceID,
		Value:     value,
		Timestamp: timestamp,
	}
	if err := s.latestRepo.SetLatest(ctx, deviceID, latestValue); err != nil {
		log.Printf("Warning: failed to update latest value in Redis: %v", err)
		// Don't return error here, as PostgreSQL save succeeded
	}

	log.Printf("Processed sensor data: device=%s, value=%.2f, timestamp=%s", deviceID, value, timestamp.Format(time.RFC3339))
	return nil
}

// GetLatestValue retrieves the latest value for a device
func (s *SensorDataService) GetLatestValue(ctx context.Context, deviceID string) (*domain.LatestValue, error) {
	latest, err := s.latestRepo.GetLatest(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest value: %w", err)
	}
	if latest == nil {
		return nil, fmt.Errorf("latest value not found for device: %s", deviceID)
	}
	return latest, nil
}

// GetHistory retrieves sensor data history for a device
func (s *SensorDataService) GetHistory(ctx context.Context, deviceID string, limit, offset int) ([]*domain.SensorData, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Max limit
	}

	records, err := s.sensorRepo.GetHistory(ctx, deviceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	return records, nil
}

// GetHistoryCount returns the total count of records for a device
func (s *SensorDataService) GetHistoryCount(ctx context.Context, deviceID string) (int64, error) {
	count, err := s.sensorRepo.Count(ctx, deviceID)
	if err != nil {
		return 0, fmt.Errorf("failed to get history count: %w", err)
	}
	return count, nil
}
