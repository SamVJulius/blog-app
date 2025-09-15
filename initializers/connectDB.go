package initializers

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var err error
	dsn := "host=" + os.Getenv("DB_HOST") + " user=" + os.Getenv("DB_USERNAME") + " password=" + os.Getenv("DB_PASSWORD") + " dbname=userjwt port=5432 sslmode=disable TimeZone=UTC"
	
	for i := 0; i < 10; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("✅ Database connection established")
			return
		} else {
			log.Printf("❌ Database connection failed: %v\n", err)
		}

		time.Sleep(2 * time.Second)
	}	
	if err != nil {
		log.Fatal("❌ Could not connect to the database after several attempts:", err)
	}	
}