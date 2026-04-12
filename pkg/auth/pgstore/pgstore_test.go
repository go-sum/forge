package pgstore_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-sum/auth/model"
	"github.com/go-sum/auth/pgstore"
	"github.com/go-sum/auth/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testPool connects to the test database. Tests are skipped when
// TEST_DATABASE_URL is not set (unit test runs without a real DB).
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// truncateUsers removes all rows from users between tests.
func truncateUsers(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "TRUNCATE users CASCADE")
	})
}

func newStore(t *testing.T) *pgstore.PgStore {
	t.Helper()
	pool := testPool(t)
	s := pgstore.New(pgstore.Config{Pool: pool})
	requireUsersTable(t, pool)
	truncateUsers(t, pool)
	return s
}

func requireUsersTable(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	var relation string
	err := pool.QueryRow(context.Background(), `
SELECT COALESCE(to_regclass('public.users')::text, '')
`).Scan(&relation)
	if err != nil {
		t.Fatalf("verify users table: %v", err)
	}
	if relation != "users" {
		t.Fatalf("users table is missing; run task db:migrate before these tests")
	}
}

// Compile-time check: PgStore satisfies repository.UserStore.
var _ repository.UserStore = (*pgstore.PgStore)(nil)

func TestCreate_ReturnsUser(t *testing.T) {
	s := newStore(t)
	u, err := s.Create(context.Background(), "ada@example.com", "Ada Lovelace", "user", true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.ID == (uuid.UUID{}) {
		t.Fatal("Create: ID is zero")
	}
	if u.Email != "ada@example.com" {
		t.Errorf("Email = %q, want %q", u.Email, "ada@example.com")
	}
	if u.DisplayName != "Ada Lovelace" {
		t.Errorf("DisplayName = %q, want %q", u.DisplayName, "Ada Lovelace")
	}
	if u.Role != "user" {
		t.Errorf("Role = %q, want %q", u.Role, "user")
	}
	if !u.Verified {
		t.Error("Verified = false, want true")
	}
	if u.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestCreate_DuplicateEmail_ReturnsErrEmailTaken(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	if _, err := s.Create(ctx, "ada@example.com", "Ada", "user", true); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	_, err := s.Create(ctx, "ada@example.com", "Ada2", "user", true)
	if !errors.Is(err, model.ErrEmailTaken) {
		t.Fatalf("second Create: err = %v, want ErrEmailTaken", err)
	}
}

func TestCreate_CaseInsensitiveDuplicate_ReturnsErrEmailTaken(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	if _, err := s.Create(ctx, "ada@example.com", "Ada", "user", true); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	_, err := s.Create(ctx, "ADA@EXAMPLE.COM", "Ada Upper", "user", true)
	if !errors.Is(err, model.ErrEmailTaken) {
		t.Fatalf("uppercase duplicate: err = %v, want ErrEmailTaken", err)
	}
}

func TestGetByID_Roundtrip(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	created, err := s.Create(ctx, "ada@example.com", "Ada Lovelace", "user", true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := s.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !reflect.DeepEqual(got, created) {
		t.Errorf("GetByID = %+v, want %+v", got, created)
	}
}

func TestGetByID_NotFound_ReturnsErrUserNotFound(t *testing.T) {
	s := newStore(t)
	_, err := s.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, model.ErrUserNotFound) {
		t.Fatalf("GetByID: err = %v, want ErrUserNotFound", err)
	}
}

func TestGetByEmail_Roundtrip(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	created, err := s.Create(ctx, "ada@example.com", "Ada Lovelace", "user", true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := s.GetByEmail(ctx, "ada@example.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if !reflect.DeepEqual(got, created) {
		t.Errorf("GetByEmail = %+v, want %+v", got, created)
	}
}

func TestGetByEmail_CaseInsensitive(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	if _, err := s.Create(ctx, "ada@example.com", "Ada", "user", true); err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, err := s.GetByEmail(ctx, "ADA@EXAMPLE.COM")
	if err != nil {
		t.Fatalf("GetByEmail (uppercase): %v", err)
	}
}

func TestGetByEmail_NotFound_ReturnsErrUserNotFound(t *testing.T) {
	s := newStore(t)
	_, err := s.GetByEmail(context.Background(), "nobody@example.com")
	if !errors.Is(err, model.ErrUserNotFound) {
		t.Fatalf("GetByEmail: err = %v, want ErrUserNotFound", err)
	}
}

func TestUpdateEmail_ReturnsUpdatedUser(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	created, err := s.Create(ctx, "ada@example.com", "Ada Lovelace", "user", true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	updated, err := s.UpdateEmail(ctx, created.ID, "ada2@example.com")
	if err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}
	if updated.Email != "ada2@example.com" {
		t.Errorf("Email = %q, want %q", updated.Email, "ada2@example.com")
	}
	if updated.ID != created.ID {
		t.Errorf("ID changed: got %s, want %s", updated.ID, created.ID)
	}
}

func TestUpdateEmail_NotFound_ReturnsErrUserNotFound(t *testing.T) {
	s := newStore(t)
	_, err := s.UpdateEmail(context.Background(), uuid.New(), "new@example.com")
	if !errors.Is(err, model.ErrUserNotFound) {
		t.Fatalf("UpdateEmail: err = %v, want ErrUserNotFound", err)
	}
}

func TestUpdateEmail_Duplicate_ReturnsErrEmailTaken(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	if _, err := s.Create(ctx, "ada@example.com", "Ada", "user", true); err != nil {
		t.Fatalf("Create ada: %v", err)
	}
	grace, err := s.Create(ctx, "grace@example.com", "Grace", "user", true)
	if err != nil {
		t.Fatalf("Create grace: %v", err)
	}
	_, err = s.UpdateEmail(ctx, grace.ID, "ada@example.com")
	if !errors.Is(err, model.ErrEmailTaken) {
		t.Fatalf("UpdateEmail duplicate: err = %v, want ErrEmailTaken", err)
	}
}

func TestGetPasskeyCredentialByIDForUser(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	// Create a user to own the credential.
	user, err := s.Create(ctx, "passkey@example.com", "Passkey User", "user", true)
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}

	// Set a WebAuthn ID.
	user, err = s.SetWebAuthnID(ctx, user.ID, []byte("webauthn-handle-12345678901234567890123456789012345678901234"))
	if err != nil {
		t.Fatalf("SetWebAuthnID: %v", err)
	}

	// Insert a credential.
	cred := model.PasskeyCredential{
		UserID:          user.ID,
		CredentialID:    []byte("cred-id-bytes"),
		Name:            "My Key",
		PublicKey:       []byte("public-key-bytes"),
		PublicKeyAlg:    -7,
		AttestationType: "none",
		AAGUID:          make([]byte, 16),
		Transports:      []string{"internal"},
	}
	created, err := s.CreateCredential(ctx, cred)
	if err != nil {
		t.Fatalf("CreateCredential: %v", err)
	}

	// Happy path: correct user and ID.
	got, err := s.GetByIDForUser(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetByIDForUser (match): %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %s, want %s", got.ID, created.ID)
	}
	if got.Name != "My Key" {
		t.Errorf("Name = %q, want %q", got.Name, "My Key")
	}

	// Wrong user: different user should not see this credential.
	otherUser, err := s.Create(ctx, "other@example.com", "Other", "user", true)
	if err != nil {
		t.Fatalf("Create other user: %v", err)
	}
	_, err = s.GetByIDForUser(ctx, otherUser.ID, created.ID)
	if !errors.Is(err, model.ErrPasskeyNotFound) {
		t.Errorf("GetByIDForUser (wrong user): err = %v, want ErrPasskeyNotFound", err)
	}

	// Wrong ID: non-existent credential.
	_, err = s.GetByIDForUser(ctx, user.ID, uuid.New())
	if !errors.Is(err, model.ErrPasskeyNotFound) {
		t.Errorf("GetByIDForUser (wrong id): err = %v, want ErrPasskeyNotFound", err)
	}
}

// ── Additional passkey persistence tests (T3-3) ───────────────────────────────

// createTestUser is a helper that creates a user with a WebAuthn ID and fails the test if it errors.
func createTestUser(t *testing.T, s *pgstore.PgStore, email string) model.User {
	t.Helper()
	ctx := context.Background()
	user, err := s.Create(ctx, email, "Test User", "user", true)
	if err != nil {
		t.Fatalf("createTestUser Create: %v", err)
	}
	webAuthnID := make([]byte, 64)
	copy(webAuthnID, []byte(email))
	user, err = s.SetWebAuthnID(ctx, user.ID, webAuthnID)
	if err != nil {
		t.Fatalf("createTestUser SetWebAuthnID: %v", err)
	}
	return user
}

// insertCred is a helper that inserts a credential and fails if it errors.
func insertCred(t *testing.T, s *pgstore.PgStore, cred model.PasskeyCredential) model.PasskeyCredential {
	t.Helper()
	created, err := s.CreateCredential(context.Background(), cred)
	if err != nil {
		t.Fatalf("insertCred: %v", err)
	}
	return created
}

func TestCreatePasskeyCredential_DuplicateCredentialID_ReturnsAlreadyRegistered(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	user := createTestUser(t, s, "dupkey@example.com")

	cred := model.PasskeyCredential{
		UserID:          user.ID,
		CredentialID:    []byte("unique-cred-id-001"),
		Name:            "Key 1",
		PublicKey:       []byte("pk-bytes"),
		PublicKeyAlg:    -7,
		AttestationType: "none",
		AAGUID:          make([]byte, 16),
		Transports:      []string{"usb"},
	}

	// First insert should succeed.
	if _, err := s.CreateCredential(ctx, cred); err != nil {
		t.Fatalf("first CreateCredential: %v", err)
	}

	// Second insert with the same CredentialID should return ErrPasskeyAlreadyRegistered.
	cred.Name = "Key 1 Duplicate"
	_, err := s.CreateCredential(ctx, cred)
	if !errors.Is(err, model.ErrPasskeyAlreadyRegistered) {
		t.Fatalf("duplicate CredentialID: err = %v, want ErrPasskeyAlreadyRegistered", err)
	}
}

func TestCreatePasskeyCredential_FullFieldRoundtrip(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	user := createTestUser(t, s, "roundtrip@example.com")

	now := time.Now().UTC().Truncate(time.Millisecond)
	aaguid := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	cred := model.PasskeyCredential{
		UserID:          user.ID,
		CredentialID:    []byte("roundtrip-cred-id-xyz"),
		Name:            "Full Field Key",
		PublicKey:       []byte("realistic-public-key-bytes-32b-x"),
		PublicKeyAlg:    -7,
		AttestationType: "none",
		AAGUID:          aaguid,
		SignCount:       42,
		CloneWarning:    false,
		BackupEligible:  true,
		BackupState:     true,
		Transports:      []string{"internal", "usb"},
		Attachment:      "platform",
		LastUsedAt:      &now,
	}

	created, err := s.CreateCredential(ctx, cred)
	if err != nil {
		t.Fatalf("CreateCredential: %v", err)
	}

	got, err := s.GetByIDForUser(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetByIDForUser: %v", err)
	}

	if got.UserID != cred.UserID {
		t.Errorf("UserID = %s, want %s", got.UserID, cred.UserID)
	}
	if !bytes.Equal(got.CredentialID, cred.CredentialID) {
		t.Errorf("CredentialID = %x, want %x", got.CredentialID, cred.CredentialID)
	}
	if got.Name != cred.Name {
		t.Errorf("Name = %q, want %q", got.Name, cred.Name)
	}
	if !bytes.Equal(got.PublicKey, cred.PublicKey) {
		t.Errorf("PublicKey = %x, want %x", got.PublicKey, cred.PublicKey)
	}
	if got.PublicKeyAlg != cred.PublicKeyAlg {
		t.Errorf("PublicKeyAlg = %d, want %d", got.PublicKeyAlg, cred.PublicKeyAlg)
	}
	if got.AttestationType != cred.AttestationType {
		t.Errorf("AttestationType = %q, want %q", got.AttestationType, cred.AttestationType)
	}
	if !bytes.Equal(got.AAGUID, cred.AAGUID) {
		t.Errorf("AAGUID = %x, want %x", got.AAGUID, cred.AAGUID)
	}
	if got.SignCount != cred.SignCount {
		t.Errorf("SignCount = %d, want %d", got.SignCount, cred.SignCount)
	}
	if got.CloneWarning != cred.CloneWarning {
		t.Errorf("CloneWarning = %v, want %v", got.CloneWarning, cred.CloneWarning)
	}
	if got.BackupEligible != cred.BackupEligible {
		t.Errorf("BackupEligible = %v, want %v", got.BackupEligible, cred.BackupEligible)
	}
	if got.BackupState != cred.BackupState {
		t.Errorf("BackupState = %v, want %v", got.BackupState, cred.BackupState)
	}
	if !reflect.DeepEqual(got.Transports, cred.Transports) {
		t.Errorf("Transports = %v, want %v", got.Transports, cred.Transports)
	}
	if got.Attachment != cred.Attachment {
		t.Errorf("Attachment = %q, want %q", got.Attachment, cred.Attachment)
	}
	if got.LastUsedAt == nil {
		t.Error("LastUsedAt is nil, want non-nil")
	} else if !got.LastUsedAt.Truncate(time.Millisecond).Equal(now) {
		t.Errorf("LastUsedAt = %v, want %v", got.LastUsedAt, now)
	}
}

func TestListPasskeyCredentials_OrderingAndIsolation(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	userA := createTestUser(t, s, "list-user-a@example.com")
	userB := createTestUser(t, s, "list-user-b@example.com")

	baseCredA := model.PasskeyCredential{
		UserID:          userA.ID,
		PublicKey:       []byte("pk"),
		AttestationType: "none",
		AAGUID:          make([]byte, 16),
		Transports:      []string{},
	}
	baseCredB := model.PasskeyCredential{
		UserID:          userB.ID,
		PublicKey:       []byte("pk"),
		AttestationType: "none",
		AAGUID:          make([]byte, 16),
		Transports:      []string{},
	}

	// Insert 3 credentials for user A, giving each a distinct CredentialID and name.
	for i := 0; i < 3; i++ {
		c := baseCredA
		c.CredentialID = []byte{byte('a'), byte(i), 0, 1, 2, 3, 4, 5}
		c.Name = "A Key " + string(rune('1'+i))
		insertCred(t, s, c)
	}

	// Insert 2 credentials for user B.
	for i := 0; i < 2; i++ {
		c := baseCredB
		c.CredentialID = []byte{byte('b'), byte(i), 0, 1, 2, 3, 4, 5}
		c.Name = "B Key " + string(rune('1'+i))
		insertCred(t, s, c)
	}

	credsA, err := s.ListByUserID(ctx, userA.ID)
	if err != nil {
		t.Fatalf("ListByUserID(A): %v", err)
	}
	if len(credsA) != 3 {
		t.Errorf("ListByUserID(A) count = %d, want 3", len(credsA))
	}
	// Verify all returned credentials belong to user A.
	for _, c := range credsA {
		if c.UserID != userA.ID {
			t.Errorf("credential UserID = %s, want %s (cross-user leakage)", c.UserID, userA.ID)
		}
	}

	credsB, err := s.ListByUserID(ctx, userB.ID)
	if err != nil {
		t.Fatalf("ListByUserID(B): %v", err)
	}
	if len(credsB) != 2 {
		t.Errorf("ListByUserID(B) count = %d, want 2", len(credsB))
	}
	// Verify all returned credentials belong to user B.
	for _, c := range credsB {
		if c.UserID != userB.ID {
			t.Errorf("credential UserID = %s, want %s (cross-user leakage)", c.UserID, userB.ID)
		}
	}

	// Ordering: most-recent first. CreatedAt should be non-decreasing from last to first.
	for i := 1; i < len(credsA); i++ {
		if credsA[i-1].CreatedAt.Before(credsA[i].CreatedAt) {
			t.Errorf("credsA[%d].CreatedAt = %v < credsA[%d].CreatedAt = %v; expected most-recent first",
				i-1, credsA[i-1].CreatedAt, i, credsA[i].CreatedAt)
		}
	}
}

func TestDeletePasskeyCredential_CrossUser_ReturnsNotFound(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	userA := createTestUser(t, s, "delete-a@example.com")
	userB := createTestUser(t, s, "delete-b@example.com")

	credA := insertCred(t, s, model.PasskeyCredential{
		UserID:          userA.ID,
		CredentialID:    []byte("delete-test-cred-a"),
		Name:            "A Key",
		PublicKey:       []byte("pk"),
		AttestationType: "none",
		AAGUID:          make([]byte, 16),
		Transports:      []string{},
	})
	credB := insertCred(t, s, model.PasskeyCredential{
		UserID:          userB.ID,
		CredentialID:    []byte("delete-test-cred-b"),
		Name:            "B Key",
		PublicKey:       []byte("pk"),
		AttestationType: "none",
		AAGUID:          make([]byte, 16),
		Transports:      []string{},
	})

	// User A attempts to delete user B's credential — must return
	// ErrPasskeyNotFound and leave both rows intact.
	err := s.DeleteCredential(ctx, credB.ID, userA.ID)
	if !errors.Is(err, model.ErrPasskeyNotFound) {
		t.Fatalf("DeleteCredential (cross-user) err = %v, want ErrPasskeyNotFound", err)
	}

	// User B's credential must still exist.
	got, err := s.GetByIDForUser(ctx, userB.ID, credB.ID)
	if err != nil {
		t.Fatalf("GetByIDForUser after cross-user delete: %v", err)
	}
	if got.ID != credB.ID {
		t.Errorf("got.ID = %s, want %s", got.ID, credB.ID)
	}

	// User A's credential must also still exist.
	if _, err := s.GetByIDForUser(ctx, userA.ID, credA.ID); err != nil {
		t.Fatalf("GetByIDForUser user A after cross-user delete: %v", err)
	}
}

func TestDeletePasskeyCredential_NotFoundID_ReturnsNotFound(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	user := createTestUser(t, s, "delete-missing@example.com")

	err := s.DeleteCredential(ctx, uuid.New(), user.ID)
	if !errors.Is(err, model.ErrPasskeyNotFound) {
		t.Fatalf("DeleteCredential (unknown id) err = %v, want ErrPasskeyNotFound", err)
	}
}

func TestDeletePasskeyCredential_OwnedRow_Deletes(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	user := createTestUser(t, s, "delete-owner@example.com")
	cred := insertCred(t, s, model.PasskeyCredential{
		UserID:          user.ID,
		CredentialID:    []byte("delete-owner-cred"),
		Name:            "Owner Key",
		PublicKey:       []byte("pk"),
		AttestationType: "none",
		AAGUID:          make([]byte, 16),
		Transports:      []string{},
	})

	if err := s.DeleteCredential(ctx, cred.ID, user.ID); err != nil {
		t.Fatalf("DeleteCredential (owned row): %v", err)
	}

	if _, err := s.GetByIDForUser(ctx, user.ID, cred.ID); !errors.Is(err, model.ErrPasskeyNotFound) {
		t.Fatalf("GetByIDForUser after delete err = %v, want ErrPasskeyNotFound", err)
	}
}

func TestTouchPasskeyCredential_MonotonicCounter(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	user := createTestUser(t, s, "touch@example.com")

	cred := insertCred(t, s, model.PasskeyCredential{
		UserID:          user.ID,
		CredentialID:    []byte("touch-test-cred"),
		Name:            "Touch Key",
		PublicKey:       []byte("pk"),
		AttestationType: "none",
		AAGUID:          make([]byte, 16),
		SignCount:       10,
		Transports:      []string{},
	})

	now := time.Now().UTC()

	// Touch with a regressed sign count (5 < 10): the DB should keep 10.
	if err := s.TouchPasskeyCredential(ctx, cred.ID, 5, true, now); err != nil {
		t.Fatalf("TouchPasskeyCredential (regressed count): %v", err)
	}

	got, err := s.GetByIDForUser(ctx, user.ID, cred.ID)
	if err != nil {
		t.Fatalf("GetByIDForUser after first touch: %v", err)
	}
	if got.SignCount != 10 {
		t.Errorf("SignCount after regressed touch = %d, want 10 (monotonic counter)", got.SignCount)
	}
	if !got.CloneWarning {
		t.Error("CloneWarning = false after touch with clone_warning=true")
	}

	// Touch with an advancing sign count (15 > 10): the DB should update to 15.
	if err := s.TouchPasskeyCredential(ctx, cred.ID, 15, false, now); err != nil {
		t.Fatalf("TouchPasskeyCredential (advancing count): %v", err)
	}

	got2, err := s.GetByIDForUser(ctx, user.ID, cred.ID)
	if err != nil {
		t.Fatalf("GetByIDForUser after second touch: %v", err)
	}
	if got2.SignCount != 15 {
		t.Errorf("SignCount after advancing touch = %d, want 15", got2.SignCount)
	}
}

func TestSetUserWebAuthnIDIfNull_Idempotent(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	// Create user without a WebAuthn ID.
	user, err := s.Create(ctx, "webauthn-idempotent@example.com", "Idempotent User", "user", true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	firstID := []byte("first-webauthn-id-32bytespadded!!")

	// Set WebAuthn ID for the first time — should succeed.
	updated, err := s.SetWebAuthnIDIfNull(ctx, user.ID, firstID)
	if err != nil {
		t.Fatalf("first SetWebAuthnIDIfNull: %v", err)
	}
	if !bytes.Equal(updated.WebAuthnID, firstID) {
		t.Errorf("WebAuthnID = %x, want %x", updated.WebAuthnID, firstID)
	}

	// Set a different WebAuthn ID — the DB should preserve the first one (race-resistant).
	secondID := []byte("second-webauthn-id-32bytespadded!")
	updated2, err := s.SetWebAuthnIDIfNull(ctx, user.ID, secondID)
	// The implementation may return an error (ErrWebAuthnIDAlreadySet) or silently ignore it.
	// Either way, the stored ID must be the original one.
	if err != nil {
		// If the store returns ErrWebAuthnIDAlreadySet, re-read to verify.
		current, readErr := s.GetByID(ctx, user.ID)
		if readErr != nil {
			t.Fatalf("GetByID after second SetWebAuthnIDIfNull: %v", readErr)
		}
		if !bytes.Equal(current.WebAuthnID, firstID) {
			t.Errorf("WebAuthnID after second set (error path) = %x, want original %x", current.WebAuthnID, firstID)
		}
	} else {
		if !bytes.Equal(updated2.WebAuthnID, firstID) {
			t.Errorf("WebAuthnID after second set (success path) = %x, want original %x", updated2.WebAuthnID, firstID)
		}
	}
}

func TestPasskeyCascade_UserDelete(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	user := createTestUser(t, s, "cascade-delete@example.com")

	cred1 := insertCred(t, s, model.PasskeyCredential{
		UserID:          user.ID,
		CredentialID:    []byte("cascade-cred-1"),
		Name:            "Cascade Key 1",
		PublicKey:       []byte("pk1"),
		AttestationType: "none",
		AAGUID:          make([]byte, 16),
		Transports:      []string{},
	})
	cred2 := insertCred(t, s, model.PasskeyCredential{
		UserID:          user.ID,
		CredentialID:    []byte("cascade-cred-2"),
		Name:            "Cascade Key 2",
		PublicKey:       []byte("pk2"),
		AttestationType: "none",
		AAGUID:          make([]byte, 16),
		Transports:      []string{},
	})

	// Delete the user directly via SQL to trigger CASCADE.
	pool := testPool(t)
	if _, err := pool.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID); err != nil {
		t.Fatalf("DELETE user: %v", err)
	}

	// Both credentials should no longer exist.
	_, err := s.GetByIDForUser(ctx, user.ID, cred1.ID)
	if !errors.Is(err, model.ErrPasskeyNotFound) {
		t.Errorf("cred1 after user delete: err = %v, want ErrPasskeyNotFound", err)
	}
	_, err = s.GetByIDForUser(ctx, user.ID, cred2.ID)
	if !errors.Is(err, model.ErrPasskeyNotFound) {
		t.Errorf("cred2 after user delete: err = %v, want ErrPasskeyNotFound", err)
	}
}
