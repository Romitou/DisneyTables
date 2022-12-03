package models

type Restaurant struct {
	ID       uint   `gorm:"primarykey"`
	DisneyID string `gorm:"unique"`
	Name     string
}
