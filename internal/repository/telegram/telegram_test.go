package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testToken = "123456:secret-token"

func TestSend(t *testing.T) {
	var got sendMessageRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/bot"+testToken+"/sendMessage" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer server.Close()

	bot := New(testToken, "-1001234")
	bot.baseURL = server.URL

	if err := bot.Send(context.Background(), "Assalamu'alaikum <b>"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	want := sendMessageRequest{ChatID: "-1001234", Text: "Assalamu'alaikum <b>", DisableWebPagePreview: true}
	if got != want {
		t.Errorf("body = %+v, want %+v", got, want)
	}
}

func TestSendErrors(t *testing.T) {
	t.Run("rejected by Telegram", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"ok":false,"description":"Bad Request: chat not found"}`))
		}))
		defer server.Close()

		bot := New(testToken, "1")
		bot.baseURL = server.URL
		err := bot.Send(context.Background(), "x")
		if err == nil || !strings.Contains(err.Error(), "chat not found") {
			t.Errorf("Send error = %v, want chat not found", err)
		}
		if err != nil && strings.Contains(err.Error(), testToken) {
			t.Errorf("error leaks the token: %v", err)
		}
	})

	t.Run("unreachable", func(t *testing.T) {
		server := httptest.NewServer(http.NotFoundHandler())
		server.Close()

		bot := New(testToken, "1")
		bot.baseURL = server.URL
		err := bot.Send(context.Background(), "x")
		if err == nil {
			t.Fatal("Send error = nil, want one")
		}
		if strings.Contains(err.Error(), testToken) {
			t.Errorf("error leaks the token: %v", err)
		}
	})
}
