// Package telegram sends messages to one Telegram chat through a bot, see
// https://core.telegram.org/bots/api#sendmessage
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL = "https://api.telegram.org"
	sendTimeout    = 10 * time.Second
)

// Bot posts plain text messages to one chat
type Bot struct {
	token   string
	chatID  string
	baseURL string
	client  *http.Client
}

// New returns a Bot that posts to chatID: a user ID, or a group's (negative) ID with the bot in it
func New(token, chatID string) *Bot {
	return &Bot{token: token, chatID: chatID, baseURL: defaultBaseURL, client: &http.Client{Timeout: sendTimeout}}
}

type sendMessageRequest struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

type apiResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

// Send posts text as is: no parse mode, so whatever a sender typed is shown, never formatted.
// Its errors never contain the bot token.
func (b *Bot) Send(ctx context.Context, text string) error {
	body, err := json.Marshal(sendMessageRequest{ChatID: b.chatID, Text: text, DisableWebPagePreview: true})
	if err != nil {
		return err
	}

	// the token is part of the URL, so the URL must not reach an error message or a log
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.baseURL+"/bot"+b.token+"/sendMessage", bytes.NewReader(body))
	if err != nil {
		return errors.New("telegram: invalid request")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := b.client.Do(req)
	if err != nil {
		if urlErr, ok := errors.AsType[*url.Error](err); ok {
			err = urlErr.Err
		}
		return fmt.Errorf("telegram: send message: %w", err)
	}
	defer res.Body.Close()

	var answer apiResponse
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&answer); err != nil {
		return fmt.Errorf("telegram: send message: status %d", res.StatusCode)
	}
	if !answer.OK {
		return fmt.Errorf("telegram: send message: status %d: %s", res.StatusCode, answer.Description)
	}

	return nil
}
