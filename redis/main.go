package redis

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v9"
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
	BookAlertID    uint   `json:"bookAlertId"`
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
