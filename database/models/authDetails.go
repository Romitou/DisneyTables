package models

import "time"

type AuthDetails struct {
	ID           uint `gorm:"primarykey"`
	AccessToken  string
	RefreshToken string
	CreatedAt    time.Time
}
