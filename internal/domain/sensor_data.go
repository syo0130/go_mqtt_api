package domain

import "time"

// SensorData represents sensor data entity
type SensorData struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DeviceID  string    `gorm:"index;not null" json:"device_id"`
	Value     float64   `gorm:"not null" json:"value"`
	Timestamp time.Time `gorm:"index;not null" json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for GORM
func (SensorData) TableName() string {
	return "sensor_data"
}

// LatestValue represents the latest value for a device (used for Redis)
type LatestValue struct {
	DeviceID  string    `json:"device_id"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

