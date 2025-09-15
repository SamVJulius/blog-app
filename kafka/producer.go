package kafka

import (
	"context"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

func InitProducer() {
	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{os.Getenv("KAFKA_BROKER")},
		Topic:    "user-signup",
		Balancer: &kafka.LeastBytes{},
	})
	log.Println("[Kafka Producer] Initialized and ready to send messages.")
}

func ProduceSignupEvent(email string) error {
	log.Printf("[Kafka Producer] Producing signup event for email: %s\n", email)

	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(email),
			Value: []byte(email),
		},
	)

	if err != nil {
		log.Println("[Kafka Producer] Error producing message: ", err)
	} else {
		log.Printf("[Kafka Producer] Successfully produced message for email: %s\n", email)
	}

	return err
}

func CloseProducer() {
	if writer != nil {
		writer.Close()
		log.Println("[Kafka Producer] Closed writer.")
	}
}
