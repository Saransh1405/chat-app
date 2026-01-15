package kafka

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// Kafka wraps IBM Sarama producer for message publishing
type Kafka struct {
	Producer sarama.SyncProducer
	Brokers  []string
}

// NewKafka creates a new Kafka producer instance
// brokers: list of Kafka broker addresses (e.g., []string{"localhost:9092"})
func NewKafka(brokers []string) (*Kafka, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("at least one Kafka broker address is required")
	}

	// Configure producer
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Retry.Backoff = 100

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &Kafka{
		Producer: producer,
		Brokers:  brokers,
	}, nil
}

// Produce sends a message to the specified Kafka topic
// topic: the Kafka topic name
// message: the message payload as byte array
// Returns error if message production fails
func (k *Kafka) Produce(topic string, message []byte) error {
	if k.Producer == nil {
		return fmt.Errorf("Kafka producer is not initialized")
	}

	if topic == "" {
		return fmt.Errorf("topic name cannot be empty")
	}

	// Create producer message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	// Send message synchronously
	partition, offset, err := k.Producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to topic %s: %w", topic, err)
	}

	log.Printf("Message sent to topic %s, partition %d, offset %d", topic, partition, offset)
	return nil
}

// ProduceWithKey sends a message to the specified Kafka topic with a partition key
// topic: the Kafka topic name
// key: the partition key (messages with same key go to same partition)
// message: the message payload as byte array
// Returns error if message production fails
func (k *Kafka) ProduceWithKey(topic string, key string, message []byte) error {
	if k.Producer == nil {
		return fmt.Errorf("Kafka producer is not initialized")
	}

	if topic == "" {
		return fmt.Errorf("topic name cannot be empty")
	}

	// Create producer message with key
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(message),
	}

	// Send message synchronously
	partition, offset, err := k.Producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to topic %s with key %s: %w", topic, key, err)
	}

	log.Printf("Message sent to topic %s, partition %d, offset %d, key: %s", topic, partition, offset, key)
	return nil
}

// Close closes the Kafka producer connection
// Should be called during application shutdown
func (k *Kafka) Close() error {
	if k.Producer != nil {
		return k.Producer.Close()
	}
	return nil
}
