package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"avito-shop/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestE2E(t *testing.T) {
	baseURL := "http://localhost:8080"


	tokenA, err := auth(baseURL, "userA", "passA")
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenA)

	tokenB, err := auth(baseURL, "userB", "passB")
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenB)

	err = sendCoin(baseURL, tokenA, "userB", 100)
	assert.NoError(t, err)

	err = buyItem(baseURL, tokenB, "cup")
	assert.NoError(t, err)

	infoA, err := getInfo(baseURL, tokenA)
	assert.NoError(t, err)
	infoB, err := getInfo(baseURL, tokenB)
	assert.NoError(t, err)

	assert.Equal(t, 900, infoA.Coins)

	assert.Equal(t, 1080, infoB.Coins)

	foundCup := false
	for _, inv := range infoB.Inventory {
		if inv.Type == "cup" && inv.Quantity == 1 {
			foundCup = true
			break
		}
	}
	assert.True(t, foundCup, "B must have 1 cup")

	sentOK := false
	for _, s := range infoA.CoinHistory.Sent {
		if s.ToUser == "userB" && s.Amount == 100 {
			sentOK = true
			break
		}
	}
	assert.True(t, sentOK, "A must have sent 100 coins to B")

	receivedOK := false
	purchaseOK := false
	for _, r := range infoB.CoinHistory.Received {
		if r.FromUser == "userA" && r.Amount == 100 {
			receivedOK = true
			break
		}
	}
	for _, s := range infoB.CoinHistory.Sent {
		if s.ToUser == "store" && s.Amount == 20 {
			purchaseOK = true
			break
		}
	}
	assert.True(t, receivedOK, "B must have received 100 from A")
	assert.True(t, purchaseOK, "B must have sent 20 to store when buying cup")
}




func auth(baseURL, username, password string) (string, error) {
	reqBody, _ := json.Marshal(models.AuthRequest{
		Username: username,
		Password: password,
	})
	resp, err := http.Post(baseURL+"/api/auth", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth failed with status %d", resp.StatusCode)
	}

	var ar models.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return "", err
	}
	return ar.Token, nil
}

func getInfo(baseURL, token string) (*models.InfoResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/api/info", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getInfo failed with status %d", resp.StatusCode)
	}
	var ir models.InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&ir); err != nil {
		return nil, err
	}
	return &ir, nil
}

func sendCoin(baseURL, token, toUser string, amount int) error {
	reqData := models.SendCoinRequest{
		ToUser: toUser,
		Amount: amount,
	}
	body, _ := json.Marshal(reqData)
	req, err := http.NewRequest("POST", baseURL+"/api/sendCoin", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sendCoin failed with status %d", resp.StatusCode)
	}
	return nil
}

func buyItem(baseURL, token, item string) error {
	req, err := http.NewRequest("GET", baseURL+"/api/buy/"+item, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("buyItem failed with status %d", resp.StatusCode)
	}
	return nil
}
