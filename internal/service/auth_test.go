package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"starter/internal/model"
	"starter/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var serviceTestUser = model.User{
	ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	Email:       "ada@example.com",
	DisplayName: "Ada Lovelace",
	Role:        "admin",
}

type fakeAuthUserRepo struct {
	createFn func(context.Context, string, string, string) (model.User, error)
	getByID  func(context.Context, uuid.UUID) (model.User, error)
}

func (r fakeAuthUserRepo) Create(ctx context.Context, email, displayName, role string) (model.User, error) {
	if r.createFn != nil {
		return r.createFn(ctx, email, displayName, role)
	}
	return model.User{}, errors.New("unexpected Create call")
}

func (r fakeAuthUserRepo) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	if r.getByID != nil {
		return r.getByID(ctx, id)
	}
	return model.User{}, errors.New("unexpected GetByID call")
}

func (fakeAuthUserRepo) GetByEmail(context.Context, string) (model.User, error) {
	return model.User{}, errors.New("unexpected GetByEmail call")
}

func (fakeAuthUserRepo) List(context.Context, int32, int32) ([]model.User, error) {
	return nil, errors.New("unexpected List call")
}

func (fakeAuthUserRepo) Update(context.Context, uuid.UUID, string, string, string) (model.User, error) {
	return model.User{}, errors.New("unexpected Update call")
}

func (fakeAuthUserRepo) Delete(context.Context, uuid.UUID) error {
	return errors.New("unexpected Delete call")
}

func (fakeAuthUserRepo) Count(context.Context) (int64, error) {
	return 0, errors.New("unexpected Count call")
}

type fakeAuthPasswordRepo struct {
	createFn          func(context.Context, uuid.UUID, string) (model.Password, error)
	getCurrentByEmail func(context.Context, string) (model.Password, error)
}

func (r fakeAuthPasswordRepo) Create(ctx context.Context, userID uuid.UUID, hash string) (model.Password, error) {
	if r.createFn != nil {
		return r.createFn(ctx, userID, hash)
	}
	return model.Password{}, errors.New("unexpected Create call")
}

func (fakeAuthPasswordRepo) GetCurrentByUserID(context.Context, uuid.UUID) (model.Password, error) {
	return model.Password{}, errors.New("unexpected GetCurrentByUserID call")
}

func (r fakeAuthPasswordRepo) GetCurrentByEmail(ctx context.Context, email string) (model.Password, error) {
	if r.getCurrentByEmail != nil {
		return r.getCurrentByEmail(ctx, email)
	}
	return model.Password{}, errors.New("unexpected GetCurrentByEmail call")
}

func (fakeAuthPasswordRepo) ListByUserID(context.Context, uuid.UUID) ([]model.Password, error) {
	return nil, errors.New("unexpected ListByUserID call")
}

type fakeTxFactory struct {
	repos *repository.Repositories
}

func (f fakeTxFactory) WithTx(pgx.Tx) *repository.Repositories { return f.repos }

type fakePool struct {
	tx  pgx.Tx
	err error
}

func (p fakePool) Begin(context.Context) (pgx.Tx, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.tx, nil
}

type fakeTx struct {
	commitErr      error
	rollbackErr    error
	commitCalled   bool
	rollbackCalled bool
}

func (tx *fakeTx) Begin(context.Context) (pgx.Tx, error) { return tx, nil }

func (tx *fakeTx) Commit(context.Context) error {
	tx.commitCalled = true
	return tx.commitErr
}

func (tx *fakeTx) Rollback(context.Context) error {
	tx.rollbackCalled = true
	return tx.rollbackErr
}

func (*fakeTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, errors.New("unexpected CopyFrom call")
}

func (*fakeTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }

