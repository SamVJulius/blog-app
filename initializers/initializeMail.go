package initializers

import (
	"net/smtp"
	"os"
)

var Auth smtp.Auth

func InitializeMail() {
	Auth = smtp.PlainAuth("", "samsonvjulius@gmail.com", os.Getenv("EMAIL_PASSWORD"), "smtp.gmail.com")
}