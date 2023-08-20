package models

import "time"

type Post struct {
	ID        uint   `gorm:"primarykey"`
	BoardID   string `gorm:"primarykey"`
	CreatedAt time.Time
	Content   string
}
