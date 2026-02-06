package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// Kafka wraps IBM Sarama producer and consumer for message publishing & consuming
type Kafka struct {
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
	Brokers  []string
}

// NewKafka creates a new Kafka producer and consumer instance
// brokers: list of Kafka broker addresses (e.g., []string{"localhost:9092"})
func NewKafka(brokers []string) (*Kafka, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("at least one Kafka broker address is required")
	}

	config := sarama.NewConfig()
	// Producer configuration
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Retry.Backoff = 100
	// Consumer configuration
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		producer.Close() // Clean up producer if consumer fails
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return &Kafka{
		Producer: producer,
		Consumer: consumer,
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

// StartConsumer consumes messages from a specific topic and partition
func (k *Kafka) StartConsumer(topic string, partition int32, offset int64) error {
	ctx := context.Background()

	if k.Consumer == nil {
		return fmt.Errorf("Kafka consumer is not initialized")
	}

	partitionConsumer, err := k.Consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		log.Printf("Failed to create partition consumer: %v", err)
		return err
	}
	defer partitionConsumer.Close()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			log.Printf("Message received: %v", msg)
		case err := <-partitionConsumer.Errors():
			log.Printf("Error from partition consumer: %v", err)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (k *Kafka) Close() error {
	var firstErr error
	if k.Producer != nil {
		if err := k.Producer.Close(); err != nil {
			firstErr = err
		}
	}
	if k.Consumer != nil {
		if err := k.Consumer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
