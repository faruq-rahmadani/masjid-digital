package controller

//go:generate go tool mockgen -source=screen.go -destination=../../../../mock/controller_screen.go -package=mock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

type ScreenUsecaseItf interface {
	List(ctx context.Context) ([]domain.Screen, error)
	Connect(ctx context.Context, req domain.ConnectScreenRequest) (domain.ScreenConnection, error)
	Update(ctx context.Context, req domain.UpdateScreenRequest) (domain.Screen, error)
	Block(ctx context.Context, id int64) (domain.Screen, error)
	Unblock(ctx context.Context, id int64) (domain.Screen, error)
	SendCommand(ctx context.Context, req domain.SendScreenCommandRequest) error
	NotifyContentChanged()
}

const (
	// screenKeepAlive is how often an idle stream gets a comment line, well under the ~60-100 s
	// after which proxies (nginx, Cloudflare) drop a silent connection
	screenKeepAlive = 25 * time.Second
	// screenRetry is how long a TV waits before reconnecting after the stream drops
	screenRetry = 5 * time.Second
	// blockedScreenRetry is how long a blocked TV waits before trying again, so an unblocked one
	// comes back on its own within minutes
	blockedScreenRetry = 5 * time.Minute
)

// ListScreens handles GET /screens: every screen, blocked ones too, which the admin shows apart
func (ctrl *Controller) ListScreens(c echo.Context) error {
	screens, err := ctrl.screen.List(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: screens})
}

// UpdateScreen handles PUT /screens/:id: its name and forced mode
func (ctrl *Controller) UpdateScreen(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}

	var req domain.UpdateScreenRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}
	req.ID = id

	screen, err := ctrl.screen.Update(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: screen})
}

// BlockScreen handles POST /screens/:id/block
func (ctrl *Controller) BlockScreen(c echo.Context) error {
	return ctrl.setBlocked(c, ctrl.screen.Block)
}

// UnblockScreen handles DELETE /screens/:id/block
func (ctrl *Controller) UnblockScreen(c echo.Context) error {
	return ctrl.setBlocked(c, ctrl.screen.Unblock)
}

func (ctrl *Controller) setBlocked(c echo.Context, set func(context.Context, int64) (domain.Screen, error)) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}

	screen, err := set(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: screen})
}

// SendScreenCommand handles POST /screens/:id/commands, and POST /screens/commands for every
// connected screen
func (ctrl *Controller) SendScreenCommand(c echo.Context) error {
	var req domain.SendScreenCommandRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}
	if c.Param("id") != "" {
		id, err := paramID(c)
		if err != nil {
			return err
		}
		req.ScreenID = id
	}

	if err := ctrl.screen.SendCommand(c.Request().Context(), req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, messageResponse{Message: "perintah dikirim ke layar"})
}

// NotifyScreensContentChanged tells every TV to reload its data now
func (ctrl *Controller) NotifyScreensContentChanged() {
	ctrl.screen.NotifyContentChanged()
}

// StreamScreenEvents handles the public GET /screens/:device_id/events?name=…&v=…: the TV's one
// long-lived connection (server-sent events). It registers the device the first time, marks it
// online while open, and carries its state, content refreshes and commands.
func (ctrl *Controller) StreamScreenEvents(c echo.Context) error {
	conn, err := ctrl.screen.Connect(c.Request().Context(), domain.ConnectScreenRequest{
		DeviceID:   c.Param("device_id"),
		NameHint:   c.QueryParam("name"),
		AppVersion: c.QueryParam("v"),
		UserAgent:  c.Request().UserAgent(),
	})
	if errors.Is(err, domain.ErrScreenBlocked) {
		// a 200 with the event, not a 403, so the browser keeps retrying (slowly) and an
		// unblocked TV comes back by itself
		stream := startEventStream(c)
		return stream.send(blockedScreenRetry, domain.ScreenEvent{Type: domain.ScreenEventBlocked})
	}
	if err != nil {
		return err
	}
	defer conn.Close()

	stream := startEventStream(c)
	if err := stream.send(screenRetry, domain.ScreenEvent{Type: domain.ScreenEventState, Data: conn.State}); err != nil {
		return nil
	}

	keepAlive := time.NewTicker(screenKeepAlive)
	defer keepAlive.Stop()
	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case <-keepAlive.C:
			if err := stream.comment("ping"); err != nil {
				return nil
			}
		case event, ok := <-conn.Events:
			if !ok {
				return nil
			}
			if err := stream.send(0, event); err != nil || event.Type == domain.ScreenEventBlocked {
				return nil
			}
		}
	}
}

// eventStream writes server-sent events to the response
type eventStream struct {
	res *echo.Response
}

func startEventStream(c echo.Context) eventStream {
	res := c.Response()
	header := res.Header()
	header.Set(echo.HeaderContentType, "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no") // nginx would otherwise hold events back
	res.WriteHeader(http.StatusOK)

	// the server's write timeout is meant for normal requests and would cut the stream
	_ = http.NewResponseController(res).SetWriteDeadline(time.Time{})
	res.Flush()
	return eventStream{res: res}
}

// send writes one event, with the reconnect delay first when retry is set
func (s eventStream) send(retry time.Duration, event domain.ScreenEvent) error {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}
	if retry > 0 {
		if _, err := fmt.Fprintf(s.res, "retry: %d\n", retry.Milliseconds()); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(s.res, "event: %s\ndata: %s\n\n", event.Type, data); err != nil {
		return err
	}
	s.res.Flush()
	return nil
}

func (s eventStream) comment(text string) error {
	if _, err := fmt.Fprintf(s.res, ": %s\n\n", text); err != nil {
		return err
	}
	s.res.Flush()
	return nil
}
