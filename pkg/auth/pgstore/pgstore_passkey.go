package pgstore

import (
	"context"
	"errors"
	"time"

	"github.com/go-sum/auth/model"
	authdb "github.com/go-sum/auth/pgstore/db"
	authrepo "github.com/go-sum/auth/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// Compile-time interface check.
var _ authrepo.PasskeyCredentialStore = (*PgStore)(nil)

// CreateCredential inserts a new WebAuthn credential and returns the persisted record.
func (s *PgStore) CreateCredential(ctx context.Context, cred model.PasskeyCredential) (model.PasskeyCredential, error) {
	var lastUsedAt pgtype.Timestamptz
	if cred.LastUsedAt != nil {
		lastUsedAt = pgtype.Timestamptz{Time: *cred.LastUsedAt, Valid: true}
	}

	row, err := s.queries.CreatePasskeyCredential(ctx, authdb.CreatePasskeyCredentialParams{
		UserID:          cred.UserID,
		CredentialID:    cred.CredentialID,
		Name:            cred.Name,
		PublicKey:       cred.PublicKey,
		PublicKeyAlg:    cred.PublicKeyAlg,
		AttestationType: cred.AttestationType,
		Aaguid:          cred.AAGUID,
		SignCount:       cred.SignCount,
		CloneWarning:    cred.CloneWarning,
		BackupEligible:  cred.BackupEligible,
		BackupState:     cred.BackupState,
		Transports:      cred.Transports,
		Attachment:      cred.Attachment,
		LastUsedAt:      lastUsedAt,
	})
	if err != nil {
		return model.PasskeyCredential{}, mapPasskeyErr(err)
	}
	return toPasskeyModel(row), nil
}

// GetByIDForUser retrieves a credential by its UUID, scoped to the given user.
// Returns model.ErrPasskeyNotFound when no matching row exists.
func (s *PgStore) GetByIDForUser(ctx context.Context, userID, id uuid.UUID) (model.PasskeyCredential, error) {
	row, err := s.queries.GetPasskeyCredentialByIDForUser(ctx, authdb.GetPasskeyCredentialByIDForUserParams{
		ID:     id,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.PasskeyCredential{}, model.ErrPasskeyNotFound
	}
	if err != nil {
		return model.PasskeyCredential{}, err
	}
	return toPasskeyModel(row), nil
}

// GetByCredentialID retrieves a credential by its raw credential ID bytes.
// Returns model.ErrPasskeyNotFound when missing.
func (s *PgStore) GetByCredentialID(ctx context.Context, credentialID []byte) (model.PasskeyCredential, error) {
	row, err := s.queries.GetPasskeyCredentialByCredentialID(ctx, credentialID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.PasskeyCredential{}, model.ErrPasskeyNotFound
	}
	if err != nil {
		return model.PasskeyCredential{}, err
	}
	return toPasskeyModel(row), nil
}

// ListByUserID returns all credentials for a user ordered by creation date descending.
func (s *PgStore) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.PasskeyCredential, error) {
	rows, err := s.queries.ListPasskeyCredentialsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	creds := make([]model.PasskeyCredential, len(rows))
	for i, r := range rows {
		creds[i] = toPasskeyModel(r)
	}
	return creds, nil
}

// TouchPasskeyCredential atomically updates the sign count (monotonically), clone warning,
// and last_used_at timestamp for a credential in a single query.
func (s *PgStore) TouchPasskeyCredential(ctx context.Context, id uuid.UUID, signCount int64, cloneWarning bool, lastUsed time.Time) error {
	return s.queries.TouchPasskeyCredential(ctx, authdb.TouchPasskeyCredentialParams{
		ID:           id,
		SignCount:    signCount,
		CloneWarning: cloneWarning,
		LastUsedAt:   pgtype.Timestamptz{Time: lastUsed, Valid: true},
	})
}

// RenameCredential updates the display name of a credential, enforcing ownership via user_id.
// Returns model.ErrPasskeyNotFound when no matching row exists.
func (s *PgStore) RenameCredential(ctx context.Context, id, userID uuid.UUID, name string) (model.PasskeyCredential, error) {
	row, err := s.queries.RenamePasskeyCredential(ctx, authdb.RenamePasskeyCredentialParams{
		ID:     id,
		UserID: userID,
		Name:   name,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.PasskeyCredential{}, model.ErrPasskeyNotFound
	}
	if err != nil {
		return model.PasskeyCredential{}, err
	}
	return toPasskeyModel(row), nil
}

// DeleteCredential removes a credential, enforcing ownership via user_id.
// Returns model.ErrPasskeyNotFound when no matching row exists — including
// cross-user attempts, which collapse to "not found" so existence is not
// leaked across users.
func (s *PgStore) DeleteCredential(ctx context.Context, id, userID uuid.UUID) error {
	_, err := s.queries.DeletePasskeyCredential(ctx, authdb.DeletePasskeyCredentialParams{
		ID:     id,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrPasskeyNotFound
	}
	if err != nil {
		return err
	}
	return nil
}

// toPasskeyModel converts a sqlc-generated WebauthnCredential to the domain model.
func toPasskeyModel(r authdb.WebauthnCredential) model.PasskeyCredential {
	var lastUsedAt *time.Time
	if r.LastUsedAt.Valid {
		t := r.LastUsedAt.Time
		lastUsedAt = &t
	}
	return model.PasskeyCredential{
		ID:              r.ID,
		UserID:          r.UserID,
		CredentialID:    r.CredentialID,
		Name:            r.Name,
		PublicKey:       r.PublicKey,
		PublicKeyAlg:    r.PublicKeyAlg,
		AttestationType: r.AttestationType,
		AAGUID:          r.Aaguid,
		SignCount:       r.SignCount,
		CloneWarning:    r.CloneWarning,
		BackupEligible:  r.BackupEligible,
		BackupState:     r.BackupState,
		Transports:      r.Transports,
		Attachment:      r.Attachment,
		LastUsedAt:      lastUsedAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

// mapPasskeyErr translates PostgreSQL unique constraint violations for credentials to domain errors.
func mapPasskeyErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return model.ErrPasskeyAlreadyRegistered
	}
	return err
}
