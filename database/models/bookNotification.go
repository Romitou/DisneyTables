package models

import "time"

type BookNotification struct {
	ID uint `gorm:"primarykey"`

	BookAlert   BookAlert
	BookAlertID uint

	BookSlot   BookSlot
	BookSlotID uint

	Active *bool

	CreatedAt time.Time
}
