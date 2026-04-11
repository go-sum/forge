package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-sum/auth"
	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
)

// ── Fake stores ───────────────────────────────────────────────────────────────

type fakePasskeyCredentialStore struct {
	createFn                func(ctx context.Context, cred model.PasskeyCredential) (model.PasskeyCredential, error)
	listByUserIDFn          func(ctx context.Context, userID uuid.UUID) ([]model.PasskeyCredential, error)
	touchPasskeyCredentialFn func(ctx context.Context, id uuid.UUID, signCount int64, cloneWarning bool, lastUsed time.Time) error
	renameFn                func(ctx context.Context, id, userID uuid.UUID, name string) (model.PasskeyCredential, error)
	deleteFn                func(ctx context.Context, id, userID uuid.UUID) error
	getByCredentialIDFn     func(ctx context.Context, credentialID []byte) (model.PasskeyCredential, error)
	getByIDForUserFn        func(ctx context.Context, userID, id uuid.UUID) (model.PasskeyCredential, error)
}

func (f *fakePasskeyCredentialStore) CreateCredential(ctx context.Context, cred model.PasskeyCredential) (model.PasskeyCredential, error) {
	if f.createFn != nil {
		return f.createFn(ctx, cred)
	}
	return cred, nil
}

func (f *fakePasskeyCredentialStore) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.PasskeyCredential, error) {
	if f.listByUserIDFn != nil {
		return f.listByUserIDFn(ctx, userID)
	}
	return []model.PasskeyCredential{}, nil
}

func (f *fakePasskeyCredentialStore) TouchPasskeyCredential(ctx context.Context, id uuid.UUID, signCount int64, cloneWarning bool, lastUsed time.Time) error {
	if f.touchPasskeyCredentialFn != nil {
		return f.touchPasskeyCredentialFn(ctx, id, signCount, cloneWarning, lastUsed)
	}
	return nil
}

func (f *fakePasskeyCredentialStore) RenameCredential(ctx context.Context, id, userID uuid.UUID, name string) (model.PasskeyCredential, error) {
	if f.renameFn != nil {
		return f.renameFn(ctx, id, userID, name)
	}
	return model.PasskeyCredential{ID: id, Name: name}, nil
}

func (f *fakePasskeyCredentialStore) DeleteCredential(ctx context.Context, id, userID uuid.UUID) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id, userID)
	}
	return nil
}

func (f *fakePasskeyCredentialStore) GetByCredentialID(ctx context.Context, credentialID []byte) (model.PasskeyCredential, error) {
	if f.getByCredentialIDFn != nil {
		return f.getByCredentialIDFn(ctx, credentialID)
	}
	return model.PasskeyCredential{}, model.ErrPasskeyNotFound
}

func (f *fakePasskeyCredentialStore) GetByIDForUser(ctx context.Context, userID, id uuid.UUID) (model.PasskeyCredential, error) {
	if f.getByIDForUserFn != nil {
		return f.getByIDForUserFn(ctx, userID, id)
	}
	return model.PasskeyCredential{}, model.ErrPasskeyNotFound
}

// fakeUserStoreWithWebAuthn extends fakeUserStore with WebAuthn tracking.
type fakeUserStoreWithWebAuthn struct {
	fakeUserStore
	setWebAuthnIDIfNullCalled int
	setWebAuthnIDIfNullFn     func(ctx context.Context, id uuid.UUID, webauthnID []byte) (model.User, error)
}

func (f *fakeUserStoreWithWebAuthn) SetWebAuthnID(ctx context.Context, id uuid.UUID, webauthnID []byte) (model.User, error) {
	u := serviceTestUser
	u.WebAuthnID = webauthnID
	return u, nil
}

func (f *fakeUserStoreWithWebAuthn) SetWebAuthnIDIfNull(ctx context.Context, id uuid.UUID, webauthnID []byte) (model.User, error) {
	f.setWebAuthnIDIfNullCalled++
	if f.setWebAuthnIDIfNullFn != nil {
		return f.setWebAuthnIDIfNullFn(ctx, id, webauthnID)
	}
	u := serviceTestUser
	u.WebAuthnID = webauthnID
	return u, nil
}

