package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"go_mqtt_api/internal/domain"

	"github.com/redis/go-redis/v9"
)

// LatestValueRepository implements domain.LatestValueRepository using Redis
type LatestValueRepository struct {
	client *redis.Client
}

// NewLatestValueRepository creates a new Redis repository
func NewLatestValueRepository(addr, password string, db int) (*LatestValueRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &LatestValueRepository{client: client}, nil
}

// getKey returns the Redis key for a device
func (r *LatestValueRepository) getKey(deviceID string) string {
	return fmt.Sprintf("device:%s:latest", deviceID)
}

// SetLatest updates the latest value for a device
func (r *LatestValueRepository) SetLatest(ctx context.Context, deviceID string, value *domain.LatestValue) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal latest value: %w", err)
	}

	key := r.getKey(deviceID)
	// Set with no expiration (or set a long expiration if needed)
	return r.client.Set(ctx, key, data, 0).Err()
}

// GetLatest retrieves the latest value for a device
func (r *LatestValueRepository) GetLatest(ctx context.Context, deviceID string) (*domain.LatestValue, error) {
	key := r.getKey(deviceID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest value: %w", err)
	}

	var value domain.LatestValue
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal latest value: %w", err)
	}

	return &value, nil
}

// Close closes the Redis connection
func (r *LatestValueRepository) Close() error {
	return r.client.Close()
}

