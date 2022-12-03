package main

import (
	"github.com/joho/godotenv"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/redis"
	"github.com/romitou/disneytables/tasker"
	"github.com/romitou/disneytables/tasker/tasks"
	"log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	database.Get().Connect()
	redis.Get().Connect()

	tasker.Get().RegisterTasks(
		tasks.SyncRestaurants(),
		tasks.FetchRestaurantSlots(),
		tasks.RenewAuthDetails(),
	)

	go redis.Get().SubscribeBookAlerts()
	tasker.Get().Start()
}