func (f *fakeUserStoreWithWebAuthn) GetByWebAuthnID(_ context.Context, _ []byte) (model.User, error) {
	return model.User{}, model.ErrUserNotFound
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func testPasskeyService(t *testing.T, users *fakeUserStoreWithWebAuthn, creds *fakePasskeyCredentialStore) *PasskeyServiceImpl {
	t.Helper()
	cfg := auth.PasskeyMethodConfig{
		Enabled:       true,
		RPDisplayName: "Test App",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost"},
	}
	svc, err := NewPasskeyService(users, creds, cfg)
	if err != nil {
		t.Fatalf("NewPasskeyService() error = %v", err)
	}
	return svc
}

// ── BeginRegistration tests ───────────────────────────────────────────────────

func TestBeginRegistration_GeneratesWebAuthnIDWhenUserHasNone(t *testing.T) {
	userWithoutHandle := serviceTestUser
	userWithoutHandle.WebAuthnID = nil

	users := fakeUserStoreWithWebAuthn{
		fakeUserStore: fakeUserStore{
			getByIDFn: func(_ context.Context, id uuid.UUID) (model.User, error) {
				return userWithoutHandle, nil
			},
		},
	}
	users.setWebAuthnIDIfNullFn = func(_ context.Context, _ uuid.UUID, handle []byte) (model.User, error) {
		if len(handle) == 0 {
			t.Fatal("SetWebAuthnIDIfNull called with empty handle")
		}
		u := userWithoutHandle
		u.WebAuthnID = handle
		return u, nil
	}
	creds := &fakePasskeyCredentialStore{}
	svc := testPasskeyService(t, &users, creds)

	_, _, err := svc.BeginRegistration(context.Background(), serviceTestUser.ID)
	if err != nil {
		t.Fatalf("BeginRegistration() error = %v", err)
	}
	if users.setWebAuthnIDIfNullCalled != 1 {
		t.Fatalf("SetWebAuthnIDIfNull called %d times, want 1", users.setWebAuthnIDIfNullCalled)
	}
}

func TestBeginRegistration_DoesNotGenerateWebAuthnIDWhenUserAlreadyHasOne(t *testing.T) {
	userWithHandle := serviceTestUser
	userWithHandle.WebAuthnID = []byte("existing-handle-12345678901234567890123456789012345678901234")

	users := fakeUserStoreWithWebAuthn{
		fakeUserStore: fakeUserStore{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (model.User, error) {
				return userWithHandle, nil
			},
		},
	}
	creds := &fakePasskeyCredentialStore{}
	svc := testPasskeyService(t, &users, creds)

	_, _, err := svc.BeginRegistration(context.Background(), serviceTestUser.ID)
	if err != nil {
		t.Fatalf("BeginRegistration() error = %v", err)
	}
	if users.setWebAuthnIDIfNullCalled != 0 {
		t.Fatalf("SetWebAuthnIDIfNull called %d times, want 0", users.setWebAuthnIDIfNullCalled)
	}
}

func TestBeginRegistration_PropagatesUserStoreError(t *testing.T) {
	storeErr := errors.New("db connection failed")
	users := fakeUserStoreWithWebAuthn{
		fakeUserStore: fakeUserStore{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (model.User, error) {
				return model.User{}, storeErr
			},
		},
	}
	creds := &fakePasskeyCredentialStore{}
	svc := testPasskeyService(t, &users, creds)

	_, _, err := svc.BeginRegistration(context.Background(), serviceTestUser.ID)
	if !errors.Is(err, storeErr) {
		t.Fatalf("err = %v, want to contain %v", err, storeErr)
	}
}

func TestBeginRegistration_PropagatesSetWebAuthnIDError(t *testing.T) {
	userWithoutHandle := serviceTestUser
	userWithoutHandle.WebAuthnID = nil
	storeErr := errors.New("db write failed")

	users := fakeUserStoreWithWebAuthn{
		fakeUserStore: fakeUserStore{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (model.User, error) {
				return userWithoutHandle, nil
			},
		},
		setWebAuthnIDIfNullFn: func(_ context.Context, _ uuid.UUID, _ []byte) (model.User, error) {
			return model.User{}, storeErr
		},
	}
	creds := &fakePasskeyCredentialStore{}
	svc := testPasskeyService(t, &users, creds)

	_, _, err := svc.BeginRegistration(context.Background(), serviceTestUser.ID)
	if !errors.Is(err, storeErr) {
		t.Fatalf("err = %v, want to contain %v", err, storeErr)
	}
}

// ── GetPasskey tests ──────────────────────────────────────────────────────────

func TestGetPasskey_ReturnsNotFoundWhenCredentialNotFound(t *testing.T) {
	unknownID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	users := fakeUserStoreWithWebAuthn{}
	creds := &fakePasskeyCredentialStore{
		getByIDForUserFn: func(_ context.Context, _, _ uuid.UUID) (model.PasskeyCredential, error) {
			return model.PasskeyCredential{}, model.ErrPasskeyNotFound
		},
	}
	svc := testPasskeyService(t, &users, creds)

	_, err := svc.GetPasskey(context.Background(), serviceTestUser.ID, unknownID)
	if !errors.Is(err, model.ErrPasskeyNotFound) {
		t.Fatalf("err = %v, want ErrPasskeyNotFound", err)
	}
}

func TestGetPasskey_ReturnsCredentialWhenFound(t *testing.T) {
	credID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	want := model.PasskeyCredential{
		ID:     credID,
		UserID: serviceTestUser.ID,
		Name:   "Face ID",
	}

	users := fakeUserStoreWithWebAuthn{}
	creds := &fakePasskeyCredentialStore{
		getByIDForUserFn: func(_ context.Context, userID, id uuid.UUID) (model.PasskeyCredential, error) {
			if userID != serviceTestUser.ID {
				t.Fatalf("GetByIDForUser called with wrong userID: %s", userID)
			}
			if id != credID {
				t.Fatalf("GetByIDForUser called with wrong id: %s", id)
			}
			return want, nil
		},
	}
	svc := testPasskeyService(t, &users, creds)

	got, err := svc.GetPasskey(context.Background(), serviceTestUser.ID, credID)
	if err != nil {
		t.Fatalf("GetPasskey() error = %v", err)
	}
	if got.ID != want.ID || got.Name != want.Name {
		t.Fatalf("got = %#v, want %#v", got, want)
	}
}

func TestGetPasskey_PropagatesRepositoryError(t *testing.T) {
	storeErr := errors.New("db read failed")

	users := fakeUserStoreWithWebAuthn{}
	creds := &fakePasskeyCredentialStore{
		getByIDForUserFn: func(_ context.Context, _, _ uuid.UUID) (model.PasskeyCredential, error) {
			return model.PasskeyCredential{}, storeErr
		},
	}
	svc := testPasskeyService(t, &users, creds)

	_, err := svc.GetPasskey(context.Background(), serviceTestUser.ID, uuid.New())
	if !errors.Is(err, storeErr) {
		t.Fatalf("err = %v, want to contain %v", err, storeErr)
	}
}

// ── RenamePasskey tests ───────────────────────────────────────────────────────

func TestRenamePasskey_ReturnsNotFoundWhenCredentialStoreReturnsIt(t *testing.T) {
	users := fakeUserStoreWithWebAuthn{}
	creds := &fakePasskeyCredentialStore{
		renameFn: func(_ context.Context, _, _ uuid.UUID, _ string) (model.PasskeyCredential, error) {
			return model.PasskeyCredential{}, model.ErrPasskeyNotFound
		},
	}
	svc := testPasskeyService(t, &users, creds)

	_, err := svc.RenamePasskey(context.Background(), serviceTestUser.ID, uuid.New(), "New Name")
	if !errors.Is(err, model.ErrPasskeyNotFound) {
		t.Fatalf("err = %v, want ErrPasskeyNotFound", err)
	}
}

func TestRenamePasskey_PassesCorrectArgsToCredentialStore(t *testing.T) {
	passkeyID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	newName := "My Security Key"

	var capturedID, capturedUserID uuid.UUID
	var capturedName string

	users := fakeUserStoreWithWebAuthn{}
	creds := &fakePasskeyCredentialStore{
		renameFn: func(_ context.Context, id, userID uuid.UUID, name string) (model.PasskeyCredential, error) {
			capturedID = id
			capturedUserID = userID
			capturedName = name
			return model.PasskeyCredential{ID: id, UserID: userID, Name: name}, nil
		},
	}
	svc := testPasskeyService(t, &users, creds)

	got, err := svc.RenamePasskey(context.Background(), serviceTestUser.ID, passkeyID, newName)
	if err != nil {
		t.Fatalf("RenamePasskey() error = %v", err)
	}
	if capturedID != passkeyID {
		t.Fatalf("capturedID = %s, want %s", capturedID, passkeyID)
	}
	if capturedUserID != serviceTestUser.ID {
		t.Fatalf("capturedUserID = %s, want %s", capturedUserID, serviceTestUser.ID)
	}
	if capturedName != newName {
		t.Fatalf("capturedName = %q, want %q", capturedName, newName)
	}
	if got.Name != newName {
		t.Fatalf("got.Name = %q, want %q", got.Name, newName)
	}
}

func TestRenamePasskey_ReturnsRenamedCredential(t *testing.T) {
	passkeyID := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	want := model.PasskeyCredential{
		ID:     passkeyID,
		UserID: serviceTestUser.ID,
		Name:   "Renamed Key",
	}

	users := fakeUserStoreWithWebAuthn{}
	creds := &fakePasskeyCredentialStore{
		renameFn: func(_ context.Context, _, _ uuid.UUID, _ string) (model.PasskeyCredential, error) {
			return want, nil
		},
	}
	svc := testPasskeyService(t, &users, creds)

	got, err := svc.RenamePasskey(context.Background(), serviceTestUser.ID, passkeyID, "Renamed Key")
	if err != nil {
		t.Fatalf("RenamePasskey() error = %v", err)
	}
	if got.ID != want.ID || got.Name != want.Name || got.UserID != want.UserID {
		t.Fatalf("got = %#v, want %#v", got, want)
	}
}

// ── DeletePasskey tests ───────────────────────────────────────────────────────

func TestDeletePasskey_ReturnsNotFoundWhenCredentialStoreReturnsIt(t *testing.T) {
	users := fakeUserStoreWithWebAuthn{}
	creds := &fakePasskeyCredentialStore{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return model.ErrPasskeyNotFound
		},
	}
	svc := testPasskeyService(t, &users, creds)

	err := svc.DeletePasskey(context.Background(), serviceTestUser.ID, uuid.New())
	if !errors.Is(err, model.ErrPasskeyNotFound) {
		t.Fatalf("err = %v, want ErrPasskeyNotFound", err)
	}
}

