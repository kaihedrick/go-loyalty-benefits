package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// KafkaProducer represents a Kafka message producer
type KafkaProducer struct {
	writer *kafka.Writer
	logger *logrus.Logger
}

// KafkaConsumer represents a Kafka message consumer
type KafkaConsumer struct {
	reader *kafka.Reader
	logger *logrus.Logger
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers  []string
	ClientID string
	GroupID  string
	Version  string
}

// Message represents a Kafka message
type Message struct {
	Key       []byte
	Value     []byte
	Topic     string
	Partition int
	Offset    int64
	Timestamp time.Time
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(config *KafkaConfig, logger *logrus.Logger) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Topic:        "", // Set per message
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		Logger:       kafka.LoggerFunc(logger.Debugf),
	}

	return &KafkaProducer{
		writer: writer,
		logger: logger,
	}
}

// Close closes the Kafka producer
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

// SendMessage sends a message to a specific topic
func (p *KafkaProducer) SendMessage(ctx context.Context, topic string, key, value []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message to topic %s: %w", topic, err)
	}

	p.logger.Debugf("Message sent to topic %s with key %s", topic, string(key))
	return nil
}

// SendJSONMessage sends a JSON message to a specific topic
func (p *KafkaProducer) SendJSONMessage(ctx context.Context, topic string, key []byte, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message value: %w", err)
	}

	return p.SendMessage(ctx, topic, key, jsonValue)
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(config *KafkaConfig, topic string, logger *logrus.Logger) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Brokers,
		Topic:    topic,
		GroupID:  config.GroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  1 * time.Second,
		Logger:   kafka.LoggerFunc(logger.Debugf),
	})

	return &KafkaConsumer{
		reader: reader,
		logger: logger,
	}
}

// Close closes the Kafka consumer
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

// ReadMessage reads a message from the topic
func (c *KafkaConsumer) ReadMessage(ctx context.Context) (*Message, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	return &Message{
		Key:       msg.Key,
		Value:     msg.Value,
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Timestamp: msg.Time,
	}, nil
}

// ReadMessageWithTimeout reads a message with a timeout
func (c *KafkaConsumer) ReadMessageWithTimeout(timeout time.Duration) (*Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return c.ReadMessage(ctx)
}

// ConsumeMessages consumes messages from the topic and calls the handler for each message
func (c *KafkaConsumer) ConsumeMessages(ctx context.Context, handler func(*Message) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.ReadMessage(ctx)
			if err != nil {
				c.logger.Errorf("Failed to read message: %v", err)
				continue
			}

			if err := handler(msg); err != nil {
				c.logger.Errorf("Failed to handle message: %v", err)
				// Continue processing other messages
				continue
			}

			c.logger.Debugf("Message consumed from topic %s at offset %d", msg.Topic, msg.Offset)
		}
	}
}

// GetStats returns consumer statistics
func (c *KafkaConsumer) GetStats() kafka.ReaderStats {
	return c.reader.Stats()
}

// SetOffset sets the consumer offset
func (c *KafkaConsumer) SetOffset(offset int64) error {
	return c.reader.SetOffset(offset)
}

// SetOffsetAt sets the consumer offset at a specific time
func (c *KafkaConsumer) SetOffsetAt(ctx context.Context, t time.Time) error {
	return c.reader.SetOffsetAt(ctx, t)
}
