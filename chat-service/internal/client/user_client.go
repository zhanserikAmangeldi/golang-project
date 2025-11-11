package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type UserClient struct {
	BaseURL string
	client  *http.Client
}

func (u *UserClient) init() {
	if u.client == nil {
		u.client = &http.Client{Timeout: 5 * time.Second}
	}
}

func (u *UserClient) UserExists(userID string) (bool, error) {
	u.init()
	// Example endpoint expected on user-service: GET /users/{id}
	url := fmt.Sprintf("%s/users/%s", u.BaseURL, userID)
	resp, err := u.client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		// optional: decode response to check active/disabled, etc.
		var body map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, nil
}
