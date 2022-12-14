package models

import (
	"time"
)

type BookAlert struct {
	ID uint `gorm:"primarykey" json:"id"`

	Restaurant   Restaurant `json:"restaurant"`
	RestaurantID uint       `json:"restaurantId"`

	DiscordID string `json:"discordId"`

	Date       string `json:"date"`
	MealPeriod string `json:"mealPeriod"`
	PartyMix   int    `json:"partyMix"`
	Completed  *bool  `json:"completed"`

	CheckedAt  time.Time `json:"lastChecked"`
	CheckCount int       `json:"checkCount"`
	ErrorCount int       `json:"errorCount"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
