package redis

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v9"
	"github.com/romitou/disneytables/database/models"
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

type Notification struct {
	BookAlertID uint              `json:"bookAlertId"`
	DiscordID   string            `json:"discordId"`
	Restaurant  models.Restaurant `json:"restaurant"`
	Date        string            `json:"date"`
	MealPeriod  string            `json:"mealPeriod"`
	PartyMix    int               `json:"partyMix"`
	Hours       []string          `json:"hours"`
}

func (r *DisneyRedis) SendBookNotification(notification Notification) error {
	marshal, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	return r.RedisClient.Publish(context.Background(), "book-notifications", string(marshal)).Err()
}
