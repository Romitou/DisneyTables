package models

import (
	"time"
)

type BookAlert struct {
	ID uint `gorm:"primarykey"`

	Restaurant   Restaurant
	RestaurantID uint `json:"restaurantId"`

	DiscordID string `json:"discordId"`

	Date       string `json:"date"`
	MealPeriod string `json:"mealPeriod"`
	PartyMix   int    `json:"partyMix"`
	Status     *bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
