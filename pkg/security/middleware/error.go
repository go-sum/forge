package middleware

// Error is a transport-facing middleware error with a safe public message.
type Error struct {
	Status  int
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

func (e *Error) StatusCode() int { return e.Status }

func (e *Error) PublicMessage() string { return e.Message }
