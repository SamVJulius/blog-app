package initializers

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

var DB *gorm.DB

func ConnectDB() {
	var err error
	dsn := "host=localhost user=" + os.Getenv("DB_USERNAME") + " password=" + os.Getenv("DB_PASSWORD") + " dbname=userjwt port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database!")
	}
}