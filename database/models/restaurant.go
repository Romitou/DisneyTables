package models

type Restaurant struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	DisneyID string `gorm:"unique" json:"disneyId"`
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl"`
}
