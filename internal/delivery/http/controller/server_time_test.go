package controller

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestGetServerTime(t *testing.T) {
	before := time.Now().UnixMilli()
	rec, err := serve(NewController(nil, nil, nil, nil, nil, nil).GetServerTime, http.MethodGet, "/api/public/v1/time", "", "")
	after := time.Now().UnixMilli()
	if err != nil {
		t.Fatalf("handler: %v", err)
	}

	var body struct {
		Data struct {
			Now int64 `json:"now"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %s: %v", rec.Body.String(), err)
	}
	if body.Data.Now < before || body.Data.Now > after {
		t.Errorf("now = %d, want between %d and %d", body.Data.Now, before, after)
	}
	if got := rec.Header().Get(echo.HeaderCacheControl); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
}
