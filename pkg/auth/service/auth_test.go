package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-sum/auth"
	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
)

var serviceTestUser = model.User{
	ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	Email:       "ada@example.com",
	DisplayName: "Ada Lovelace",
	Role:        model.RoleUser,
	Verified:    true,
	CreatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
	UpdatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
}

type fakeUserStore struct {
	getByEmailFn  func(context.Context, string) (model.User, error)
	getByIDFn     func(context.Context, uuid.UUID) (model.User, error)
	createFn      func(context.Context, string, string, string, bool) (model.User, error)
	updateEmailFn func(context.Context, uuid.UUID, string) (model.User, error)
}

func (f fakeUserStore) GetByEmail(ctx context.Context, email string) (model.User, error) {
	if f.getByEmailFn != nil {
		return f.getByEmailFn(ctx, email)
	}
	return model.User{}, errors.New("unexpected GetByEmail call")
}

func (f fakeUserStore) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return model.User{}, errors.New("unexpected GetByID call")
}

func (f fakeUserStore) Create(ctx context.Context, email, displayName, role string, verified bool) (model.User, error) {
	if f.createFn != nil {
		return f.createFn(ctx, email, displayName, role, verified)
	}
	return model.User{}, errors.New("unexpected Create call")
}

func (f fakeUserStore) UpdateEmail(ctx context.Context, id uuid.UUID, email string) (model.User, error) {
	if f.updateEmailFn != nil {
		return f.updateEmailFn(ctx, id, email)
	}
	return model.User{}, errors.New("unexpected UpdateEmail call")
}

type capturingNotifier struct {
	deliveries []model.DeliveryInput
	err        error
}

func (n *capturingNotifier) SendVerification(_ context.Context, input model.DeliveryInput) error {
	if n.err != nil {
		return n.err
	}
	n.deliveries = append(n.deliveries, input)
	return nil
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time { return c.now }

func testService(t *testing.T, users fakeUserStore, notifier Notifier, now time.Time) *AuthService {
	t.Helper()
	return NewAuthService(users, Config{
		Method: auth.EmailTOTPMethodConfig{
			Enabled:       true,
			Issuer:        "Forge",
			PeriodSeconds: 300,
		},
		Notifier: notifier,
		TokenCodec: NewEncryptedTokenCodec(
			strings.Repeat("a", 32),
			strings.Repeat("b", 32),
		),
		Clock: fixedClock{now: now},
	})
}

func TestBeginSignupSendsVerificationEmail(t *testing.T) {
	now := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	notifier := &capturingNotifier{}
	svc := testService(t, fakeUserStore{
		getByEmailFn: func(context.Context, string) (model.User, error) {
			return model.User{}, model.ErrUserNotFound
		},
	}, notifier, now)

	flow, err := svc.BeginSignup(context.Background(), model.BeginSignupInput{
		Email:       "ada@example.com",
		DisplayName: "Ada",
	}, "https://example.com/verify")
	if err != nil {
		t.Fatalf("BeginSignup() error = %v", err)
	}
	if flow.Purpose != model.FlowPurposeSignup || flow.Email != "ada@example.com" || flow.DisplayName != "Ada" {
		t.Fatalf("flow = %#v", flow)
	}
	if len(notifier.deliveries) != 1 {
		t.Fatalf("deliveries = %#v", notifier.deliveries)
	}
	code, err := svc.generateCode(flow.Secret, flow.IssuedAt)
	if err != nil {
		t.Fatalf("generateCode() error = %v", err)
	}
	if notifier.deliveries[0].Code != code || !strings.Contains(notifier.deliveries[0].VerifyURL, "token=") {
		t.Fatalf("delivery = %#v", notifier.deliveries[0])
	}
}

func TestBeginSignupRejectsDuplicateEmail(t *testing.T) {
	now := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	svc := testService(t, fakeUserStore{
		getByEmailFn: func(context.Context, string) (model.User, error) {
			return serviceTestUser, nil
		},
	}, &capturingNotifier{}, now)

	_, err := svc.BeginSignup(context.Background(), model.BeginSignupInput{
		Email:       serviceTestUser.Email,
		DisplayName: "Ada",
	}, "https://example.com/verify")
	if !errors.Is(err, model.ErrEmailTaken) {
		t.Fatalf("err = %v", err)
	}
}

func TestSignupAlwaysCreatesRoleUser(t *testing.T) {
	now := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	var capturedRole string
	created := model.User{
		ID:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Email:       "new@example.com",
		DisplayName: "New User",
		Role:        model.RoleUser,
		Verified:    true,
	}
	svc := testService(t, fakeUserStore{
		getByEmailFn: func(_ context.Context, _ string) (model.User, error) {
			return model.User{}, model.ErrUserNotFound
		},
		createFn: func(_ context.Context, _, _, role string, _ bool) (model.User, error) {
			capturedRole = role
			return created, nil
		},
	}, &capturingNotifier{}, now)

	flow, err := svc.BeginSignup(context.Background(), model.BeginSignupInput{
		Email:       "new@example.com",
		DisplayName: "New User",
	}, "https://example.com/verify")
	if err != nil {
		t.Fatalf("BeginSignup() error = %v", err)
	}

	code, err := svc.generateCode(flow.Secret, flow.IssuedAt)
	if err != nil {
		t.Fatalf("generateCode() error = %v", err)
	}

	result, err := svc.VerifyPendingFlow(context.Background(), flow, model.VerifyInput{Code: code})
	if err != nil {
		t.Fatalf("VerifyPendingFlow() error = %v", err)
	}

	if capturedRole != model.RoleUser {
		t.Errorf("users.Create received role = %q, want %q", capturedRole, model.RoleUser)
	}
	if result.User.Role != model.RoleUser {
		t.Errorf("result.User.Role = %q, want %q", result.User.Role, model.RoleUser)
	}
}

func TestBeginSigninSuppressesDeliveryForUnknownEmail(t *testing.T) {
	now := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	notifier := &capturingNotifier{}
	svc := testService(t, fakeUserStore{
		getByEmailFn: func(context.Context, string) (model.User, error) {
			return model.User{}, model.ErrUserNotFound
		},
	}, notifier, now)

	flow, err := svc.BeginSignin(context.Background(), model.BeginSigninInput{
		Email: "missing@example.com",
	}, "https://example.com/verify")
	if err != nil {
		t.Fatalf("BeginSignin() error = %v", err)
	}
	if flow.Purpose != model.FlowPurposeSignin || flow.Email != "missing@example.com" {
		t.Fatalf("flow = %#v", flow)
	}
	if len(notifier.deliveries) != 0 {
		t.Fatalf("deliveries = %#v", notifier.deliveries)
	}
}

func TestVerifyPendingFlowCreatesVerifiedUserOnSignup(t *testing.T) {
	now := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	var createdVerified bool
	svc := testService(t, fakeUserStore{
		createFn: func(_ context.Context, email, displayName, role string, verified bool) (model.User, error) {
			createdVerified = verified
			if email != "ada@example.com" || displayName != "Ada" || role != model.RoleUser {
				t.Fatalf("email=%q displayName=%q role=%q", email, displayName, role)
			}
			return serviceTestUser, nil
		},
	}, &capturingNotifier{}, now)

	flow, code, err := svc.newPendingFlow(model.FlowPurposeSignup, "ada@example.com", "Ada", model.RoleUser, uuid.Nil)
	if err != nil {
		t.Fatalf("newPendingFlow() error = %v", err)
	}
	result, err := svc.VerifyPendingFlow(context.Background(), flow, model.VerifyInput{Code: code})
	if err != nil {
		t.Fatalf("VerifyPendingFlow() error = %v", err)
	}
	if result.User != serviceTestUser || !createdVerified {
		t.Fatalf("result=%#v createdVerified=%v", result, createdVerified)
	}
}

func TestVerifyTokenPrefillsAndCompletesEmailChange(t *testing.T) {
	now := time.Date(2099, 3, 28, 12, 0, 0, 0, time.UTC)
	targetEmail := "new@example.com"
	svc := testService(t, fakeUserStore{
		getByIDFn: func(context.Context, uuid.UUID) (model.User, error) {
			return serviceTestUser, nil
		},
		updateEmailFn: func(_ context.Context, id uuid.UUID, email string) (model.User, error) {
			if id != serviceTestUser.ID || email != targetEmail {
				t.Fatalf("id=%s email=%q", id, email)
			}
			updated := serviceTestUser
			updated.Email = targetEmail
			return updated, nil
		},
	}, &capturingNotifier{}, now)

	flow, code, err := svc.newPendingFlow(model.FlowPurposeEmailChange, targetEmail, "", "", serviceTestUser.ID)
	if err != nil {
		t.Fatalf("newPendingFlow() error = %v", err)
	}
	token, err := svc.tokenCodec.Encode(model.VerificationToken{
		Purpose:   flow.Purpose,
		Email:     flow.Email,
		UserID:    flow.UserID,
		Secret:    flow.Secret,
		IssuedAt:  flow.IssuedAt,
		ExpiresAt: flow.ExpiresAt,
	})
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	state, err := svc.VerifyPageState(token)
	if err != nil {
		t.Fatalf("VerifyPageState() error = %v", err)
	}
	if state.Code != code || state.Email != targetEmail || state.Purpose != model.FlowPurposeEmailChange {
		t.Fatalf("state = %#v", state)
	}

	result, err := svc.VerifyToken(context.Background(), token, model.VerifyInput{Code: code})
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}
	if result.User.Email != targetEmail {
		t.Fatalf("result = %#v", result)
	}
}

