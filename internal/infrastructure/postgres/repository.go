package postgres

import (
	"context"
	"fmt"

	"go_mqtt_api/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SensorDataRepository implements domain.SensorDataRepository using PostgreSQL
type SensorDataRepository struct {
	db *gorm.DB
}

// NewSensorDataRepository creates a new PostgreSQL repository
func NewSensorDataRepository(dsn string) (*SensorDataRepository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&domain.SensorData{}); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}

	return &SensorDataRepository{db: db}, nil
}

// Create saves a new sensor data record
func (r *SensorDataRepository) Create(ctx context.Context, data *domain.SensorData) error {
	return r.db.WithContext(ctx).Create(data).Error
}

// GetHistory retrieves sensor data history for a device
func (r *SensorDataRepository) GetHistory(ctx context.Context, deviceID string, limit, offset int) ([]*domain.SensorData, error) {
	var records []*domain.SensorData
	err := r.db.WithContext(ctx).
		Where("device_id = ?", deviceID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error
	return records, err
}

// Count returns the total number of records for a device
func (r *SensorDataRepository) Count(ctx context.Context, deviceID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.SensorData{}).
		Where("device_id = ?", deviceID).
		Count(&count).Error
	return count, err
}

// Close closes the database connection
func (r *SensorDataRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

