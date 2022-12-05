package tasks

import (
	"github.com/getsentry/sentry-go"
	"github.com/romitou/disneytables/api"
	"github.com/romitou/disneytables/core"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"github.com/romitou/disneytables/tasker"
)

func FetchRestaurantSlots() *tasker.Task {
	return &tasker.Task{
		Cron:        "*/5 * * * *",
		Immediately: true,
		Run: func() {
			datesToCheck, err := database.Get().RestaurantsToCheck()
			if err != nil {
				sentry.CaptureException(err)
				return
			}

			for _, dateToCheck := range datesToCheck {
				for _, restaurantToCheck := range dateToCheck.Restaurants {
					for _, partyMix := range restaurantToCheck.PartyMixes {
						restaurantAvailabilities, apiErr := api.RestaurantAvailabilities(api.RestaurantAvailabilitySearch{
							Date:         dateToCheck.Date,
							RestaurantID: restaurantToCheck.Restaurant.DisneyID,
							PartyMix:     partyMix,
						})
						if apiErr != nil {
							sentry.WithScope(func(scope *sentry.Scope) {
								scope.SetExtra("date", dateToCheck.Date)
								scope.SetExtra("restaurantId", restaurantToCheck.Restaurant.DisneyID)
								scope.SetExtra("partyMix", partyMix)
								sentry.CaptureException(apiErr)
							})
							continue
						}

						for _, availability := range restaurantAvailabilities {
							for _, mealPeriod := range availability.MealPeriods {
								for _, slot := range mealPeriod.MealSlots {
									available := slot.Available == "true"
									err = database.Get().UpsertBookSlot(models.BookSlot{
										RestaurantID: restaurantToCheck.Restaurant.ID,
										Date:         dateToCheck.Date,
										MealPeriod:   mealPeriod.MealPeriod,
										PartyMix:     partyMix,
										Available:    &available,
										Hour:         slot.Time,
									})
									if err != nil {
										sentry.CaptureException(err)
										continue
									}
								}
							}
						}
					}
				}
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