func (*fakeTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

func (*fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, errors.New("unexpected Prepare call")
}

func (*fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec call")
}

func (*fakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query call")
}

func (*fakeTx) QueryRow(context.Context, string, ...any) pgx.Row { return nil }

func (*fakeTx) Conn() *pgx.Conn { return nil }

func TestAuthServiceRegisterCreatesUserAndPasswordInTransaction(t *testing.T) {
	tx := &fakeTx{}
	var createdHash string
	service := NewAuthService(
		nil,
		nil,
		fakeTxFactory{repos: &repository.Repositories{
			User: fakeAuthUserRepo{
				createFn: func(_ context.Context, email, displayName, role string) (model.User, error) {
					if email != "ada@example.com" || displayName != "Ada" || role != "user" {
						t.Fatalf("email=%q displayName=%q role=%q", email, displayName, role)
					}
					return serviceTestUser, nil
				},
			},
			Password: fakeAuthPasswordRepo{
				createFn: func(_ context.Context, userID uuid.UUID, hash string) (model.Password, error) {
					if userID != serviceTestUser.ID {
						t.Fatalf("userID = %s", userID)
					}
					createdHash = hash
					return model.Password{UserID: userID, Hash: hash}, nil
				},
			},
		}},
		fakePool{tx: tx},
	)

	user, err := service.Register(context.Background(), model.CreateUserInput{
		Email:       "ada@example.com",
		DisplayName: "Ada",
		Password:    "correct-password",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user != serviceTestUser {
		t.Fatalf("user = %#v", user)
	}
	if createdHash == "" || createdHash == "correct-password" {
		t.Fatalf("createdHash = %q", createdHash)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(createdHash), []byte("correct-password")); err != nil {
		t.Fatalf("stored hash did not match password: %v", err)
	}
	if !tx.commitCalled {
		t.Fatal("Commit was not called")
	}
	if !tx.rollbackCalled {
		t.Fatal("Rollback defer was not called")
	}
}

func TestAuthServiceRegisterPropagatesEmailTaken(t *testing.T) {
	tx := &fakeTx{}
	service := NewAuthService(
		nil,
		nil,
		fakeTxFactory{repos: &repository.Repositories{
			User: fakeAuthUserRepo{
				createFn: func(context.Context, string, string, string) (model.User, error) {
					return model.User{}, model.ErrEmailTaken
				},
			},
			Password: fakeAuthPasswordRepo{},
		}},
		fakePool{tx: tx},
	)

	_, err := service.Register(context.Background(), model.CreateUserInput{
		Email:       "ada@example.com",
		DisplayName: "Ada",
		Password:    "correct-password",
	})
	if !errors.Is(err, model.ErrEmailTaken) {
		t.Fatalf("err = %v", err)
	}
}

func TestAuthServiceRegisterHandlesBeginAndCommitFailures(t *testing.T) {
	beginSvc := NewAuthService(nil, nil, fakeTxFactory{}, fakePool{err: errors.New("begin failed")})
	_, err := beginSvc.Register(context.Background(), model.CreateUserInput{
		Email:       "ada@example.com",
		DisplayName: "Ada",
		Password:    "correct-password",
	})
	if err == nil || !strings.Contains(err.Error(), "begin tx") {
		t.Fatalf("err = %v", err)
	}

	tx := &fakeTx{commitErr: errors.New("commit failed")}
	commitSvc := NewAuthService(
		nil,
		nil,
		fakeTxFactory{repos: &repository.Repositories{
			User: fakeAuthUserRepo{
				createFn: func(context.Context, string, string, string) (model.User, error) { return serviceTestUser, nil },
			},
			Password: fakeAuthPasswordRepo{
				createFn: func(context.Context, uuid.UUID, string) (model.Password, error) { return model.Password{}, nil },
			},
		}},
		fakePool{tx: tx},
	)
	_, err = commitSvc.Register(context.Background(), model.CreateUserInput{
		Email:       "ada@example.com",
		DisplayName: "Ada",
		Password:    "correct-password",
	})
	if err == nil || !strings.Contains(err.Error(), "commit tx") {
		t.Fatalf("err = %v", err)
	}
}

func TestAuthServiceLoginOutcomes(t *testing.T) {
	hashed, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	tests := []struct {
		name      string
		users     fakeAuthUserRepo
		passwords fakeAuthPasswordRepo
		input     model.LoginInput
		wantUser  model.User
		wantErrIs error
		wantErr   string
	}{
		{
			name: "success",
			users: fakeAuthUserRepo{
				getByID: func(context.Context, uuid.UUID) (model.User, error) { return serviceTestUser, nil },
			},
			passwords: fakeAuthPasswordRepo{
				getCurrentByEmail: func(context.Context, string) (model.Password, error) {
					return model.Password{UserID: serviceTestUser.ID, Hash: string(hashed)}, nil
				},
			},
			input:    model.LoginInput{Email: "ada@example.com", Password: "correct-password"},
			wantUser: serviceTestUser,
		},
		{
			name: "missing credentials record",
			passwords: fakeAuthPasswordRepo{getCurrentByEmail: func(context.Context, string) (model.Password, error) {
				return model.Password{}, model.ErrInvalidCredentials
			}},
			input:     model.LoginInput{Email: "ada@example.com", Password: "correct-password"},
			wantErrIs: model.ErrInvalidCredentials,
		},
		{
			name: "wrong password",
			passwords: fakeAuthPasswordRepo{
				getCurrentByEmail: func(context.Context, string) (model.Password, error) {
					return model.Password{UserID: serviceTestUser.ID, Hash: string(hashed)}, nil
				},
			},
			input:     model.LoginInput{Email: "ada@example.com", Password: "wrong-password"},
			wantErrIs: model.ErrInvalidCredentials,
		},
		{
			name:      "password repository failure",
			passwords: fakeAuthPasswordRepo{getCurrentByEmail: func(context.Context, string) (model.Password, error) { return model.Password{}, errors.New("db down") }},
			input:     model.LoginInput{Email: "ada@example.com", Password: "correct-password"},
			wantErr:   "get password",
		},
		{
			name: "user repository failure",
			users: fakeAuthUserRepo{
				getByID: func(context.Context, uuid.UUID) (model.User, error) {
					return model.User{}, errors.New("load user failed")
				},
			},
			passwords: fakeAuthPasswordRepo{
				getCurrentByEmail: func(context.Context, string) (model.Password, error) {
					return model.Password{UserID: serviceTestUser.ID, Hash: string(hashed)}, nil
				},
			},
			input:   model.LoginInput{Email: "ada@example.com", Password: "correct-password"},
			wantErr: "get user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := NewAuthService(tc.users, tc.passwords, fakeTxFactory{}, fakePool{})
			user, err := service.Login(context.Background(), tc.input)
			if tc.wantErrIs != nil {
				if !errors.Is(err, tc.wantErrIs) {
					t.Fatalf("err = %v", err)
				}
				return
			}
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Login() error = %v", err)
			}
			if user != tc.wantUser {
				t.Fatalf("user = %#v", user)
			}
		})
	}
}