func TestDeletePasskey_PassesCorrectArgsToCredentialStore(t *testing.T) {
	passkeyID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")

	var capturedID, capturedUserID uuid.UUID
	users := fakeUserStoreWithWebAuthn{}
	creds := &fakePasskeyCredentialStore{
		deleteFn: func(_ context.Context, id, userID uuid.UUID) error {
			capturedID = id
			capturedUserID = userID
			return nil
		},
	}
	svc := testPasskeyService(t, &users, creds)

	err := svc.DeletePasskey(context.Background(), serviceTestUser.ID, passkeyID)
	if err != nil {
		t.Fatalf("DeletePasskey() error = %v", err)
	}
	if capturedID != passkeyID {
		t.Fatalf("capturedID = %s, want %s", capturedID, passkeyID)
	}
	if capturedUserID != serviceTestUser.ID {
		t.Fatalf("capturedUserID = %s, want %s", capturedUserID, serviceTestUser.ID)
	}
}

// ── ListPasskeys tests ────────────────────────────────────────────────────────

func TestListPasskeys_ReturnsEmptySliceWhenUserHasNoCredentials(t *testing.T) {
	users := fakeUserStoreWithWebAuthn{}
	creds := &fakePasskeyCredentialStore{
		listByUserIDFn: func(_ context.Context, _ uuid.UUID) ([]model.PasskeyCredential, error) {
			return []model.PasskeyCredential{}, nil
		},
	}
	svc := testPasskeyService(t, &users, creds)

	result, err := svc.ListPasskeys(context.Background(), serviceTestUser.ID)
	if err != nil {
		t.Fatalf("ListPasskeys() error = %v", err)
	}
	if result == nil {
		t.Fatal("ListPasskeys() returned nil, want empty slice")
	}
	if len(result) != 0 {
		t.Fatalf("len(result) = %d, want 0", len(result))
	}
}

