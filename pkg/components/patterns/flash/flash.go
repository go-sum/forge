// Package flash manages one-time user-facing messages stored in session cookies.
// Type values are deliberately aligned to shadcn/ui alert variant names so callers
// can map flash.Type to feedback.AlertVariant with a simple switch.
package flash

import (
	"encoding/json"
	"net/http"
)

// Type identifies the visual style of a flash message.
type Type string

const (
	TypeSuccess Type = "success"
	TypeInfo    Type = "info"
	TypeWarning Type = "warning"
	TypeError   Type = "error"
)

// Message is a single flash notification.
type Message struct {
	Type Type
	Text string
}

const cookieName = "flash"

// Set encodes msgs as JSON into a short-lived cookie on w.
func Set(w http.ResponseWriter, msgs []Message) error {
	data, err := json.Marshal(msgs)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    string(data),
		Path:     "/",
		MaxAge:   60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// GetAll reads and clears all flash messages from the request cookie.
// Returns an empty slice (not nil) when no messages are set.
func GetAll(r *http.Request, w http.ResponseWriter) ([]Message, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		// No cookie present — not an error.
		return []Message{}, nil
	}

	var msgs []Message
	if err := json.Unmarshal([]byte(cookie.Value), &msgs); err != nil {
		return nil, err
	}

	// Clear the cookie so messages are shown only once.
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return msgs, nil
}

// Convenience constructors.

func Success(w http.ResponseWriter, text string) error {
	return Set(w, []Message{{Type: TypeSuccess, Text: text}})
}

func Info(w http.ResponseWriter, text string) error {
	return Set(w, []Message{{Type: TypeInfo, Text: text}})
}

func Warning(w http.ResponseWriter, text string) error {
	return Set(w, []Message{{Type: TypeWarning, Text: text}})
}

func Error(w http.ResponseWriter, text string) error {
	return Set(w, []Message{{Type: TypeError, Text: text}})
}
