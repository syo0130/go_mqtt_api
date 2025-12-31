package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go_mqtt_api/config"
	"go_mqtt_api/internal/infrastructure/mqtt"
	postgresRepo "go_mqtt_api/internal/infrastructure/postgres"
	redisRepo "go_mqtt_api/internal/infrastructure/redis"
	httpHandler "go_mqtt_api/internal/interfaces/http"
	"go_mqtt_api/internal/usecase"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize repositories
	postgresRepository, err := postgresRepo.NewSensorDataRepository(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL repository: %v", err)
	}
	defer postgresRepository.Close()

	redisRepository, err := redisRepo.NewLatestValueRepository(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to initialize Redis repository: %v", err)
	}
	defer redisRepository.Close()

	// Initialize usecase
	service := usecase.NewSensorDataService(postgresRepository, redisRepository)

	// Initialize MQTT client
	mqttConfig := mqtt.Config{
		BrokerURL: cfg.MQTTBrokerURL,
		ClientID:  cfg.MQTTClientID,
		Username:  cfg.MQTTUsername,
		Password:  cfg.MQTTPassword,
		Topic:     cfg.MQTTTopic,
		QoS:       cfg.MQTTQoS,
	}

	mqttClient, err := mqtt.NewClient(mqttConfig)
	if err != nil {
		log.Fatalf("Failed to initialize MQTT client: %v", err)
	}
	defer mqttClient.Disconnect(250)

	// Set MQTT message handler
	mqttClient.SetMessageHandler(func(deviceID string, value float64, timestamp time.Time) error {
		ctx := context.Background()
		return service.ProcessSensorData(ctx, deviceID, value, timestamp)
	})

	// Subscribe to MQTT topic
	if err := mqttClient.Subscribe(); err != nil {
		log.Fatalf("Failed to subscribe to MQTT topic: %v", err)
	}

	// Initialize HTTP server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Register routes
	handler := httpHandler.NewHandler(service)
	handler.RegisterRoutes(e)

	// Start HTTP server in a goroutine
	serverAddr := fmt.Sprintf(":%s", cfg.HTTPPort)
	go func() {
		log.Printf("Starting HTTP server on %s", serverAddr)
		if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}

	// Disconnect MQTT client
	mqttClient.Disconnect(250)

	log.Println("Server shutdown complete")
}
