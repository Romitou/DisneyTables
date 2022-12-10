package main

import (
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/redis"
	"github.com/romitou/disneytables/tasker"
	"github.com/romitou/disneytables/tasker/tasks"
	"github.com/romitou/disneytables/webserver"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	})

	database.Get().Connect()
	redis.Get().Connect()

	tasker.Get().RegisterTasks(
		tasks.SyncRestaurants(),
		tasks.FetchRestaurantSlots(),
		tasks.RenewAuthDetails(),
		tasks.CleanupOldBookAlerts(),
	)

	go webserver.Start()
	tasker.Get().Start()
}
