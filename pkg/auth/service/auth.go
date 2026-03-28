// Package service provides the AuthService which handles passwordless auth flows.
package service

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // RFC 6238 TOTP uses HMAC-SHA1 by default.
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-sum/auth"
	"github.com/go-sum/auth/model"
	"github.com/go-sum/auth/repository"
	email "github.com/go-sum/componentry/email"
	"github.com/go-sum/send"
	"github.com/google/uuid"
	g "maragu.dev/gomponents"
)

type clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

// TokenCodec encodes and decodes self-contained verification link payloads.
type TokenCodec interface {
	Encode(model.VerificationToken) (string, error)
	Decode(string) (model.VerificationToken, error)
}

// AuthService handles email-TOTP signup, signin, and email-change flows.
type AuthService struct {
	users      repository.UserStore
	sender     send.Sender
	method     auth.EmailTOTPMethodConfig
	sendFrom   string
	tokenCodec TokenCodec
	clock      clock
}

// Config parameterises the passwordless auth service.
type Config struct {
	Method     auth.EmailTOTPMethodConfig
	SendFrom   string
	TokenCodec TokenCodec
	Clock      clock
}

// NewAuthService constructs an AuthService from its repository, email, and token dependencies.
func NewAuthService(users repository.UserStore, sender send.Sender, cfg Config) *AuthService {
	if cfg.Clock == nil {
		cfg.Clock = systemClock{}
	}
	return &AuthService{
		users:      users,
		sender:     sender,
		method:     cfg.Method,
		sendFrom:   cfg.SendFrom,
		tokenCodec: cfg.TokenCodec,
		clock:      cfg.Clock,
	}
}

// BeginSignup starts a signup verification flow and sends the verification email.
func (s *AuthService) BeginSignup(ctx context.Context, input model.BeginSignupInput, verifyPath string) (model.PendingFlow, error) {
	if !s.method.Enabled {
		return model.PendingFlow{}, model.ErrUnsupportedMethod
	}

	if _, err := s.users.GetByEmail(ctx, input.Email); err == nil {
		return model.PendingFlow{}, model.ErrEmailTaken
	} else if !errors.Is(err, model.ErrUserNotFound) {
		return model.PendingFlow{}, fmt.Errorf("lookup signup email: %w", err)
	}

	role := input.Role
	if role == "" {
		role = model.RoleUser
	}

	flow, code, err := s.newPendingFlow(model.FlowPurposeSignup, input.Email, input.DisplayName, role, uuid.Nil)
	if err != nil {
		return model.PendingFlow{}, err
	}
	if err := s.deliver(ctx, flow, code, verifyPath); err != nil {
		return model.PendingFlow{}, err
	}
	return flow, nil
}

// BeginSignin starts a signin verification flow. It always returns a pending flow
// so the caller can redirect to verification without leaking account existence.
func (s *AuthService) BeginSignin(ctx context.Context, input model.BeginSigninInput, verifyPath string) (model.PendingFlow, error) {
	if !s.method.Enabled {
		return model.PendingFlow{}, model.ErrUnsupportedMethod
	}

	flow, code, err := s.newPendingFlow(model.FlowPurposeSignin, input.Email, "", "", uuid.Nil)
	if err != nil {
		return model.PendingFlow{}, err
	}

	user, err := s.users.GetByEmail(ctx, input.Email)
	switch {
	case err == nil && user.Verified:
		if err := s.deliver(ctx, flow, code, verifyPath); err != nil {
			return model.PendingFlow{}, err
		}
	case err == nil:
		// Preserve anti-enumeration by suppressing delivery for non-verified accounts.
	case errors.Is(err, model.ErrUserNotFound):
		// Preserve anti-enumeration by pretending the flow exists.
	case err != nil:
		return model.PendingFlow{}, fmt.Errorf("lookup signin email: %w", err)
	}

	return flow, nil
}

