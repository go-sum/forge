package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const passkeyBodyLimitBytes = 64 * 1024 // 64 KB

// passkeyService is the private interface used by PasskeyHandler.
// See also the authService convention in handler.go.
type passkeyService interface {
	BeginRegistration(ctx context.Context, userID uuid.UUID) (model.PasskeyCreationOptions, model.PasskeyCeremony, error)
	FinishRegistration(ctx context.Context, userID uuid.UUID, name string, ceremony model.PasskeyCeremony, r *http.Request) (model.PasskeyCredential, error)
	BeginAuthentication(ctx context.Context) (model.PasskeyRequestOptions, model.PasskeyCeremony, error)
	FinishAuthentication(ctx context.Context, ceremony model.PasskeyCeremony, r *http.Request) (model.VerifyResult, error)
	GetPasskey(ctx context.Context, userID, passkeyID uuid.UUID) (model.PasskeyCredential, error)
	ListPasskeys(ctx context.Context, userID uuid.UUID) ([]model.PasskeyCredential, error)
	DeletePasskey(ctx context.Context, userID, passkeyID uuid.UUID) error
	RenamePasskey(ctx context.Context, userID, passkeyID uuid.UUID, name string) (model.PasskeyCredential, error)
}

// PasskeyHandler owns passkey ceremony and management HTTP endpoints.
type PasskeyHandler struct {
	service   passkeyService
	sessions  SessionManager
	pages     PasskeyPageRenderer
	homePath  func() string
	requestFn func(c *echo.Context) Request
}

// NewPasskeyHandler constructs a PasskeyHandler.
func NewPasskeyHandler(svc PasskeyService, cfg PasskeyHandlerConfig) *PasskeyHandler {
	return &PasskeyHandler{
		service:   svc,
		sessions:  cfg.Sessions,
		pages:     cfg.Pages,
		homePath:  resolvePath(cfg.HomePath, cfg.HomePathFn),
		requestFn: cfg.RequestFn,
	}
}

func (h *PasskeyHandler) req(c *echo.Context) Request {
	if h.requestFn != nil {
		return h.requestFn(c)
	}
	return Request{}
}

// requireJSON rejects requests whose Content-Type is not application/json.
func requireJSON(c *echo.Context) error {
	ct := c.Request().Header.Get(echo.HeaderContentType)
	if !strings.HasPrefix(ct, echo.MIMEApplicationJSON) {
		return errBadRequest("Content-Type must be application/json")
	}
	return nil
}

// validatePasskeyName enforces naming constraints for passkey display names.
func validatePasskeyName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errBadRequest("Passkey name must not be empty.")
	}
	if utf8.RuneCountInString(name) > 64 {
		return errBadRequest("Passkey name must be 64 characters or fewer.")
	}
	if strings.ContainsAny(name, "\r\n\t") {
		return errBadRequest("Passkey name contains invalid characters.")
	}
	return nil
}

// RegisterBegin starts a passkey registration ceremony for the authenticated user.
func (h *PasskeyHandler) RegisterBegin(c *echo.Context) error {
	if err := requireJSON(c); err != nil {
		return err
	}

	ctx := c.Request().Context()

	state, err := h.sessions.Load(c.Request())
	if err != nil {
		return errInternal(err)
	}
	userIDRaw, ok := getUserID(state)
	if !ok {
		return errUnauthorized("Please sign in again.")
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return errUnauthorized("Please sign in again.")
	}

	creation, ceremony, err := h.service.BeginRegistration(ctx, userID)
	if err != nil {
		return errInternal(err)
	}

	clearPasskeyCeremony(state)
	if err := setPasskeyCeremony(state, passkeyCeremonyState{
		Operation: "register",
		Ceremony:  ceremony,
		UserID:    userID,
	}); err != nil {
		return errInternal(err)
	}
	if err := h.sessions.Commit(c.Response(), c.Request(), state); err != nil {
		return errInternal(err)
	}

	return c.JSON(http.StatusOK, creation)
}

