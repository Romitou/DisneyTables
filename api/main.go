package api

import (
	"bytes"
	"encoding/json"
	"github.com/romitou/disneytables/database"
	"net/http"
	"os"
)

func addCustomHeaders(request *http.Request) {
	rawEnvHeaders := os.Getenv("CUSTOM_HEADERS")
	if rawEnvHeaders == "" {
		return
	}

	var customHeaders map[string]string
	err := json.Unmarshal([]byte(rawEnvHeaders), &customHeaders)
	if err != nil {
		return
	}

	for key, value := range customHeaders {
		request.Header.Set(key, value)
	}
}

func addAuthHeaders(request *http.Request) {
	request.Header.Set("x-api-key", os.Getenv("API_KEY"))
	firstAuthDetails, err := database.Get().LastAuthDetails()
	if err != nil {
		return
	}
	request.Header.Set("authorization", "BEARER "+firstAuthDetails.AccessToken)
}

type RestaurantAvailabilitySearch struct {
	Date         string `json:"date"`
	RestaurantID string `json:"restaurantId"`
	PartyMix     int    `json:"partyMix"`
}

type RestaurantAvailability struct {
	StartTime   string                 `json:"startTime"`
	EndTime     string                 `json:"endTime"`
	Date        string                 `json:"date"`
	Status      string                 `json:"status"`
	MealPeriods []RestaurantMealPeriod `json:"mealPeriods"`
}

type RestaurantMealPeriod struct {
	MealPeriod string               `json:"mealPeriod"`
	MealSlots  []RestaurantMealSlot `json:"slotList"`
}

type RestaurantMealSlot struct {
	Time      string `json:"time"`
	Available string `json:"available"`
}

func RestaurantAvailabilities(data RestaurantAvailabilitySearch) ([]RestaurantAvailability, error) {
	marshalData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", os.Getenv("AVAILABILITIES_ENDPOINT"), bytes.NewBuffer(marshalData))
	if err != nil {
		return nil, err
	}

	addCustomHeaders(req)
	addAuthHeaders(req)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var responseData []RestaurantAvailability
	err = json.NewDecoder(response.Body).Decode(&responseData)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

type Restaurant struct {
	Name             string `json:"name"`
	DisneyID         string `json:"id"`
	BookingAvailable bool   `json:"drsApp"`
	HeroMediaMobile  struct {
		URL string `json:"url"`
	} `json:"heroMediaMobile"`
}

type RestaurantResponse struct {
	Data struct {
		Activities []Restaurant `json:"activities"`
	} `json:"data"`
}

func Restaurants() ([]Restaurant, error) {
	query := os.Getenv("RESTAURANTS_QUERY")
	req, err := http.NewRequest("POST", os.Getenv("GRAPHQL_ENDPOINT"), bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, err
	}

	addCustomHeaders(req)
	addAuthHeaders(req)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var responseData RestaurantResponse
	err = json.NewDecoder(response.Body).Decode(&responseData)
	if err != nil {
		return nil, err
	}

	return responseData.Data.Activities, nil
}

type DisneyAuth struct {
	Data struct {
		DisneyToken DisneyToken `json:"token"`
	} `json:"data"`
}

type DisneyToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func RefreshAuth(refreshToken string) (DisneyToken, error) {
	jsonData := `{"refreshToken":"` + refreshToken + `"}`
	req, err := http.NewRequest("POST", os.Getenv("REFRESH_AUTH_ENDPOINT"), bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return DisneyToken{}, err
	}

	addCustomHeaders(req)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return DisneyToken{}, err
	}

	var responseData DisneyAuth
	err = json.NewDecoder(response.Body).Decode(&responseData)
	if err != nil {
		return DisneyToken{}, err
	}

	return responseData.Data.DisneyToken, nil
}
