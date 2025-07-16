package messaging

import (
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var producer *kafka.Producer

func Init(brokers string) {
	var err error
	producer, err = kafka.NewProducer(&kafka.ConfigMap{
		bootstrapServersKey: brokers,
	})
	if err != nil {
		log.Fatalf("Kafka Producer init error: %v", err)
	} else {
		log.Printf("Kafka Producer init")
	}
}

func Send(topic string, data []byte) error {
	deliveryChan := make(chan kafka.Event)

	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: data,
	}, deliveryChan)

	if err != nil {
		return err
	}

	e := <-deliveryChan
	close(deliveryChan)
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}

	return nil
}
