package pgstore_test

import (
	"context"
	"errors"
	"os"
	"testing"

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
		t.Fatalf("users table is missing; run make db-migrate before these tests")
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
	if got != created {
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
	if got != created {
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
