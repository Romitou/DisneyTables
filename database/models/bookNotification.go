package models

import "time"

type BookNotification struct {
	ID uint `gorm:"primarykey"`

	BookAlert   BookAlert `json:"bookAlert"`
	BookAlertID uint      `json:"bookAlertId"`

	BookSlot   BookSlot `json:"bookSlot"`
	BookSlotID uint     `json:"bookSlotId"`

	Active *bool `json:"active"`

	CreatedAt time.Time `json:"createdAt"`
}
