package initializers

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnvVariables() {
	// Detect if running inside Docker by checking cgroup file or env var
	if _, err := os.Stat("/.dockerenv"); err == nil {
		log.Println("🐳 Running inside Docker — skipping .env file load")
		return
	}

	// Local dev: try loading .env
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Printf("⚠️ Could not load .env file: %v\n", err)
		} else {
			log.Println("✅ .env file loaded successfully")
		}
	} else {
		log.Println("⚠️ No .env file found locally — relying on system environment variables")
	}
}
