package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MessageHandler is a callback function type for handling MQTT messages
type MessageHandler func(deviceID string, value float64, timestamp time.Time) error

// Client wraps the MQTT client
type Client struct {
	client  mqtt.Client
	topic   string
	qos     byte
	handler MessageHandler
}

// Config holds MQTT client configuration
type Config struct {
	BrokerURL string
	ClientID  string
	Username  string
	Password  string
	Topic     string
	QoS       byte
}

// NewClient creates a new MQTT client
func NewClient(cfg Config) (*Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.BrokerURL)
	opts.SetClientID(cfg.ClientID)
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Unexpected message on topic %s: %s", msg.Topic(), string(msg.Payload()))
	})
	opts.OnConnect = func(client mqtt.Client) {
		log.Printf("Connected to MQTT broker: %s", cfg.BrokerURL)
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("Connection lost: %v", err)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	return &Client{
		client: client,
		topic:  cfg.Topic,
		qos:    cfg.QoS,
	}, nil
}

// SetMessageHandler sets the message handler callback
func (c *Client) SetMessageHandler(handler MessageHandler) {
	c.handler = handler
}

// Subscribe subscribes to the configured topic
func (c *Client) Subscribe() error {
	if c.handler == nil {
		return fmt.Errorf("message handler not set")
	}

	token := c.client.Subscribe(c.topic, c.qos, c.messageHandler)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", c.topic, token.Error())
	}

	log.Printf("Subscribed to topic: %s", c.topic)
	return nil
}

// messageHandler handles incoming MQTT messages
func (c *Client) messageHandler(client mqtt.Client, msg mqtt.Message) {
	var payload struct {
		Device struct {
			ID string `json:"id"`
		} `json:"device"`
		Data struct {
			Value     float64 `json:"value"`
			Timestamp string  `json:"timestamp"`
		} `json:"data"`
	}

	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		log.Printf("Failed to parse message from topic %s: %v", msg.Topic(), err)
		return
	}

	if payload.Device.ID == "" {
		log.Printf("Device ID is empty in message from topic %s", msg.Topic())
		return
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, payload.Data.Timestamp)
	if err != nil {
		// If parsing fails, use current time
		timestamp = time.Now()
		log.Printf("Failed to parse timestamp, using current time: %v", err)
	}

	// Call the handler
	if err := c.handler(payload.Device.ID, payload.Data.Value, timestamp); err != nil {
		log.Printf("Handler error for device %s: %v", payload.Device.ID, err)
	}
}

// Disconnect disconnects from the MQTT broker
func (c *Client) Disconnect(quiesce uint) {
	c.client.Disconnect(quiesce)
	log.Println("Disconnected from MQTT broker")
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

