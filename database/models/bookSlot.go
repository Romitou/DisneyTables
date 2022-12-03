package models

import "time"

type BookSlot struct {
	ID           uint `gorm:"primarykey"`
	Restaurant   Restaurant
	RestaurantID uint

	Date       string
	MealPeriod string
	PartyMix   int
	Hour       string

	WasAvailable *bool
	Available    *bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
