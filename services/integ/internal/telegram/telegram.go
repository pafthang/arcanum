package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is a minimal Telegram Bot API client.
type Client struct {
	Token      string
	HTTPClient *http.Client
}

// New creates a Telegram client with a bot token.
func New(token string) *Client {
	return &Client{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendMessageRequest is Telegram API /sendMessage payload.
type SendMessageRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

// SendMessageResponse is Telegram API response.
type SendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

// SendMessage posts text to a Telegram chat.
func (c *Client) SendMessage(ctx context.Context, chatID, text string) error {
	if c.Token == "" {
		return fmt.Errorf("telegram token missing")
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.Token)
	reqBody, err := json.Marshal(SendMessageRequest{ChatID: chatID, Text: text})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res SendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	if !res.OK {
		return fmt.Errorf("telegram api error: %s", res.Description)
	}
	return nil
}
