package database

import (
	"errors"
	"github.com/romitou/disneytables/database/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

var disneyDatabase *DisneyDatabase

type DisneyDatabase struct {
	gorm *gorm.DB
}

func Get() *DisneyDatabase {
	if disneyDatabase == nil {
		disneyDatabase = &DisneyDatabase{}
	}
	return disneyDatabase
}

func (d *DisneyDatabase) Connect() {
	loggerMode := logger.Error
	if os.Getenv("DEBUG_MODE") == "true" {
		loggerMode = logger.Info
	}

	database, err := gorm.Open(mysql.Open(os.Getenv("MYSQL_DSN")), &gorm.Config{
		Logger: logger.Default.LogMode(loggerMode),
	})
	if err != nil {
		log.Fatal(err)
	}

	err = database.AutoMigrate(&models.BookAlert{}, &models.AuthDetails{}, &models.Restaurant{}, &models.BookSlot{}, &models.BookNotification{})
	if err != nil {
		log.Println(err)
	}

	d.gorm = database
}

func (d *DisneyDatabase) Restaurants() ([]models.Restaurant, error) {
	var restaurants []models.Restaurant
	err := d.gorm.Find(&restaurants).Error
	return restaurants, err
}

func (d *DisneyDatabase) CreateRestaurant(restaurant models.Restaurant) error {
	log.Println("Creating restaurant: ", restaurant)
	return d.gorm.Create(&restaurant).Error
}

func (d *DisneyDatabase) FirstAuthDetails() (models.AuthDetails, error) {
	var authDetails models.AuthDetails
	err := d.gorm.First(&authDetails).Error
	return authDetails, err
}

func (d *DisneyDatabase) PendingBookAlerts() ([]models.BookAlert, error) {
	var bookAlerts []models.BookAlert
	f := false
	err := d.gorm.Where(models.BookAlert{
		Status: &f,
	}).Preload("Restaurant").Find(&bookAlerts).Error
	return bookAlerts, err
}

type DateToCheck struct {
	Date        string
	Restaurants []RestaurantToCheck
}

type RestaurantToCheck struct {
	Restaurant    models.Restaurant
	LowerPartyMix int
}

func (d *DisneyDatabase) RestaurantsToCheck() ([]DateToCheck, error) {
	bookAlerts, err := d.PendingBookAlerts()
	if err != nil {
		return nil, err
	}

	var datesToCheck []DateToCheck
	for _, bookAlert := range bookAlerts {
		var found bool
		for i, dateToCheck := range datesToCheck {
			if dateToCheck.Date == bookAlert.Date {
				found = true
				var foundRestaurant bool
				for j, restaurantToCheck := range dateToCheck.Restaurants {
					if restaurantToCheck.Restaurant.ID == bookAlert.Restaurant.ID {
						foundRestaurant = true
						if restaurantToCheck.LowerPartyMix > bookAlert.PartyMix {
							datesToCheck[i].Restaurants[j].LowerPartyMix = bookAlert.PartyMix
						}
						break
					}
				}
				if !foundRestaurant {
					datesToCheck[i].Restaurants = append(datesToCheck[i].Restaurants, RestaurantToCheck{
						Restaurant:    bookAlert.Restaurant,
						LowerPartyMix: bookAlert.PartyMix,
					})
				}
				break
			}
		}
		if !found {
			datesToCheck = append(datesToCheck, DateToCheck{
				Date: bookAlert.Date,
				Restaurants: []RestaurantToCheck{
					{
						Restaurant:    bookAlert.Restaurant,
						LowerPartyMix: bookAlert.PartyMix,
					},
				},
			})
		}
	}

	return datesToCheck, nil
}

func (d *DisneyDatabase) CreateBookAlert(bookAlert models.BookAlert) error {
	return d.gorm.Create(&bookAlert).Error
}

func (d *DisneyDatabase) UpsertBookSlot(bookSlot models.BookSlot) error {
	var existingBookSlot models.BookSlot
	err := d.gorm.Where(models.BookSlot{
		RestaurantID: bookSlot.RestaurantID,
		Date:         bookSlot.Date,
		MealPeriod:   bookSlot.MealPeriod,
		PartyMix:     bookSlot.PartyMix,
		Hour:         bookSlot.Hour,
	}).First(&existingBookSlot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		f := false
		bookSlot.WasAvailable = &f
		return d.gorm.Create(&bookSlot).Error
	}
	if err != nil {
		return err
	}

	existingBookSlot.WasAvailable = existingBookSlot.Available
	existingBookSlot.Available = bookSlot.Available

	return d.gorm.Save(&existingBookSlot).Error
}

func (d *DisneyDatabase) InsertAuthDetails(authDetails models.AuthDetails) error {
	return d.gorm.Create(&authDetails).Error
}

func (d *DisneyDatabase) FindAvailableSlotsForAlert(alert models.BookAlert) ([]models.BookSlot, error) {
	var bookSlots []models.BookSlot
	available := true
	err := d.gorm.Where(models.BookSlot{
		RestaurantID: alert.RestaurantID,
		Date:         alert.Date,
		MealPeriod:   alert.MealPeriod,
		PartyMix:     alert.PartyMix,
		Available:    &available,
	}).Preload("Restaurant").Find(&bookSlots).Error
	return bookSlots, err
}

func (d *DisneyDatabase) CreateNotification(notification models.BookNotification) error {
	return d.gorm.Create(&notification).Error
}

func (d *DisneyDatabase) NotificationExists(alert models.BookAlert, bookSlot models.BookSlot) (bool, error) {
	active := true

	var existingNotification models.BookNotification
	err := d.gorm.Where(models.BookNotification{
		BookAlertID: alert.ID,
		BookSlotID:  bookSlot.ID,
		Active:      &active,
	}).First(&existingNotification).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	return true, err
}

func (d *DisneyDatabase) ActiveNotifications() ([]models.BookNotification, error) {
	active := true
	var notifications []models.BookNotification
	err := d.gorm.Where(models.BookNotification{
		Active: &active,
	}).Preload("BookAlert").Preload("BookSlot").Find(&notifications).Error
	return notifications, err
}

func (d *DisneyDatabase) DeactivateNotification(notification models.BookNotification) error {
	notification.Active = nil
	return d.gorm.Save(&notification).Error
}
