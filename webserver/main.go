package webserver

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/romitou/disneytables/api"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"log"
	"net/http"
	"os"
	"strings"
)

type CreateBookAlert struct {
	DiscordID          string `json:"discordId"`
	RestaurantDisneyID string `json:"restaurantDisneyId"`
	Date               string `json:"date"`
	MealPeriod         string `json:"mealPeriod"`
	PartyMix           int    `json:"partyMix"`
}

func Start() {
	r := gin.Default()

	r.Use(func(context *gin.Context) {
		authorization := context.GetHeader("Authorization")
		if authorization == "" {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if strings.TrimPrefix(authorization, "Bearer ") != os.Getenv("WEBSERVER_TOKEN") {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		context.Next()
	})

	r.GET("/restaurants", func(c *gin.Context) {
		restaurants, err := database.Get().Restaurants()
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, restaurants)
	})

	r.POST("/restaurantAvailabilities", func(c *gin.Context) {
		var search api.RestaurantAvailabilitySearch
		err := c.ShouldBindBodyWith(&search, binding.JSON)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		availabilities, err := api.RestaurantAvailabilities(search)
		if err != nil {
			log.Println(err)
			return
		}

		c.JSON(http.StatusOK, availabilities)
		return
	})

	r.POST("/bookAlerts", func(c *gin.Context) {
		var alert CreateBookAlert
		err := c.ShouldBindBodyWith(&alert, binding.JSON)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		restaurants, err := database.Get().Restaurants()
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var foundRestaurant models.Restaurant
		for _, restaurant := range restaurants {
			if restaurant.DisneyID == alert.RestaurantDisneyID {
				foundRestaurant = restaurant
				break
			}
		}

		completed := false
		bookAlert := models.BookAlert{
			DiscordID:  alert.DiscordID,
			Restaurant: foundRestaurant,
			Date:       alert.Date,
			MealPeriod: alert.MealPeriod,
			PartyMix:   alert.PartyMix,
			Completed:  &completed,
		}

		err = database.Get().CreateBookAlert(&bookAlert)
		if err != nil {
			log.Println(err)
			return
		}

		c.JSON(http.StatusOK, &bookAlert)
		return
	})

	r.GET("/bookAlerts", func(c *gin.Context) {
		bookAlerts, err := database.Get().PendingBookAlerts()
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, bookAlerts)
	})

	log.Println("Starting webserver...")
	err := r.Run("0.0.0.0:8080")
	if err != nil {
		log.Fatal(err)
	}
}