// RegisterFinish completes the passkey registration ceremony.
func (h *PasskeyHandler) RegisterFinish(c *echo.Context) error {
	if err := requireJSON(c); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// 1. Read and parse body FIRST, before touching ceremony state.
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, passkeyBodyLimitBytes)
	raw, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return errBadRequest("Request body too large or unreadable.")
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return errBadRequest("Invalid request body.")
	}

	// Apply name default before validation.
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		body.Name = "Passkey " + time.Now().UTC().Format("2006-01-02")
	}
	if err := validatePasskeyName(body.Name); err != nil {
		return err
	}

	// Rewind body so the webauthn library can read credential data from it.
	c.Request().Body = io.NopCloser(bytes.NewReader(raw))

	// 2. Session + auth check.
	state, err := h.sessions.Load(c.Request())
	if err != nil {
		return errUnauthorized("Please sign in again.")
	}
	userIDRaw, ok := getUserID(state)
	if !ok {
		return errUnauthorized("Please sign in again.")
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return errUnauthorized("Please sign in again.")
	}

	// 3. Ceremony check.
	ceremonyState, ok := getPasskeyCeremony(state)
	if !ok || ceremonyState.Operation != "register" {
		return errBadRequest("No active registration ceremony. Start again.")
	}

	// 4. Cross-user check: ensure this ceremony belongs to the authenticated user.
	if ceremonyState.UserID != userID {
		clearPasskeyCeremony(state)
		if commitErr := h.sessions.Commit(c.Response(), c.Request(), state); commitErr != nil {
			slog.WarnContext(ctx, "failed to commit session after cross-user ceremony mismatch", "error", commitErr)
		}
		return errBadRequest("No active registration ceremony. Start again.")
	}

	// 5. Clear ceremony AFTER all validation.
	clearPasskeyCeremony(state)

	// 6. Service call.
	cred, err := h.service.FinishRegistration(ctx, userID, body.Name, ceremonyState.Ceremony, c.Request())
	if err != nil {
		if commitErr := h.sessions.Commit(c.Response(), c.Request(), state); commitErr != nil {
			slog.WarnContext(ctx, "failed to commit session after passkey registration error", "error", commitErr)
		}
		if errors.Is(err, model.ErrPasskeyVerificationFailed) || errors.Is(err, model.ErrPasskeyServerState) {
			return errBadRequest("Passkey verification failed. Try again.")
		}
		if errors.Is(err, model.ErrPasskeyAlreadyRegistered) {
			return errConflict("This passkey is already registered.")
		}
		return errInternal(err)
	}

	// 7. On success, rotate session ID to bind the new credential.
	if err := h.sessions.RotateID(c.Response(), c.Request(), state); err != nil {
		// Credential is persisted — log the failure but still return success.
		slog.ErrorContext(ctx, "passkey registered but session rotate failed",
			"credential_id", cred.ID,
			"user_id", userID,
			"error", err,
		)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"id":        cred.ID,
		"name":      cred.Name,
		"createdAt": cred.CreatedAt,
	})
}

// AuthenticateBegin starts a passkey authentication ceremony.
func (h *PasskeyHandler) AuthenticateBegin(c *echo.Context) error {
	if err := requireJSON(c); err != nil {
		return err
	}

	ctx := c.Request().Context()

	assertion, ceremony, err := h.service.BeginAuthentication(ctx)
	if err != nil {
		return errInternal(err)
	}

	state, err := h.sessions.Load(c.Request())
	if err != nil {
		return errInternal(err)
	}
	clearPasskeyCeremony(state)
	if err := setPasskeyCeremony(state, passkeyCeremonyState{
		Operation: "authenticate",
		Ceremony:  ceremony,
	}); err != nil {
		return errInternal(err)
	}
	if err := h.sessions.Commit(c.Response(), c.Request(), state); err != nil {
		return errInternal(err)
	}

	return c.JSON(http.StatusOK, assertion)
}

// AuthenticateFinish completes the passkey authentication ceremony.
func (h *PasskeyHandler) AuthenticateFinish(c *echo.Context) error {
	if err := requireJSON(c); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Apply body size limit before reading.
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, passkeyBodyLimitBytes)

	state, err := h.sessions.Load(c.Request())
	if err != nil {
		return errInternal(err)
	}

	ceremonyState, ok := getPasskeyCeremony(state)
	if !ok || ceremonyState.Operation != "authenticate" {
		return errBadRequest("No active authentication ceremony. Start again.")
	}

	// Always clear ceremony state, even on error.
	clearPasskeyCeremony(state)

	result, err := h.service.FinishAuthentication(ctx, ceremonyState.Ceremony, c.Request())
	if err != nil {
		_ = h.sessions.Commit(c.Response(), c.Request(), state)
		if errors.Is(err, model.ErrInvalidCredentials) {
			return errUnauthorized("Authentication failed.")
		}
		if errors.Is(err, model.ErrPasskeyCloneDetected) {
			return errUnauthorized("Authentication failed. Please try a different sign-in method.")
		}
		if errors.Is(err, model.ErrPasskeyVerificationFailed) || errors.Is(err, model.ErrPasskeyServerState) {
			return errUnauthorized("Passkey verification failed.")
		}
		return errInternal(err)
	}

	if err := setAuth(state, result.User.ID.String(), result.User.DisplayName); err != nil {
		return errInternal(err)
	}
	if err := h.sessions.RotateID(c.Response(), c.Request(), state); err != nil {
		return errInternal(err)
	}
	if bindErr := h.sessions.BindSession(ctx, state.ID(), result.User.ID.String(), SessionMeta{
		AuthMethod: result.Method,
		IPAddress:  c.RealIP(),
		UserAgent:  c.Request().UserAgent(),
	}); bindErr != nil {
		slog.ErrorContext(ctx, "failed to bind session to user", "user_id", result.User.ID, "error", bindErr)
	}

	return c.JSON(http.StatusOK, map[string]string{"redirect": h.homePath()})
}

