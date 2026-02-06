package account

import (
	"fmt"
	"regexp"
)

// KafkaTopicKind is the unique identifier for the Apache Kafka backend.
// Use this constant when creating CommAddresses for Kafka topics.
const KafkaTopicKind = "kafka-topic"

// MaxKafkaTopicLength is the maximum length for a Kafka topic name.
const MaxKafkaTopicLength = 249

// init registers the Kafka backend at package initialization.
func init() {
	RegisterBackend(&kafkaBackend{})
}

// kafkaTopicPattern validates Kafka topic name format.
// Kafka topics can contain letters, numbers, dots, underscores, and hyphens.
var kafkaTopicPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// kafkaBackend implements the Backend interface for Apache Kafka topics.
type kafkaBackend struct{}

// Kind returns the unique identifier for the Kafka backend.
func (k *kafkaBackend) Kind() string {
	return KafkaTopicKind
}

// ValidateLocator checks if a locator is a valid Kafka topic name.
// Valid names contain only [a-zA-Z0-9._-] and are at most 249 characters.
func (k *kafkaBackend) ValidateLocator(locator string) error {
	if locator == "" {
		return errInvalidAddress("kafkaBackend.ValidateLocator", "empty Kafka topic locator", nil)
	}

	if len(locator) > MaxKafkaTopicLength {
		return errInvalidAddress("kafkaBackend.ValidateLocator",
			fmt.Sprintf("Kafka topic name too long (max %d chars), got %d", MaxKafkaTopicLength, len(locator)), nil)
	}

	if !kafkaTopicPattern.MatchString(locator) {
		return errInvalidAddress("kafkaBackend.ValidateLocator",
			fmt.Sprintf("invalid Kafka topic name, must contain only [a-zA-Z0-9._-], got: %s", locator), nil)
	}

	return nil
}

// ParseLocator validates and returns the locator unchanged.
// Kafka topic names are already in canonical form.
func (k *kafkaBackend) ParseLocator(locator string) (string, error) {
	if err := k.ValidateLocator(locator); err != nil {
		return "", err
	}
	return locator, nil
}

// Metadata returns descriptive information about the Kafka backend.
func (k *kafkaBackend) Metadata() BackendMetadata {
	return BackendMetadata{
		DisplayName:    "Apache Kafka",
		Description:    "High-throughput distributed streaming platform for event-driven architectures",
		LocatorFormat:  "topic-name",
		LocatorExample: "my-events-topic",
		RequiresConfig: true, // Requires Kafka broker configuration
		Version:        "1.0.0",
		Properties: map[string]string{
			"maxPartitions":     "configurable",
			"replicationFactor": "configurable",
		},
	}
}

// Technology returns the TopicTechnology for the Kafka backend.
func (k *kafkaBackend) Technology() TopicTechnology {
	return TopicTechnologyKafka
}

// NewKafkaTopicAddress creates a CommAddress for a Kafka topic.
// The locator is the topic name.
func NewKafkaTopicAddress(locator string) (CommAddress, error) {
	return NewCommAddress(KafkaTopicKind, locator)
}

// ValidateKafkaTopicLocator validates a Kafka topic name format.
//
// This function delegates to the registered Kafka backend for validation.
func ValidateKafkaTopicLocator(locator string) error {
	backend, ok := GetBackend(KafkaTopicKind)
	if !ok {
		return errInvalidAddress("ValidateKafkaTopicLocator", "kafka backend not registered", nil)
	}
	return backend.ValidateLocator(locator)
}