// BeginEmailChange starts an email-change verification flow for the signed-in user.
func (s *AuthService) BeginEmailChange(ctx context.Context, userID uuid.UUID, input model.BeginEmailChangeInput, verifyPath string) (model.PendingFlow, error) {
	if !s.method.Enabled {
		return model.PendingFlow{}, model.ErrUnsupportedMethod
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return model.PendingFlow{}, fmt.Errorf("lookup current user: %w", err)
	}

	if strings.EqualFold(user.Email, input.Email) {
		return model.PendingFlow{}, model.ErrEmailTaken
	}

	if _, err := s.users.GetByEmail(ctx, input.Email); err == nil {
		return model.PendingFlow{}, model.ErrEmailTaken
	} else if !errors.Is(err, model.ErrUserNotFound) {
		return model.PendingFlow{}, fmt.Errorf("lookup email change target: %w", err)
	}

	flow, code, err := s.newPendingFlow(model.FlowPurposeEmailChange, input.Email, "", "", userID)
	if err != nil {
		return model.PendingFlow{}, err
	}
	if err := s.deliver(ctx, flow, code, verifyPath); err != nil {
		return model.PendingFlow{}, err
	}
	return flow, nil
}

// ResendPendingFlow starts a fresh verification cycle for the current pending flow.
func (s *AuthService) ResendPendingFlow(ctx context.Context, flow model.PendingFlow, verifyPath string) (model.PendingFlow, error) {
	switch flow.Purpose {
	case model.FlowPurposeSignup:
		return s.BeginSignup(ctx, model.BeginSignupInput{
			Email:       flow.Email,
			DisplayName: flow.DisplayName,
			Role:        flow.Role,
		}, verifyPath)
	case model.FlowPurposeSignin:
		return s.BeginSignin(ctx, model.BeginSigninInput{
			Email: flow.Email,
		}, verifyPath)
	case model.FlowPurposeEmailChange:
		return s.BeginEmailChange(ctx, flow.UserID, model.BeginEmailChangeInput{
			Email: flow.Email,
		}, verifyPath)
	default:
		return model.PendingFlow{}, model.ErrUnsupportedMethod
	}
}

// VerifyPendingFlow completes a same-browser verification using pending session state.
func (s *AuthService) VerifyPendingFlow(ctx context.Context, flow model.PendingFlow, input model.VerifyInput) (model.VerifyResult, error) {
	if err := s.validateCode(flow.Secret, flow.IssuedAt, flow.ExpiresAt, input.Code); err != nil {
		return model.VerifyResult{}, err
	}
	return s.finishVerification(ctx, flow)
}

// VerifyToken completes a cross-browser verification using a self-contained token.
func (s *AuthService) VerifyToken(ctx context.Context, token string, input model.VerifyInput) (model.VerifyResult, error) {
	payload, err := s.tokenCodec.Decode(token)
	if err != nil {
		return model.VerifyResult{}, err
	}

	if err := s.validateCode(payload.Secret, payload.IssuedAt, payload.ExpiresAt, input.Code); err != nil {
		return model.VerifyResult{}, err
	}

	return s.finishVerification(ctx, model.PendingFlow{
		Purpose:     payload.Purpose,
		Email:       payload.Email,
		DisplayName: payload.DisplayName,
		Role:        payload.Role,
		UserID:      payload.UserID,
		Secret:      payload.Secret,
		IssuedAt:    payload.IssuedAt,
		ExpiresAt:   payload.ExpiresAt,
	})
}

// VerifyPageState decodes an emailed token so the verification page can prefill the code.
func (s *AuthService) VerifyPageState(token string) (model.VerifyPageState, error) {
	payload, err := s.tokenCodec.Decode(token)
	if err != nil {
		return model.VerifyPageState{}, err
	}
	code, err := s.generateCode(payload.Secret, payload.IssuedAt)
	if err != nil {
		return model.VerifyPageState{}, err
	}
	return model.VerifyPageState{
		Purpose: payload.Purpose,
		Code:    code,
		Token:   token,
		Email:   payload.Email,
	}, nil
}

