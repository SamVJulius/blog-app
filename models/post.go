package models

import (
	"time"
)

type Post struct {
	ID     uint	`gorm:"primaryKey"`
	Title  string
	Body   string
	Tags   []Tag `gorm:"many2many:post_tags;"`
	CreatedAt time.Time
	UserID uint
}

type Tag struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}
