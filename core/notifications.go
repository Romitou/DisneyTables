package core

import (
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"github.com/romitou/disneytables/redis"
)

func CreateNotifications() []error {
	bookAlerts, err := database.Get().PendingBookAlerts()
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
				err = database.Get().CreateNotification(bookNotification)
				if err != nil {
					errors = append(errors, err)
				}

				err = SendNotification(bookNotification)
				if err != nil {
					errors = append(errors, err)
				}
			}
		}
	}
	return errors
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

func SendNotification(notification models.BookNotification) error {
	redisErr := redis.Get().SendBookNotification(redis.BookNotification{
		BookAlertID:    notification.BookAlert.ID,
		DiscordID:      notification.BookAlert.DiscordID,
		RestaurantName: notification.BookSlot.Restaurant.Name,
		Date:           notification.BookSlot.Date,
		MealPeriod:     notification.BookSlot.MealPeriod,
		PartyMix:       notification.BookSlot.PartyMix,
		Hour:           notification.BookSlot.Hour,
	})
	if redisErr != nil {
		return redisErr
	}
	return nil
}