func (s *AuthService) newPendingFlow(
	purpose model.FlowPurpose,
	email, displayName, role string,
	userID uuid.UUID,
) (model.PendingFlow, string, error) {
	now := s.clock.Now()
	secret, err := randomSecret()
	if err != nil {
		return model.PendingFlow{}, "", fmt.Errorf("generate verification secret: %w", err)
	}

	period := time.Duration(s.method.PeriodSeconds) * time.Second
	if period <= 0 {
		period = 5 * time.Minute
	}

	flow := model.PendingFlow{
		Purpose:     purpose,
		Email:       email,
		DisplayName: displayName,
		Role:        role,
		UserID:      userID,
		Secret:      secret,
		IssuedAt:    now,
		ExpiresAt:   now.Add(period),
	}

	code, err := s.generateCode(secret, now)
	if err != nil {
		return model.PendingFlow{}, "", err
	}

	return flow, code, nil
}

func (s *AuthService) deliver(ctx context.Context, flow model.PendingFlow, code, verifyPath string) error {
	token, err := s.tokenCodec.Encode(model.VerificationToken{
		Purpose:     flow.Purpose,
		Email:       flow.Email,
		DisplayName: flow.DisplayName,
		Role:        flow.Role,
		UserID:      flow.UserID,
		Secret:      flow.Secret,
		IssuedAt:    flow.IssuedAt,
		ExpiresAt:   flow.ExpiresAt,
	})
	if err != nil {
		return fmt.Errorf("encode verification token: %w", err)
	}

	verifyURL, err := appendVerifyToken(verifyPath, token)
	if err != nil {
		return fmt.Errorf("build verify url: %w", err)
	}

	msg := verificationEmail(model.DeliveryInput{
		Purpose:   flow.Purpose,
		Email:     flow.Email,
		Code:      code,
		VerifyURL: verifyURL,
		ExpiresAt: flow.ExpiresAt,
	}, s.sendFrom)

	if err := s.sender.Send(ctx, msg); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}

func (s *AuthService) finishVerification(ctx context.Context, flow model.PendingFlow) (model.VerifyResult, error) {
	switch flow.Purpose {
	case model.FlowPurposeSignup:
		return s.finishSignup(ctx, flow)
	case model.FlowPurposeSignin:
		return s.finishSignin(ctx, flow)
	case model.FlowPurposeEmailChange:
		return s.finishEmailChange(ctx, flow)
	default:
		return model.VerifyResult{}, model.ErrUnsupportedMethod
	}
}

func (s *AuthService) finishSignup(ctx context.Context, flow model.PendingFlow) (model.VerifyResult, error) {
	user, err := s.users.Create(ctx, flow.Email, flow.DisplayName, defaultRole(flow.Role), true)
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			existing, getErr := s.users.GetByEmail(ctx, flow.Email)
			if getErr == nil && existing.Verified {
				return model.VerifyResult{Purpose: model.FlowPurposeSignup, User: existing}, nil
			}
		}
		return model.VerifyResult{}, fmt.Errorf("create verified user: %w", err)
	}
	return model.VerifyResult{Purpose: model.FlowPurposeSignup, User: user}, nil
}

func (s *AuthService) finishSignin(ctx context.Context, flow model.PendingFlow) (model.VerifyResult, error) {
	user, err := s.users.GetByEmail(ctx, flow.Email)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return model.VerifyResult{}, model.ErrInvalidCredentials
		}
		return model.VerifyResult{}, fmt.Errorf("lookup signin user: %w", err)
	}
	if !user.Verified {
		return model.VerifyResult{}, model.ErrInvalidCredentials
	}
	return model.VerifyResult{Purpose: model.FlowPurposeSignin, User: user}, nil
}

