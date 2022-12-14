package tasks

import (
	"github.com/getsentry/sentry-go"
	"github.com/romitou/disneytables/api"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"github.com/romitou/disneytables/tasker"
)

func SyncRestaurants() *tasker.Task {
	return &tasker.Task{
		Cron:        "0 0 * * *",
		Immediately: true,
		Run: func() {
			apiRestaurants, err := api.Restaurants()
			if err != nil {
				sentry.CaptureException(err)
				return
			}

			databaseRestaurants, err := database.Get().Restaurants()
			if err != nil {
				sentry.CaptureException(err)
				return
			}

			for _, apiRestaurant := range apiRestaurants {
				if !apiRestaurant.BookingAvailable {
					continue
				}

				var found bool

				for _, databaseRestaurant := range databaseRestaurants {
					if databaseRestaurant.DisneyID == apiRestaurant.DisneyID {
						found = true
						break
					}
				}

				if !found {
					err = database.Get().CreateRestaurant(models.Restaurant{
						DisneyID: apiRestaurant.DisneyID,
						Name:     apiRestaurant.Name,
						ImageURL: apiRestaurant.HeroMediaMobile.URL,
					})
					if err != nil {
						sentry.CaptureException(err)
						continue
					}
				}
			}
		},
	}
}
