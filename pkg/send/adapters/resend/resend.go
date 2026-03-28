// Package resend delivers email via the Resend HTTP API.
// See https://resend.com/docs/api-reference/emails/send-email.
package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-sum/send/core"
)

const defaultAPIURL = "https://api.resend.com/emails"

// Sender delivers messages via the Resend HTTP API.
type Sender struct {
	apiKey   string
	sendFrom string
	apiURL   string
	client   *http.Client
}

// New constructs a Sender using the given API key and default sender address.
func New(apiKey, sendFrom string) *Sender {
	return &Sender{
		apiKey:   apiKey,
		sendFrom: sendFrom,
		apiURL:   defaultAPIURL,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// NewWithURL constructs a Sender with a custom API endpoint. Use in tests to
// point the sender at a local httptest.Server instead of the live Resend API.
func NewWithURL(apiKey, sendFrom, url string) *Sender {
	return &Sender{apiKey: apiKey, sendFrom: sendFrom, apiURL: url, client: &http.Client{Timeout: 10 * time.Second}}
}

type payload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

// Send delivers msg via the Resend API. It returns an error when the request
// cannot be constructed, the HTTP call fails, or the API returns a non-2xx status.
func (s *Sender) Send(ctx context.Context, msg core.Message) error {
	from := msg.From
	if from == "" {
		from = s.sendFrom
	}

	body, err := json.Marshal(payload{
		From:    from,
		To:      []string{msg.To},
		Subject: msg.Subject,
		HTML:    msg.HTML,
		Text:    msg.Text,
	})
	if err != nil {
		return fmt.Errorf("resend: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("resend: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("resend: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("resend: status %d: %s", resp.StatusCode, bytes.TrimSpace(errBody))
	}
	return nil
}
