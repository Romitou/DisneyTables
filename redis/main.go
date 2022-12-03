package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"log"
	"os"
)

var disneyRedis *DisneyRedis

type DisneyRedis struct {
	RedisClient *redis.Client
}

func Get() *DisneyRedis {
	if disneyRedis == nil {
		disneyRedis = &DisneyRedis{}
	}
	return disneyRedis
}

func (r *DisneyRedis) Connect() {
	r.RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})
}

type BookNotification struct {
	DiscordID      string `json:"discordId"`
	RestaurantName string `json:"restaurantName"`
	Date           string `json:"date"`
	MealPeriod     string `json:"mealPeriod"`
	PartyMix       int    `json:"partyMix"`
	Hour           string `json:"hour"`
}

func (r *DisneyRedis) SendBookNotification(alert BookNotification) error {
	marshal, err := json.Marshal(alert)
	if err != nil {
		return err
	}
	return r.RedisClient.Publish(context.Background(), "book-notifications", string(marshal)).Err()
}

func (r *DisneyRedis) SubscribeBookAlerts() {
	subscribe := r.RedisClient.Subscribe(context.Background(), "book-alerts")
	defer subscribe.Close()
	for {
		message, err := subscribe.ReceiveMessage(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Println(message.Channel, message.Payload)

		var bookAlert models.BookAlert
		err = json.Unmarshal([]byte(message.Payload), &bookAlert)
		if err != nil {
			log.Println(err)
			continue
		}

		err = database.Get().CreateBookAlert(bookAlert)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
