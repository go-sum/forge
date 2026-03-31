// Package model defines auth domain types: user identity, verification flows, and auth errors.
package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Role constants for user access levels.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// FlowPurpose identifies the verification workflow in progress.
type FlowPurpose string

const (
	FlowPurposeSignup      FlowPurpose = "signup"
	FlowPurposeSignin      FlowPurpose = "signin"
	FlowPurposeEmailChange FlowPurpose = "email_change"
)

// User is the auth module's view of an application user.
// It contains only the fields required for authentication and session management.
type User struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
	Role        string
	Verified    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// BeginSignupInput carries validated data for starting a signup verification flow.
type BeginSignupInput struct {
	Email       string `form:"email"         validate:"required,email,max=255"`
	DisplayName string `form:"display_name"  validate:"required,min=1,max=255"`
}

// BeginSigninInput carries the email address for starting a signin verification flow.
type BeginSigninInput struct {
	Email string `form:"email" validate:"required,email,max=255"`
}

// BeginEmailChangeInput carries the target email address for a signed-in user.
type BeginEmailChangeInput struct {
	Email string `form:"email" validate:"required,email,max=255"`
}

// VerifyInput carries data from the verification form.
type VerifyInput struct {
	Code  string `form:"code"  validate:"required,len=6,numeric"`
	Token string `form:"token" validate:"omitempty"`
}

// PendingFlow is the browser-bound verification state retained between the begin
// and verify steps.
type PendingFlow struct {
	Purpose     FlowPurpose `json:"purpose"`
	Email       string      `json:"email"`
	DisplayName string      `json:"display_name,omitempty"`
	Role        string      `json:"role,omitempty"`
	UserID      uuid.UUID   `json:"user_id,omitempty"`
	Secret      string      `json:"secret"`
	IssuedAt    time.Time   `json:"issued_at"`
	ExpiresAt   time.Time   `json:"expires_at"`
}

// VerificationToken is the self-contained payload embedded in emailed verify links.
type VerificationToken struct {
	Purpose     FlowPurpose `json:"purpose"`
	Email       string      `json:"email"`
	DisplayName string      `json:"display_name,omitempty"`
	Role        string      `json:"role,omitempty"`
	UserID      uuid.UUID   `json:"user_id,omitempty"`
	Secret      string      `json:"secret"`
	IssuedAt    time.Time   `json:"issued_at"`
	ExpiresAt   time.Time   `json:"expires_at"`
}

// DeliveryInput is the email payload required for sending a verification message.
type DeliveryInput struct {
	Purpose   FlowPurpose
	Email     string
	Code      string
	VerifyURL string
	ExpiresAt time.Time
}

// VerifyResult describes a successful verification.
type VerifyResult struct {
	Purpose FlowPurpose
	User    User
}

// VerifyPageState supplies the verification screen with a purpose and prefilled code.
type VerifyPageState struct {
	Purpose   FlowPurpose
	Code      string
	Token     string
	Email     string
	CanResend bool
}

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrEmailTaken              = errors.New("email already in use")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrVerificationExpired     = errors.New("verification expired")
	ErrVerificationMissing     = errors.New("verification missing")
	ErrUnsupportedMethod       = errors.New("unsupported auth method")
)
