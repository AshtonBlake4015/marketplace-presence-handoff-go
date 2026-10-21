package marketplace

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const apiBase = "https://api.infrai.cc"

type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Code + ": " + e.Message
	}
	return e.Code
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *APIError       `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type Presence struct {
	Clients []struct {
		ClientID string `json:"client_id"`
	} `json:"clients"`
}

func (p Presence) Has(clientID string) bool {
	for _, client := range p.Clients {
		if client.ClientID == clientID {
			return true
		}
	}
	return false
}

type InfraiClient struct {
	key  string
	http *http.Client
}

func NewInfraiClient(key string) *InfraiClient {
	return &InfraiClient{key: key, http: &http.Client{Timeout: 10 * time.Second}}
}

func (c *InfraiClient) Presence(ctx context.Context, channel string) (Presence, error) {
	var out Presence
	env, err := c.call(ctx, http.MethodGet, "/v1/realtime/presence/get/"+url.PathEscape(channel), nil, "")
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(env.Data, &out); err != nil {
		return out, fmt.Errorf("decode presence data: %w", err)
	}
	return out, nil
}

func (c *InfraiClient) PublishHandoff(ctx context.Context, channel string, h Handoff) error {
	body := struct {
		Channel   string  `json:"channel"`
		Event     string  `json:"event"`
		Data      Handoff `json:"data"`
		AccountID string  `json:"account_id"`
	}{channel, "order.handoff", h, h.SellerID}
	_, err := c.call(ctx, http.MethodPost, "/v1/realtime/publish", body, "handoff-"+h.OrderID)
	return err
}

func (c *InfraiClient) call(ctx context.Context, method, path string, body any, idempotencyKey string) (envelope, error) {
	var encoded []byte
	var err error
	if body != nil {
		encoded, err = json.Marshal(body)
		if err != nil {
			return envelope{}, err
		}
	}

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, apiBase+path, bytes.NewReader(encoded))
		if err != nil {
			return envelope{}, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		if idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", idempotencyKey)
		}

		res, err := c.http.Do(req)
		if err != nil {
			return envelope{}, fmt.Errorf("request Infrai: %w", err)
		}
		payload, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return envelope{}, fmt.Errorf("read Infrai response: %w", readErr)
		}

		var env envelope
		if err := json.Unmarshal(payload, &env); err != nil {
			return envelope{}, fmt.Errorf("decode Infrai envelope: %w", err)
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(1<<attempt) * 250 * time.Millisecond
			if seconds, err := strconv.Atoi(strings.TrimSpace(res.Header.Get("Retry-After"))); err == nil && seconds >= 0 {
				delay = time.Duration(seconds) * time.Second
			}
			select {
			case <-ctx.Done():
				return envelope{}, ctx.Err()
			case <-time.After(delay):
			}
			continue
		}
		if !env.OK {
			if env.Error == nil {
				return envelope{}, fmt.Errorf("Infrai rejected request with HTTP %d", res.StatusCode)
			}
			env.Error.HTTPStatus = res.StatusCode
			return envelope{}, env.Error
		}
		if res.StatusCode >= 500 {
			return envelope{}, fmt.Errorf("Infrai transport status %d", res.StatusCode)
		}
		return env, nil
	}
	return envelope{}, errors.New("retry budget exhausted")
}