func TestVerifyPendingFlowRejectsExpiredCode(t *testing.T) {
	now := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	svc := testService(t, fakeUserStore{}, &capturingNotifier{}, now)
	flow, code, err := svc.newPendingFlow(model.FlowPurposeSignin, "ada@example.com", "", "", uuid.Nil)
	if err != nil {
		t.Fatalf("newPendingFlow() error = %v", err)
	}

	svc.clock = fixedClock{now: flow.ExpiresAt.Add(time.Second)}
	_, err = svc.VerifyPendingFlow(context.Background(), flow, model.VerifyInput{Code: code})
	if !errors.Is(err, model.ErrVerificationExpired) {
		t.Fatalf("err = %v", err)
	}
}

func TestResendPendingFlowStartsFreshCycle(t *testing.T) {
	now := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	notifier := &capturingNotifier{}
	svc := testService(t, fakeUserStore{
		getByEmailFn: func(context.Context, string) (model.User, error) {
			return serviceTestUser, nil
		},
	}, notifier, now)

	flow, _, err := svc.newPendingFlow(model.FlowPurposeSignin, serviceTestUser.Email, "", "", uuid.Nil)
	if err != nil {
		t.Fatalf("newPendingFlow() error = %v", err)
	}
	next, err := svc.ResendPendingFlow(context.Background(), flow, "https://example.com/verify")
	if err != nil {
		t.Fatalf("ResendPendingFlow() error = %v", err)
	}
	if next.Purpose != model.FlowPurposeSignin || next.Email != serviceTestUser.Email || next.Secret == flow.Secret {
		t.Fatalf("next = %#v", next)
	}
	if len(notifier.deliveries) != 1 {
		t.Fatalf("deliveries = %#v", notifier.deliveries)
	}
}
