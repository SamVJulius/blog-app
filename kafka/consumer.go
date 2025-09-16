package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"user-jwt/initializers"
	"user-jwt/models"

	"github.com/IBM/sarama"
	"github.com/golang-jwt/jwt/v5"
	"github.com/segmentio/kafka-go"
)

var reader *kafka.Reader

func InitConsumer() {
	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{os.Getenv("KAFKA_BROKER")},
		Topic:    "user-signup",
		GroupID:  "login-consumer-group",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
}

func StopConsumer() {
	if reader != nil {
		err := reader.Close()
		if err != nil {
			log.Println("[Kafka Consumer] Error closing reader:", err)
		}
	}
}

func StartConsumer() {
	log.Println("[Kafka Consumer] Listening for messages...")

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println("[Kafka Consumer] Kafka read error: ", err)
			continue
		}
		log.Printf("[Kafka Consumer] Received message: key=%s value=%s partition=%d offset=%d\n",
			string(msg.Key), string(msg.Value), msg.Partition, msg.Offset)

		email := string(msg.Value)
		log.Println("Begin consuming message for email: ", email)
		handleSignupLogin(email)
		log.Println("Message consumed for email: ", email)

		deleteRecords([]string{os.Getenv("KAFKA_BROKER")}, "user-signup", int32(msg.Partition), msg.Offset)

	}
}

func handleSignupLogin(email string) {
	log.Printf("[Kafka Consumer] Processing login for user: %s\n", email)

	var user models.User
	initializers.DB.First(&user, "email = ?", email)
	if user.ID == 0 {
		log.Println("[Kafka Consumer] User not found in autoLogin:", email)
		return
	}

	var token *jwt.Token

	for token == nil {
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": user.ID,
			"exp": time.Now().Add(time.Hour * 72).Unix(),
		})

		tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			log.Println("[Kafka Consumer] Token generation error:", err)
			return
		}

		// Store token in DB
		user.JWTToken = tokenString
		initializers.DB.Save(&user)
		log.Printf("[Kafka Consumer] Generated and stored token for user: %s\n", email)
		fmt.Printf("[Kafka Consumer] JWT: %s\n", tokenString)
	}

}

// deleteRecords deletes Kafka messages from the beginning of a given partition
// up to (but not including) the specified offset.
//
// Parameters:
//   brokerList []string - list of Kafka bootstrap brokers (e.g., []string{"localhost:9092"}).
//   topic string        - the name of the Kafka topic to operate on.
//   partition int32     - the partition number to delete from.
//   offset int64        - delete all messages with offsets < this value.
//
// How It Works:
//   - Connects to Kafka using Sarama Admin API.
//   - Logs the earliest available offset before deletion.
//   - Issues a DeleteRecords request to shift the log start offset forward.
//   - Logs the earliest offset after deletion to confirm that it moved.
//
// Notes:
//   - Offsets are monotonically increasing. They don't reset after deletion.
//   - This operation is irreversible: once deleted, messages cannot be recovered.
//   - You can only delete from the start of the partition (no selective middle deletion).
func deleteRecords(brokerList []string, topic string, partition int32, offset int64) error {
	log.Printf("[Kafka Admin] Deleting records up to offset %d in partition %d of topic %s\n", offset, partition, topic)

	// 1. Configure Sarama with correct Kafka cluster version
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0 // ⚠️ Change this to match your Kafka version (e.g. sarama.V3_5_0_0 for Kafka 3.5)

	// 2. Create ClusterAdmin to perform admin operations (like DeleteRecords)
	admin, err := sarama.NewClusterAdmin(brokerList, config)
	if err != nil {
		return fmt.Errorf("error creating cluster admin: %v", err)
	}
	defer admin.Close()

	// 3. Create a regular Sarama client to query offsets
	client, err := sarama.NewClient(brokerList, config)
	if err != nil {
		return fmt.Errorf("error creating sarama client: %v", err)
	}
	defer client.Close()

	// 4. Get the oldest offset BEFORE deletion (verification baseline)
	oldestBefore, err := client.GetOffset(topic, partition, sarama.OffsetOldest)
	if err != nil {
		return fmt.Errorf("error getting oldest offset: %v", err)
	}
	log.Printf("[Kafka Admin] Oldest offset before deletion: %d\n", oldestBefore)

	// 5. Prepare map of partition -> offset for deletion
	partitions := map[int32]int64{
		partition: offset,
	}

	// 6. Perform the delete operation
	if err := admin.DeleteRecords(topic, partitions); err != nil {
		return fmt.Errorf("error deleting records: %v", err)
	}

	log.Printf("[Kafka Admin] Successfully issued delete request up to offset %d\n", offset)

	// 7. Get the oldest offset AFTER deletion (verification)
	oldestAfter, err := client.GetOffset(topic, partition, sarama.OffsetOldest)
	if err != nil {
		return fmt.Errorf("error getting oldest offset: %v", err)
	}
	log.Printf("[Kafka Admin] Oldest offset after deletion: %d\n", oldestAfter)

	// 8. Verify that deletion took effect
	if oldestAfter < offset {
		log.Printf("[Kafka Admin] WARNING: Expected oldest offset >= %d, but got %d (deletion may not have fully applied yet)\n", offset, oldestAfter)
	} else {
		log.Printf("[Kafka Admin] Deletion verified: partition now starts at offset %d\n", oldestAfter)
	}

	return nil
}
