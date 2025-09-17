package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Password string
	JWTToken string `gorm:"column:jwt_token"`
	Posts    []Post
}

