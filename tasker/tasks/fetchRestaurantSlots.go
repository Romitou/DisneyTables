package tasks

import (
	"encoding/json"
	"github.com/getsentry/sentry-go"
	"github.com/romitou/disneytables/api"
	"github.com/romitou/disneytables/core"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"github.com/romitou/disneytables/tasker"
	"log"
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

			rawModifiers := os.Getenv("REQUEST_MODIFIERS")
			if rawModifiers != "" {
				var customModifiers map[string]float64
				err := json.Unmarshal([]byte(rawModifiers), &customModifiers)
				if err != nil {
					log.Printf("Invalid value for REQUEST_MODIFIERS: %s", rawModifiers)
				}

				hour := strconv.Itoa(time.Now().Hour())
				modifier := customModifiers[hour]
				if modifier != 0 {
					oldRequestsPerMinute := maxRequestsPerMinute
					maxRequestsPerMinute = int(float64(maxRequestsPerMinute) * modifier)
					log.Println("Using custom modifier for requests per minute:", modifier)
					log.Println("Requests per minute passing from", oldRequestsPerMinute, "to", maxRequestsPerMinute)
				}
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