// ListPasskeys renders the passkey management page.
func (h *PasskeyHandler) ListPasskeys(c *echo.Context) error {
	req := h.req(c)
	ctx := c.Request().Context()

	userID, err := uuid.Parse(UserID(c))
	if err != nil {
		return errUnauthorized("Your session is invalid. Please sign in again.")
	}

	creds, err := h.service.ListPasskeys(ctx, userID)
	if err != nil {
		return errInternal(err)
	}

	data := PasskeyListData{
		Passkeys:  creds,
		CSRFToken: req.CSRFToken,
	}
	if req.IsPartial() {
		return renderOK(c, h.pages.PasskeyListRegion(req, data))
	}
	return renderOK(c, h.pages.PasskeyListPage(req, data))
}

// GetRenameForm renders the inline edit form for a passkey credential.
func (h *PasskeyHandler) GetRenameForm(c *echo.Context) error {
	req := h.req(c)
	ctx := c.Request().Context()

	passkeyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errBadRequest("Invalid passkey ID.")
	}
	userID, err := uuid.Parse(UserID(c))
	if err != nil {
		return errUnauthorized("Your session is invalid. Please sign in again.")
	}

	cred, err := h.service.GetPasskey(ctx, userID, passkeyID)
	if err != nil {
		if errors.Is(err, model.ErrPasskeyNotFound) {
			return errNotFound("Passkey not found.")
		}
		return errInternal(err)
	}

	return renderOK(c, h.pages.PasskeyEditForm(req, PasskeyRowData{
		Passkey:   cred,
		CSRFToken: req.CSRFToken,
	}))
}

// GetPasskeyRow renders the read-only row fragment for a passkey credential.
func (h *PasskeyHandler) GetPasskeyRow(c *echo.Context) error {
	req := h.req(c)
	ctx := c.Request().Context()

	passkeyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errBadRequest("Invalid passkey ID.")
	}
	userID, err := uuid.Parse(UserID(c))
	if err != nil {
		return errUnauthorized("Your session is invalid. Please sign in again.")
	}

	cred, err := h.service.GetPasskey(ctx, userID, passkeyID)
	if err != nil {
		if errors.Is(err, model.ErrPasskeyNotFound) {
			return errNotFound("Passkey not found.")
		}
		return errInternal(err)
	}

	return renderOK(c, h.pages.PasskeyRow(req, PasskeyRowData{
		Passkey:   cred,
		CSRFToken: req.CSRFToken,
	}))
}

// RenamePasskey updates the display name of a passkey credential.
func (h *PasskeyHandler) RenamePasskey(c *echo.Context) error {
	req := h.req(c)
	ctx := c.Request().Context()

	passkeyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errBadRequest("Invalid passkey ID.")
	}
	userID, err := uuid.Parse(UserID(c))
	if err != nil {
		return errUnauthorized("Your session is invalid. Please sign in again.")
	}

	name := c.FormValue("name")
	if err := validatePasskeyName(name); err != nil {
		return err
	}

	cred, err := h.service.RenamePasskey(ctx, userID, passkeyID, strings.TrimSpace(name))
	if err != nil {
		if errors.Is(err, model.ErrPasskeyNotFound) {
			return errNotFound("Passkey not found.")
		}
		return errInternal(err)
	}

	return renderOK(c, h.pages.PasskeyRow(req, PasskeyRowData{
		Passkey:   cred,
		CSRFToken: req.CSRFToken,
	}))
}

// DeletePasskey removes a passkey credential.
func (h *PasskeyHandler) DeletePasskey(c *echo.Context) error {
	ctx := c.Request().Context()

	passkeyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errBadRequest("Invalid passkey ID.")
	}
	userID, err := uuid.Parse(UserID(c))
	if err != nil {
		return errUnauthorized("Your session is invalid. Please sign in again.")
	}

	if err := h.service.DeletePasskey(ctx, userID, passkeyID); err != nil {
		if errors.Is(err, model.ErrPasskeyNotFound) {
			return errNotFound("Passkey not found.")
		}
		return errInternal(err)
	}

	return c.String(http.StatusOK, "")
}