func (s *AuthService) finishEmailChange(ctx context.Context, flow model.PendingFlow) (model.VerifyResult, error) {
	user, err := s.users.GetByID(ctx, flow.UserID)
	if err != nil {
		return model.VerifyResult{}, fmt.Errorf("lookup email-change user: %w", err)
	}
	if strings.EqualFold(user.Email, flow.Email) {
		return model.VerifyResult{Purpose: model.FlowPurposeEmailChange, User: user}, nil
	}

	user, err = s.users.UpdateEmail(ctx, flow.UserID, flow.Email)
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			current, getErr := s.users.GetByID(ctx, flow.UserID)
			if getErr == nil && strings.EqualFold(current.Email, flow.Email) {
				return model.VerifyResult{Purpose: model.FlowPurposeEmailChange, User: current}, nil
			}
		}
		return model.VerifyResult{}, fmt.Errorf("update verified email: %w", err)
	}
	return model.VerifyResult{Purpose: model.FlowPurposeEmailChange, User: user}, nil
}

func (s *AuthService) generateCode(secret string, issuedAt time.Time) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("decode verification secret: %w", err)
	}

	period := int64(s.method.PeriodSeconds)
	if period <= 0 {
		period = int64((5 * time.Minute).Seconds())
	}

	counter := uint64(issuedAt.UTC().Unix() / period)
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], counter)

	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(msg[:])
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	code := truncated % 1_000_000
	return fmt.Sprintf("%06d", code), nil
}

func (s *AuthService) validateCode(secret string, issuedAt, expiresAt time.Time, code string) error {
	if s.clock.Now().After(expiresAt) {
		return model.ErrVerificationExpired
	}

	expected, err := s.generateCode(secret, issuedAt)
	if err != nil {
		return err
	}
	if subtleConstantCompare(expected, code) {
		return nil
	}
	return model.ErrInvalidVerificationCode
}

func defaultRole(role string) string {
	if role == "" {
		return model.RoleUser
	}
	return role
}

func randomSecret() (string, error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw), nil
}

func subtleConstantCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

func appendVerifyToken(basePath, token string) (string, error) {
	u, err := url.Parse(basePath)
	if err != nil {
		return "", err
	}
	query := u.Query()
	query.Set("token", token)
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func verificationEmail(input model.DeliveryInput, sendFrom string) send.Message {
	subject := verificationSubject(input.Purpose)
	return send.Message{
		To:      input.Email,
		From:    sendFrom,
		Subject: subject,
		HTML:    renderEmailHTML(verificationBody(input)),
		Text:    verificationText(input),
	}
}

func verificationSubject(purpose model.FlowPurpose) string {
	switch purpose {
	case model.FlowPurposeSignup:
		return "Verify your signup code"
	case model.FlowPurposeEmailChange:
		return "Verify your email change"
	default:
		return "Verify your sign in"
	}
}

func renderEmailHTML(body g.Node) string {
	var buf bytes.Buffer
	_ = body.Render(&buf)
	return buf.String()
}

func verificationBody(input model.DeliveryInput) g.Node {
	title := verificationSubject(input.Purpose)
	return email.Layout(title, g.Group([]g.Node{
		email.H1(title),
		email.P("Use this 6-digit code to continue: " + input.Code),
		email.P("Or open this secure verification link: " + input.VerifyURL),
		email.P("This code expires at " + input.ExpiresAt.UTC().Format(time.RFC1123) + "."),
	}))
}

func verificationText(input model.DeliveryInput) string {
	return strings.Join([]string{
		verificationSubject(input.Purpose),
		"",
		"Code: " + input.Code,
		"",
		"Verify link: " + input.VerifyURL,
		"",
		"Expires at: " + input.ExpiresAt.UTC().Format(time.RFC1123),
	}, "\r\n")
}

// EncryptedTokenCodec produces self-contained encrypted verification tokens.
type EncryptedTokenCodec struct {
	authKey    []byte
	encryptKey []byte
}

// NewEncryptedTokenCodec constructs a TokenCodec backed by signed AES-GCM tokens.
func NewEncryptedTokenCodec(authKey, encryptKey string) *EncryptedTokenCodec {
	return &EncryptedTokenCodec{
		authKey:    []byte(authKey),
		encryptKey: []byte(encryptKey),
	}
}

// Encode serializes a verification token into a compact transport-safe string.
func (c *EncryptedTokenCodec) Encode(token model.VerificationToken) (string, error) {
	payload, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("marshal token: %w", err)
	}
	return encryptCompactToken(c.encryptKey, c.authKey, payload)
}

