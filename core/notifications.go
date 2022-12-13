package core

import (
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"github.com/romitou/disneytables/redis"
)

func CreateNotifications() []error {
	bookAlerts, err := database.Get().ActiveBookAlerts()
	if err != nil {
		return []error{err}
	}

	var errors []error
	for _, bookAlert := range bookAlerts {
		bookSlots, apiErr := database.Get().FindAvailableSlotsForAlert(bookAlert)
		if apiErr != nil {
			errors = append(errors, apiErr)
			continue
		}

		var bookNotifications []*models.BookNotification
		for _, bookSlot := range bookSlots {
			exists, existsErr := database.Get().NotificationExists(bookAlert, bookSlot)
			if existsErr != nil {
				errors = append(errors, existsErr)
				exists = true
			}

			if !exists {
				active := true
				bookNotification := models.BookNotification{
					BookAlert: bookAlert,
					BookSlot:  bookSlot,
					Active:    &active,
				}
				err = database.Get().CreateNotification(&bookNotification)
				if err != nil {
					errors = append(errors, err)
				}
				bookNotifications = append(bookNotifications, &bookNotification)
			}
		}

		redisNotifications := GenerateNotifications(bookNotifications)
		for _, redisNotification := range redisNotifications {
			redisErr := redis.Get().SendBookNotification(*redisNotification)
			if redisErr != nil {
				errors = append(errors, redisErr)
			}
		}
	}
	return errors
}

func GenerateNotifications(bookNotifications []*models.BookNotification) []*redis.Notification {
	var notifications []*redis.Notification
	for _, bookNotification := range bookNotifications {
		found := false
		for _, notification := range notifications {
			if notification.BookAlertID == bookNotification.BookAlert.ID {
				found = true
				notification.Hours = append(notification.Hours, bookNotification.BookSlot.Hour)
			}
		}
		if !found {
			notifications = append(notifications, &redis.Notification{
				BookAlertID: bookNotification.BookAlert.ID,
				DiscordID:   bookNotification.BookAlert.DiscordID,
				Restaurant:  bookNotification.BookSlot.Restaurant,
				Date:        bookNotification.BookSlot.Date,
				MealPeriod:  bookNotification.BookSlot.MealPeriod,
				PartyMix:    bookNotification.BookSlot.PartyMix,
				Hours:       []string{bookNotification.BookSlot.Hour},
			})
		}
	}
	return notifications
}

func CleanupActiveNotifications() error {
	activeNotifications, err := database.Get().ActiveNotifications()
	if err != nil {
		return err
	}

	for _, activeNotification := range activeNotifications {
		if !*activeNotification.BookSlot.Available {
			err = database.Get().DeactivateNotification(activeNotification)
			if err != nil {
				return err
			}
		}
	}

	return err
}
