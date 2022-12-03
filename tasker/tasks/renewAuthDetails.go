package tasks

import (
	"github.com/romitou/disneytables/api"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"github.com/romitou/disneytables/tasker"
	"log"
)

func RenewAuthDetails() *tasker.Task {
	return &tasker.Task{
		Cron:        "0 */6 * * *",
		Immediately: true,
		Run: func() {
			authDetails, err := database.Get().LastAuthDetails()
			if err != nil {
				log.Println(err)
				return
			}

			disneyToken, err := api.RefreshAuth(authDetails.RefreshToken)
			if err != nil {
				log.Println(err)
				return
			}

			err = database.Get().InsertAuthDetails(models.AuthDetails{
				AccessToken:  disneyToken.AccessToken,
				RefreshToken: disneyToken.RefreshToken,
			})
			if err != nil {
				log.Println(err)
				return
			}
		},
	}
}
