package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"user-jwt/initializers"
	"user-jwt/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/segmentio/kafka-go"
)

var reader *kafka.Reader

func InitConsumer() {
	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER")},
		Topic:   "user-signup",
		GroupID: "login-consumer-group",
		MinBytes: 1,
		MaxBytes: 10e6, 		
	})
	log.Println("[Kafka Consumer] Initialized and ready to consume messages.")
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
		handleSignupLogin(email)

		if err := reader.CommitMessages(context.Background(), msg); err != nil {
			log.Printf("[Kafka Consumer] ❌ Failed to commit message offset %d: %v", msg.Offset, err)
		} else {
			log.Printf("[Kafka Consumer] ✅ Committed offset %d", msg.Offset)
		}
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
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
