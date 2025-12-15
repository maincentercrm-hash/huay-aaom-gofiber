package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"go-server/models"
)

func GetCurrentBet(config models.Config, userID string, startDate, endDate time.Time) (float64, error) {
	log.Println("utils.GetCurrentBet: Starting")

	// Return a fixed value of 500 for currentBet
	/*
		currentBet := 500.0
		log.Printf("utils.GetCurrentBet: Returning fixed currentBet value of %.2f for user %s", currentBet, userID)
		return currentBet, nil
	*/

	url := fmt.Sprintf("%s/players/v1/line/bets?line_id=%s&line_at=%s&start_date=%d&end_date=%d",
		config.ApiEndpoint,
		userID,
		config.LineAt,
		startDate.Unix(),
		endDate.Unix())

	log.Printf("utils.GetCurrentBet: Constructed URL - %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("utils.GetCurrentBet: Error creating request - %v", err)
		return 0, err
	}

	req.Header.Add("API-KEY", config.ApiKey)
	log.Println("utils.GetCurrentBet: Added API-KEY to request header")

	client := &http.Client{}
	log.Println("utils.GetCurrentBet: Sending request")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("utils.GetCurrentBet: Error sending request - %v", err)
		return 0, err
	}
	defer resp.Body.Close()

	log.Printf("utils.GetCurrentBet: Received response with status code %d", resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("utils.GetCurrentBet: Error reading response body - %v", err)
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("utils.GetCurrentBet: API returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Bet float64 `json:"bet"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("utils.GetCurrentBet: Error unmarshaling response - %v", err)
		return 0, err
	}

	log.Printf("utils.GetCurrentBet: Successfully retrieved bet for user %s: %.2f", userID, result.Bet)
	return result.Bet, nil

}
