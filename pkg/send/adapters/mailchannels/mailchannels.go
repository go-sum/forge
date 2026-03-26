// Package mailchannels delivers email via the MailChannels TX HTTP API.
// See https://api.mailchannels.net/tx/v1/documentation.
package mailchannels

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-sum/send/core"
)

const defaultAPIURL = "https://api.mailchannels.net/tx/v1/send"

// Sender delivers messages via the MailChannels TX HTTP API.
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
		client:   &http.Client{},
	}
}

// NewWithURL constructs a Sender with a custom API endpoint. Use in tests to
// point the sender at a local httptest.Server instead of the live MailChannels API.
func NewWithURL(apiKey, sendFrom, url string) *Sender {
	return &Sender{apiKey: apiKey, sendFrom: sendFrom, apiURL: url, client: &http.Client{}}
}

type mcAddress struct {
	Email string `json:"email"`
}

type mcContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type mcPersonalisation struct {
	To []mcAddress `json:"to"`
}

type mcPayload struct {
	From             mcAddress           `json:"from"`
	Subject          string              `json:"subject"`
	Personalizations []mcPersonalisation `json:"personalizations"`
	Content          []mcContent         `json:"content"`
}

// Send delivers msg via the MailChannels TX API. It returns an error when the
// request cannot be constructed, the HTTP call fails, or the API returns a
// non-2xx status.
func (s *Sender) Send(ctx context.Context, msg core.Message) error {
	from := msg.From
	if from == "" {
		from = s.sendFrom
	}

	p := mcPayload{
		From:    mcAddress{Email: from},
		Subject: msg.Subject,
		Personalizations: []mcPersonalisation{
			{To: []mcAddress{{Email: msg.To}}},
		},
	}
	if msg.HTML != "" {
		p.Content = append(p.Content, mcContent{Type: "text/html", Value: msg.HTML})
	}
	if msg.Text != "" {
		p.Content = append(p.Content, mcContent{Type: "text/plain", Value: msg.Text})
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("mailchannels: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("mailchannels: build request: %w", err)
	}
	req.Header.Set("X-API-Key", s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("mailchannels: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("mailchannels: unexpected status %d", resp.StatusCode)
	}
	return nil
}
