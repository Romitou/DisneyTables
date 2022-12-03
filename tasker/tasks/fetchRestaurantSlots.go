package tasks

import (
	"github.com/romitou/disneytables/api"
	"github.com/romitou/disneytables/core"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"github.com/romitou/disneytables/tasker"
	"log"
)

func FetchRestaurantSlots() *tasker.Task {
	return &tasker.Task{
		Cron:        "*/5 * * * *",
		Immediately: true,
		Run: func() {
			datesToCheck, err := database.Get().RestaurantsToCheck()
			if err != nil {
				log.Println(err)
				return
			}

			for _, dateToCheck := range datesToCheck {
				for _, restaurantToCheck := range dateToCheck.Restaurants {
					restaurantAvailabilities, apiErr := api.RestaurantAvailabilities(api.RestaurantAvailabilitySearch{
						Date:         dateToCheck.Date,
						RestaurantID: restaurantToCheck.Restaurant.DisneyID,
						PartyMix:     restaurantToCheck.LowerPartyMix,
					})
					if apiErr != nil {
						log.Println(apiErr)
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
									PartyMix:     restaurantToCheck.LowerPartyMix,
									Available:    &available,
									Hour:         slot.Time,
								})
								if err != nil {
									log.Println(err)
									continue
								}
							}
						}
					}
				}
			}

			errors := core.CreateNotifications()
			if errors != nil {
				log.Println(errors)
			}
			err = core.CleanupActiveNotifications()
			if err != nil {
				log.Println(err)
			}
		},
	}
}
