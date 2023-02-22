package webserver

import (
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/romitou/disneytables/api"
	"github.com/romitou/disneytables/database"
	"github.com/romitou/disneytables/database/models"
	"github.com/romitou/disneytables/webserver/middlewares"
	"log"
	"net/http"
)

type CreateBookAlert struct {
	DiscordID          string `json:"discordId"`
	RestaurantDisneyID string `json:"restaurantDisneyId"`
	Date               string `json:"date"`
	MealPeriod         string `json:"mealPeriod"`
	PartyMix           int    `json:"partyMix"`
}

type CompleteBookAlert struct {
	ID uint `json:"id"`
}

func Start() {
	r := gin.Default()

	r.Use(middlewares.Auth())
	r.Use(middlewares.Sentry())

	r.GET("/restaurants", func(c *gin.Context) {
		restaurants, err := database.Get().Restaurants()
		if err != nil {
			sentrygin.GetHubFromContext(c).CaptureException(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, restaurants)
	})

	r.POST("/restaurantAvailabilities", func(c *gin.Context) {
		var search api.RestaurantAvailabilitySearch
		err := c.ShouldBindBodyWith(&search, binding.JSON)
		if err != nil {
			sentrygin.GetHubFromContext(c).CaptureException(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		availabilities, apiErr := api.RestaurantAvailabilities(search)
		if apiErr != nil {
			sentrygin.GetHubFromContext(c).WithScope(func(scope *sentry.Scope) {
				scope.SetExtra("date", search.Date)
				scope.SetExtra("restaurantId", search.RestaurantID)
				scope.SetExtra("partyMix", search.PartyMix)
				scope.SetExtra("rawData", apiErr.RawData)
				sentrygin.GetHubFromContext(c).CaptureException(apiErr.Err)
			})
			c.AbortWithStatus(apiErr.HttpStatusCode)
			return
		}

		c.JSON(http.StatusOK, availabilities)
		return
	})

	r.POST("/bookAlerts", func(c *gin.Context) {
		var alert CreateBookAlert
		err := c.ShouldBindBodyWith(&alert, binding.JSON)
		if err != nil {
			sentrygin.GetHubFromContext(c).CaptureException(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		restaurants, err := database.Get().Restaurants()
		if err != nil {
			sentrygin.GetHubFromContext(c).CaptureException(err)
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
			sentrygin.GetHubFromContext(c).CaptureException(err)
			return
		}

		c.JSON(http.StatusOK, &bookAlert)
		return
	})

	r.POST("/completeBookAlert", func(c *gin.Context) {
		var completeBookAlert CompleteBookAlert
		err := c.ShouldBindBodyWith(&completeBookAlert, binding.JSON)
		if err != nil {
			sentrygin.GetHubFromContext(c).CaptureException(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		bookAlert, err := database.Get().FindBookAlertByID(completeBookAlert.ID)
		if err != nil {
			sentrygin.GetHubFromContext(c).CaptureException(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		err = database.Get().CompleteBookAlert(&bookAlert)
		if err != nil {
			sentrygin.GetHubFromContext(c).CaptureException(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, &bookAlert)
		return
	})

	r.GET("/bookAlerts", func(c *gin.Context) {
		bookAlerts, err := database.Get().ActiveBookAlerts()
		if err != nil {
			sentrygin.GetHubFromContext(c).CaptureException(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, bookAlerts)
	})

	r.GET("/statistics", func(c *gin.Context) {
		statistics, err := database.Get().Statistics()
		if err != nil {
			sentrygin.GetHubFromContext(c).CaptureException(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, statistics)
	})

	r.GET("/dailyReport", func(c *gin.Context) {
		dailyReport, err := database.Get().DailyReport()
		if err != nil {
			sentrygin.GetHubFromContext(c).CaptureException(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, dailyReport)
	})

	log.Println("Starting webserver...")
	err := r.Run()
	if err != nil {
		sentry.CaptureException(err)
	}
}