// Decode validates and decrypts a previously-encoded verification token.
func (c *EncryptedTokenCodec) Decode(raw string) (model.VerificationToken, error) {
	payload, err := decryptCompactToken(c.encryptKey, c.authKey, raw)
	if err != nil {
		return model.VerificationToken{}, model.ErrVerificationMissing
	}

	var token model.VerificationToken
	if err := json.Unmarshal(payload, &token); err != nil {
		return model.VerificationToken{}, model.ErrVerificationMissing
	}
	if (systemClock{}).Now().After(token.ExpiresAt) {
		return model.VerificationToken{}, model.ErrVerificationExpired
	}
	return token, nil
}

func encryptCompactToken(encryptKey, authKey, payload []byte) (string, error) {
	envelope := struct {
		IssuedAt int64  `json:"iat"`
		Payload  string `json:"payload"`
	}{
		IssuedAt: time.Now().UTC().Unix(),
		Payload:  base64.RawURLEncoding.EncodeToString(payload),
	}

	rawEnvelope, err := json.Marshal(envelope)
	if err != nil {
		return "", err
	}

	block, err := aesCipher(encryptKey)
	if err != nil {
		return "", err
	}
	gcm, err := newGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	header := base64.RawURLEncoding.EncodeToString([]byte("forge.auth.verify.v1"))
	ciphertext := gcm.Seal(nil, nonce, rawEnvelope, []byte(header))
	body := base64.RawURLEncoding.EncodeToString(append(nonce, ciphertext...))

	mac := hmac.New(sha256.New, authKey)
	_, _ = mac.Write([]byte(header))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write([]byte(body))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return strings.Join([]string{header, body, sig}, "."), nil
}

func decryptCompactToken(encryptKey, authKey []byte, token string) ([]byte, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, model.ErrVerificationMissing
	}

	header, body, sig := parts[0], parts[1], parts[2]

	mac := hmac.New(sha256.New, authKey)
	_, _ = mac.Write([]byte(header))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write([]byte(body))
	expected := mac.Sum(nil)
	actual, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil || !hmac.Equal(expected, actual) {
		return nil, model.ErrVerificationMissing
	}

	rawBody, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return nil, model.ErrVerificationMissing
	}

	block, err := aesCipher(encryptKey)
	if err != nil {
		return nil, err
	}
	gcm, err := newGCM(block)
	if err != nil {
		return nil, err
	}
	if len(rawBody) < gcm.NonceSize() {
		return nil, model.ErrVerificationMissing
	}
	nonce := rawBody[:gcm.NonceSize()]
	ciphertext := rawBody[gcm.NonceSize():]
	rawEnvelope, err := gcm.Open(nil, nonce, ciphertext, []byte(header))
	if err != nil {
		return nil, model.ErrVerificationMissing
	}

	var envelope struct {
		IssuedAt int64  `json:"iat"`
		Payload  string `json:"payload"`
	}
	if err := json.Unmarshal(rawEnvelope, &envelope); err != nil {
		return nil, model.ErrVerificationMissing
	}
	payload, err := base64.RawURLEncoding.DecodeString(envelope.Payload)
	if err != nil {
		return nil, model.ErrVerificationMissing
	}
	return payload, nil
}

func aesCipher(key []byte) (cipher.Block, error) {
	return aes.NewCipher(key)
}

func newGCM(block cipher.Block) (cipher.AEAD, error) {
	return cipher.NewGCM(block)
}
