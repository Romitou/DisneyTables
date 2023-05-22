package tasks

import (
	"github.com/getsentry/sentry-go"
	"github.com/romitou/disneytables/api"
	"github.com/romitou/disneytables/core"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"github.com/romitou/disneytables/tasker"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

const DefaultMaxRequestsPerMinute = 5

func FetchRestaurantSlots() *tasker.Task {
	return &tasker.Task{
		Cron:        "* * * * *",
		Immediately: false,
		Run: func() {
			maxRequestsPerMinute := DefaultMaxRequestsPerMinute

			// Try to check if a custom value is set for the maximum number of requests per minute
			rawEnv := os.Getenv("MAX_REQUESTS_PER_MINUTE")
			if rawEnv != "" {
				parsedMaxRequest, err := strconv.Atoi(rawEnv)
				if err != nil {
					log.Printf("Invalid value for MAX_REQUESTS_PER_MINUTE: %s", rawEnv)
				} else {
					maxRequestsPerMinute = parsedMaxRequest
				}
			}

			// Eco mode modifier
			// This eco mode modifier is used to reduce the number of requests, especially at night. All the tables
			// are for the most part booked during the day, so we can reduce the number of requests at night.
			hour := time.Now().Hour()
			if hour <= 8 {
				rawModifier := 1 - ((-1 / 16) * hour * (hour - 8))
				modifier := float64(rawModifier) - 0.2
				if modifier > 1 {
					modifier = 1
				}

				maxRequestsPerMinute = maxRequestsPerMinute * int(math.Round(modifier))
			}

			bookAlerts, err := database.Get().ActiveAlertsToCheck(maxRequestsPerMinute)
			if err != nil {
				sentry.CaptureException(err)
				return
			}

			log.Println("Checking", len(bookAlerts), "alerts...")

			timeToWait := 0
			for i := range bookAlerts {
				bookAlert := bookAlerts[i]
				time.AfterFunc(time.Duration(timeToWait)*time.Second, func() {
					log.Println("Checking alert #", bookAlert.ID, " for ", bookAlert.Restaurant.Name, " on ", bookAlert.Date, " for ", bookAlert.PartyMix, " peoples for ", bookAlert.MealPeriod)
					restaurantAvailabilities, apiErr := api.RestaurantAvailabilities(api.RestaurantAvailabilitySearch{
						Date:         bookAlert.Date,
						RestaurantID: bookAlert.Restaurant.DisneyID,
						PartyMix:     bookAlert.PartyMix,
					})
					if apiErr != nil {
						sentry.WithScope(func(scope *sentry.Scope) {
							scope.SetExtra("date", bookAlert.Date)
							scope.SetExtra("restaurantId", bookAlert.Restaurant.DisneyID)
							scope.SetExtra("partyMix", bookAlert.PartyMix)
							scope.SetExtra("rawData", apiErr.RawData)
							sentry.CaptureException(apiErr.Err)
						})
						err = database.Get().MarkAlertAsErrored(bookAlert)
						if err != nil {
							sentry.CaptureException(err)
						}
						return
					}

					err = database.Get().MarkAlertAsChecked(bookAlert)
					if err != nil {
						sentry.CaptureException(err)
					}

					InsertAvailabilities(restaurantAvailabilities, bookAlert)
				})
				timeToWait += int(time.Minute.Seconds() / float64(maxRequestsPerMinute))
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
