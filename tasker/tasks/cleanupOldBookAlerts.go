package tasks

import (
	"github.com/getsentry/sentry-go"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/tasker"
	"time"
)

func CleanupOldBookAlerts() *tasker.Task {
	return &tasker.Task{
		Cron:        "0 0 * * *",
		Immediately: true,
		Run: func() {
			bookAlerts, err := database.Get().ActiveBookAlerts()
			if err != nil {
				return
			}
			for _, bookAlert := range bookAlerts {
				date, parseErr := time.Parse("2006-01-02", bookAlert.Date)
				if parseErr != nil {
					sentry.CaptureException(parseErr)
					continue
				}
				oneDayBehind := time.Now().AddDate(0, 0, -1)
				if date.Before(oneDayBehind) {
					err = database.Get().CompleteBookAlert(&bookAlert)
					if err != nil {
						sentry.CaptureException(err)
						continue
					}
				}
			}
		},
	}
}
