package models

import "time"

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"uniqueIndex;size:64"`
	Email     string `gorm:"uniqueIndex;size:128"`
	ArgonPHC  string `gorm:"size:512"` // PHC-encoded argon2id string
	CreatedAt time.Time
}
