package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/romitou/disneytables/database"
	"io"
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

type RestaurantAvailabilityError struct {
	Err            error
	HttpStatusCode int
	RawData        string
}

func RestaurantAvailabilities(data RestaurantAvailabilitySearch) ([]RestaurantAvailability, *RestaurantAvailabilityError) {
	marshalData, err := json.Marshal(data)
	if err != nil {
		return nil, &RestaurantAvailabilityError{Err: err}
	}

	req, err := http.NewRequest("POST", os.Getenv("AVAILABILITIES_ENDPOINT"), bytes.NewBuffer(marshalData))
	if err != nil {
		return nil, &RestaurantAvailabilityError{Err: err}
	}

	addCustomHeaders(req)
	addAuthHeaders(req)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &RestaurantAvailabilityError{Err: err}
	}

	var responseData []RestaurantAvailability
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, &RestaurantAvailabilityError{Err: err, RawData: string(body)}
	}

	if response.StatusCode != 200 {
		return nil, &RestaurantAvailabilityError{
			Err:            errors.New("non-200 status code"),
			RawData:        string(body),
			HttpStatusCode: response.StatusCode,
		}
	}

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, &RestaurantAvailabilityError{Err: err, RawData: string(body)}
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