func TestListPasskeys_ReturnsAllCredentialsForUser(t *testing.T) {
	cred1 := model.PasskeyCredential{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Name: "Key 1"}
	cred2 := model.PasskeyCredential{ID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Name: "Key 2"}

	users := fakeUserStoreWithWebAuthn{}
	creds := &fakePasskeyCredentialStore{
		listByUserIDFn: func(_ context.Context, userID uuid.UUID) ([]model.PasskeyCredential, error) {
			if userID != serviceTestUser.ID {
				t.Fatalf("ListByUserID called with wrong userID: %s", userID)
			}
			return []model.PasskeyCredential{cred1, cred2}, nil
		},
	}
	svc := testPasskeyService(t, &users, creds)

	result, err := svc.ListPasskeys(context.Background(), serviceTestUser.ID)
	if err != nil {
		t.Fatalf("ListPasskeys() error = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].ID != cred1.ID || result[1].ID != cred2.ID {
		t.Fatalf("result = %#v", result)
	}
}

func TestListPasskeys_PropagatesListByUserIDError(t *testing.T) {
	storeErr := errors.New("db read failed")

	users := fakeUserStoreWithWebAuthn{}
	creds := &fakePasskeyCredentialStore{
		listByUserIDFn: func(_ context.Context, _ uuid.UUID) ([]model.PasskeyCredential, error) {
			return nil, storeErr
		},
	}
	svc := testPasskeyService(t, &users, creds)

	_, err := svc.ListPasskeys(context.Background(), serviceTestUser.ID)
	if !errors.Is(err, storeErr) {
		t.Fatalf("err = %v, want to contain %v", err, storeErr)
	}
}
