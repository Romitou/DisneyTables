package tasks

import (
	"github.com/getsentry/sentry-go"
	"github.com/romitou/disneytables/api"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"github.com/romitou/disneytables/tasker"
)

func RenewAuthDetails() *tasker.Task {
	return &tasker.Task{
		Cron:        "0 */6 * * *",
		Immediately: true,
		Run: func() {
			authDetails, err := database.Get().LastAuthDetails()
			if err != nil {
				sentry.CaptureException(err)
				return
			}

			disneyToken, err := api.RefreshAuth(authDetails.RefreshToken)
			if err != nil {
				sentry.CaptureException(err)
				return
			}

			err = database.Get().InsertAuthDetails(models.AuthDetails{
				AccessToken:  disneyToken.AccessToken,
				RefreshToken: disneyToken.RefreshToken,
			})
			if err != nil {
				sentry.CaptureException(err)
				return
			}
		},
	}
}
