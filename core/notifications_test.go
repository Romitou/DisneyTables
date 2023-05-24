package core

import (
	"reflect"
	"testing"

	"github.com/romitou/disneytables/database/models"
)

func TestGenerateNotifications_Empty(t *testing.T) {
	result := GenerateNotifications(nil)
	if len(result) != 0 {
		t.Fatalf("expected 0 notifications, got %d", len(result))
	}
}

func TestGenerateNotifications_BatchingHours(t *testing.T) {
	alert := models.BookAlert{
		ID:        1,
		DiscordID: "user-123456",
	}
	restaurant := models.Restaurant{
		ID:       10,
		DisneyID: "disney-rest-1",
		Name:     "Bistrot Chez Rémy",
	}

	input := []*models.BookNotification{
		{
			BookAlert: alert,
			BookSlot: models.BookSlot{
				Restaurant: restaurant,
				Date:       "2023-06-15",
				MealPeriod: "Dinner",
				PartyMix:   2,
				Hour:       "18:30",
			},
		},
		{
			BookAlert: alert,
			BookSlot: models.BookSlot{
				Restaurant: restaurant,
				Date:       "2023-06-15",
				MealPeriod: "Dinner",
				PartyMix:   2,
				Hour:       "19:00",
			},
		},
	}

	notifications := GenerateNotifications(input)
	if len(notifications) != 1 {
		t.Fatalf("expected 1 batched notification, got %d", len(notifications))
	}

	expectedHours := []string{"18:30", "19:00"}
	if !reflect.DeepEqual(notifications[0].Hours, expectedHours) {
		t.Fatalf("expected hours %v, got %v", expectedHours, notifications[0].Hours)
	}

	if notifications[0].BookAlertID != 1 {
		t.Fatalf("expected alert ID 1, got %d", notifications[0].BookAlertID)
	}
	if notifications[0].DiscordID != "user-123456" {
		t.Fatalf("expected discordId 'user-123456', got %s", notifications[0].DiscordID)
	}
}

func TestGenerateNotifications_MultipleAlerts(t *testing.T) {
	alert1 := models.BookAlert{ID: 1, DiscordID: "user-1"}
	alert2 := models.BookAlert{ID: 2, DiscordID: "user-2"}
	restaurant := models.Restaurant{ID: 10, DisneyID: "rest-1", Name: "Captain Jack's"}

	input := []*models.BookNotification{
		{
			BookAlert: alert1,
			BookSlot: models.BookSlot{
				Restaurant: restaurant,
				Date:       "2023-07-01",
				MealPeriod: "Lunch",
				PartyMix:   4,
				Hour:       "12:00",
			},
		},
		{
			BookAlert: alert2,
			BookSlot: models.BookSlot{
				Restaurant: restaurant,
				Date:       "2023-07-01",
				MealPeriod: "Lunch",
				PartyMix:   2,
				Hour:       "12:30",
			},
		},
		{
			BookAlert: alert1,
			BookSlot: models.BookSlot{
				Restaurant: restaurant,
				Date:       "2023-07-01",
				MealPeriod: "Lunch",
				PartyMix:   4,
				Hour:       "13:00",
			},
		},
	}

	notifications := GenerateNotifications(input)
	if len(notifications) != 2 {
		t.Fatalf("expected 2 notifications for distinct alerts, got %d", len(notifications))
	}

	if len(notifications[0].Hours) != 2 {
		t.Fatalf("expected 2 batched hours for alert 1, got %d", len(notifications[0].Hours))
	}
	if len(notifications[1].Hours) != 1 {
		t.Fatalf("expected 1 hour for alert 2, got %d", len(notifications[1].Hours))
	}
}
