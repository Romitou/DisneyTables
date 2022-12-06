package tasks

import (
	"github.com/getsentry/sentry-go"
	"github.com/romitou/disneytables/api"
	"github.com/romitou/disneytables/core"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"github.com/romitou/disneytables/tasker"
	"time"
)

const MaxRequestsPerMinute = 5

func FetchRestaurantSlots() *tasker.Task {
	return &tasker.Task{
		Cron:        "* * * * *",
		Immediately: false,
		Run: func() {
			bookAlerts, err := database.Get().ActiveAlertsToCheck(MaxRequestsPerMinute)
			if err != nil {
				sentry.CaptureException(err)
				return
			}

			timeToWait := 0
			for i := range bookAlerts {
				bookAlert := bookAlerts[i]
				time.AfterFunc(time.Duration(timeToWait)*time.Second, func() {
					restaurantAvailabilities, apiErr := api.RestaurantAvailabilities(api.RestaurantAvailabilitySearch{
						Date:         bookAlert.Date,
						RestaurantID: bookAlert.Restaurant.DisneyID,
						PartyMix:     bookAlert.PartyMix,
					})
					if apiErr != nil {
						sentry.WithScope(func(scope *sentry.Scope) {
							scope.SetExtra("date", bookAlert.Date)
							scope.SetExtra("restaurantId", bookAlert.Restaurant.DisneyID)
							scope.SetExtra("partyMix", bookAlert)
							sentry.CaptureException(apiErr)
						})
						return
					}

					err = database.Get().MarkAlertAsChecked(bookAlert)
					if err != nil {
						sentry.CaptureException(err)
					}

					InsertAvailabilities(restaurantAvailabilities, bookAlert)
				})
				timeToWait += int(time.Minute.Seconds() / MaxRequestsPerMinute)
			}

			errors := core.CreateNotifications()
			if errors != nil {
				for _, err := range errors {
					sentry.CaptureException(err)
				}
			}
			err = core.CleanupActiveNotifications()
			if err != nil {
				sentry.CaptureException(err)
			}
		},
	}
}

func InsertAvailabilities(restaurantAvailabilities []api.RestaurantAvailability, bookAlert models.BookAlert) {
	for _, availability := range restaurantAvailabilities {
		for _, mealPeriod := range availability.MealPeriods {
			for _, slot := range mealPeriod.MealSlots {
				available := slot.Available == "true"
				err := database.Get().UpsertBookSlot(models.BookSlot{
					RestaurantID: bookAlert.Restaurant.ID,
					Date:         availability.Date,
					MealPeriod:   mealPeriod.MealPeriod,
					PartyMix:     bookAlert.PartyMix,
					Available:    &available,
					Hour:         slot.Time,
				})
				if err != nil {
					sentry.WithScope(func(scope *sentry.Scope) {
						scope.SetExtra("date", bookAlert.Date)
						scope.SetExtra("restaurantId", bookAlert.Restaurant.DisneyID)
						scope.SetExtra("partyMix", bookAlert)
						sentry.CaptureException(err)
					})
					continue
				}
			}
		}
	}
}
