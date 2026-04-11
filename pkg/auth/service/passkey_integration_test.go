package service

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-sum/auth"
	"github.com/go-sum/auth/model"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// ── Software authenticator helpers ───────────────────────────────────────────

// softwareCredential holds the state produced during a simulated registration.
type softwareCredential struct {
	credentialID []byte
	privateKey   *ecdsa.PrivateKey
	signCount    uint32
}

// buildCOSEPublicKey encodes an ECDSA public key in COSE_Key / CBOR format.
func buildCOSEPublicKey(pub *ecdsa.PublicKey) ([]byte, error) {
	xPad := padCoord(pub.X.Bytes())
	yPad := padCoord(pub.Y.Bytes())

	coseKey := webauthncose.EC2PublicKeyData{
		PublicKeyData: webauthncose.PublicKeyData{
			KeyType:   int64(webauthncose.EllipticKey),
			Algorithm: int64(webauthncose.AlgES256),
		},
		Curve:  int64(webauthncose.P256),
		XCoord: xPad,
		YCoord: yPad,
	}
	return webauthncbor.Marshal(coseKey)
}

// padCoord pads a big-endian coordinate to exactly 32 bytes for P-256.
func padCoord(b []byte) []byte {
	if len(b) == 32 {
		return b
	}
	padded := make([]byte, 32)
	copy(padded[32-len(b):], b)
	return padded
}

// buildAuthData builds the authenticator data byte string.
// flags: 0x41 = UP | AT (with attested credential data), 0x01 = UP only.
func buildAuthData(rpID string, flags byte, signCount uint32, credentialID []byte, cosePublicKey []byte) []byte {
	rpIDHash := sha256.Sum256([]byte(rpID))

	var buf bytes.Buffer

	// RP ID hash (32 bytes)
	buf.Write(rpIDHash[:])

	// Flags (1 byte)
	buf.WriteByte(flags)

	// Sign count (4 bytes, big-endian)
	countBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(countBuf, signCount)
	buf.Write(countBuf)

	// Attested credential data only when AT flag is set
	if flags&0x40 != 0 {
		// AAGUID (16 zero bytes)
		buf.Write(make([]byte, 16))

		// Credential ID length (2 bytes, big-endian)
		credIDLen := make([]byte, 2)
		binary.BigEndian.PutUint16(credIDLen, uint16(len(credentialID)))
		buf.Write(credIDLen)

		// Credential ID
		buf.Write(credentialID)

		// COSE public key
		buf.Write(cosePublicKey)
	}

	return buf.Bytes()
}

// buildAttestationObject wraps authData in a CBOR attestation object with fmt=none.
func buildAttestationObject(authData []byte) ([]byte, error) {
	attObj := map[string]any{
		"fmt":      "none",
		"attStmt":  map[string]any{},
		"authData": authData,
	}
	return webauthncbor.Marshal(attObj)
}

// simulateRegistration performs a fake registration and returns the credential + POST request body.
func simulateRegistration(t *testing.T, rpID string, creationOptionsJSON []byte) (*softwareCredential, []byte) {
	t.Helper()

	// Parse the creation options to extract the challenge.
	var pubKey struct {
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(creationOptionsJSON, &pubKey); err != nil {
		t.Fatalf("simulateRegistration: parse creation options: %v", err)
	}

	// Generate ES256 key pair.
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("simulateRegistration: generate key: %v", err)
	}

	// Build clientDataJSON.
	clientData := fmt.Sprintf(`{"type":"webauthn.create","challenge":%q,"origin":"https://%s","crossOrigin":false}`,
		pubKey.Challenge, rpID)

	// Encode public key as COSE.
	coseKey, err := buildCOSEPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatalf("simulateRegistration: build COSE key: %v", err)
	}

	// Random 16-byte credential ID.
	credID := make([]byte, 16)
	if _, err := rand.Read(credID); err != nil {
		t.Fatalf("simulateRegistration: generate credID: %v", err)
	}

	// Build authenticator data: UP | AT flags = 0x41.
	authData := buildAuthData(rpID, 0x41, 0, credID, coseKey)

	// Build attestation object.
	attObj, err := buildAttestationObject(authData)
	if err != nil {
		t.Fatalf("simulateRegistration: build attestation object: %v", err)
	}

	// Assemble the JSON body (base64url-encoded per protocol.URLEncodedBase64).
	body := buildRegistrationBody(credID, attObj, []byte(clientData))

	sc := &softwareCredential{
		credentialID: credID,
		privateKey:   privKey,
		signCount:    0,
	}
	return sc, body
}

// buildRegistrationBody assembles the JSON body for FinishRegistration.
// Values are raw bytes — the go-webauthn library expects URLEncodedBase64.
func buildRegistrationBody(credID, attObj, clientDataJSON []byte) []byte {
	body := map[string]any{
		"id":    base64URLEncode(credID),
		"rawId": base64URLEncode(credID),
		"type":  "public-key",
		"response": map[string]any{
			"attestationObject": base64URLEncode(attObj),
			"clientDataJSON":    base64URLEncode(clientDataJSON),
			"transports":        []string{"usb"},
		},
	}
	b, _ := json.Marshal(body)
	return b
}

// simulateAuthentication produces a fake assertion POST body using the stored credential.
func simulateAuthentication(t *testing.T, rpID string, requestOptionsJSON []byte, sc *softwareCredential) []byte {
	t.Helper()

	var pubKey struct {
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(requestOptionsJSON, &pubKey); err != nil {
		t.Fatalf("simulateAuthentication: parse request options: %v", err)
	}

	sc.signCount++

	// Build clientDataJSON.
	clientData := fmt.Sprintf(`{"type":"webauthn.get","challenge":%q,"origin":"https://%s","crossOrigin":false}`,
		pubKey.Challenge, rpID)

	// Build authenticator data: UP flag only = 0x01.
	authData := buildAuthData(rpID, 0x01, sc.signCount, nil, nil)

	// Compute signature: sign( sha256(authData || sha256(clientDataJSON)) ).
	clientDataHash := sha256.Sum256([]byte(clientData))
	sigInput := append(authData, clientDataHash[:]...)
	sigHash := sha256.Sum256(sigInput)

	sig, err := ecdsa.SignASN1(rand.Reader, sc.privateKey, sigHash[:])
	if err != nil {
		t.Fatalf("simulateAuthentication: sign: %v", err)
	}

	body := map[string]any{
		"id":    base64URLEncode(sc.credentialID),
		"rawId": base64URLEncode(sc.credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"authenticatorData": base64URLEncode(authData),
			"clientDataJSON":    base64URLEncode([]byte(clientData)),
			"signature":         base64URLEncode(sig),
			"userHandle":        "", // discoverable login — server looks up by handle
		},
	}
	b, _ := json.Marshal(body)
	return b
}

// base64URLEncode returns the base64url (no-padding) representation.
func base64URLEncode(b []byte) string {
	const table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	return encodeBase64URL(b, table)
}

func encodeBase64URL(src []byte, table string) string {
	n := len(src)
	buf := make([]byte, (n*8+5)/6) // ceiling of n*8/6
	pos := 0
	for i := 0; i < n; i += 3 {
		var b [3]byte
		b[0] = src[i]
		end := 1
		if i+1 < n {
			b[1] = src[i+1]
			end = 2
		}
		if i+2 < n {
			b[2] = src[i+2]
			end = 3
		}
		buf[pos] = table[b[0]>>2]
		buf[pos+1] = table[(b[0]&0x03)<<4|b[1]>>4]
		pos += 2
		if end >= 2 {
			buf[pos] = table[(b[1]&0x0F)<<2|b[2]>>6]
			pos++
		}
		if end >= 3 {
			buf[pos] = table[b[2]&0x3F]
			pos++
		}
	}
	return string(buf[:pos])
}

// ── In-memory stores ───────────────────────────────────────────────────────

// integrationUserStore is a full in-memory UserStore for integration tests.
type integrationUserStore struct {
	users map[uuid.UUID]model.User
}

func newIntegrationUserStore() *integrationUserStore {
	return &integrationUserStore{users: make(map[uuid.UUID]model.User)}
}

func (s *integrationUserStore) addUser(u model.User) {
	s.users[u.ID] = u
}

func (s *integrationUserStore) GetByID(_ context.Context, id uuid.UUID) (model.User, error) {
	u, ok := s.users[id]
	if !ok {
		return model.User{}, model.ErrUserNotFound
	}
	return u, nil
}

func (s *integrationUserStore) GetByEmail(_ context.Context, email string) (model.User, error) {
	for _, u := range s.users {
		if u.Email == email {
			return u, nil
		}
	}
	return model.User{}, model.ErrUserNotFound
}

func (s *integrationUserStore) Create(_ context.Context, email, displayName, role string, verified bool) (model.User, error) {
	u := model.User{
		ID:          uuid.New(),
		Email:       email,
		DisplayName: displayName,
		Role:        role,
		Verified:    verified,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.users[u.ID] = u
	return u, nil
}

func (s *integrationUserStore) UpdateEmail(_ context.Context, id uuid.UUID, email string) (model.User, error) {
	u, ok := s.users[id]
	if !ok {
		return model.User{}, model.ErrUserNotFound
	}
	u.Email = email
	s.users[id] = u
	return u, nil
}

func (s *integrationUserStore) SetWebAuthnID(_ context.Context, id uuid.UUID, webauthnID []byte) (model.User, error) {
	u, ok := s.users[id]
	if !ok {
		return model.User{}, model.ErrUserNotFound
	}
	u.WebAuthnID = webauthnID
	s.users[id] = u
	return u, nil
}

func (s *integrationUserStore) SetWebAuthnIDIfNull(_ context.Context, id uuid.UUID, webauthnID []byte) (model.User, error) {
	u, ok := s.users[id]
	if !ok {
		return model.User{}, model.ErrUserNotFound
	}
	if len(u.WebAuthnID) > 0 {
		return u, model.ErrWebAuthnIDAlreadySet
	}
	u.WebAuthnID = webauthnID
	s.users[id] = u
	return u, nil
}

func (s *integrationUserStore) GetByWebAuthnID(_ context.Context, webauthnID []byte) (model.User, error) {
	for _, u := range s.users {
		if bytes.Equal(u.WebAuthnID, webauthnID) {
			return u, nil
		}
	}
	return model.User{}, model.ErrUserNotFound
}

// integrationCredentialStore is a full in-memory PasskeyCredentialStore.
type integrationCredentialStore struct {
	creds []model.PasskeyCredential
}

func newIntegrationCredentialStore() *integrationCredentialStore {
	return &integrationCredentialStore{}
}

func (s *integrationCredentialStore) CreateCredential(_ context.Context, cred model.PasskeyCredential) (model.PasskeyCredential, error) {
	// Check for duplicate credential ID.
	for _, c := range s.creds {
		if bytes.Equal(c.CredentialID, cred.CredentialID) {
			return model.PasskeyCredential{}, model.ErrPasskeyAlreadyRegistered
		}
	}
	cred.ID = uuid.New()
	cred.CreatedAt = time.Now()
	cred.UpdatedAt = time.Now()
	s.creds = append(s.creds, cred)
	return cred, nil
}

func (s *integrationCredentialStore) GetByCredentialID(_ context.Context, credentialID []byte) (model.PasskeyCredential, error) {
	for _, c := range s.creds {
		if bytes.Equal(c.CredentialID, credentialID) {
			return c, nil
		}
	}
	return model.PasskeyCredential{}, model.ErrPasskeyNotFound
}

func (s *integrationCredentialStore) GetByIDForUser(_ context.Context, userID, id uuid.UUID) (model.PasskeyCredential, error) {
	for _, c := range s.creds {
		if c.UserID == userID && c.ID == id {
			return c, nil
		}
	}
	return model.PasskeyCredential{}, model.ErrPasskeyNotFound
}

func (s *integrationCredentialStore) ListByUserID(_ context.Context, userID uuid.UUID) ([]model.PasskeyCredential, error) {
	var result []model.PasskeyCredential
	for _, c := range s.creds {
		if c.UserID == userID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (s *integrationCredentialStore) TouchPasskeyCredential(_ context.Context, id uuid.UUID, signCount int64, cloneWarning bool, lastUsed time.Time) error {
	for i, c := range s.creds {
		if c.ID == id {
			// Monotonic counter: only advance if new count is higher.
			if signCount > s.creds[i].SignCount {
				s.creds[i].SignCount = signCount
			}
			s.creds[i].CloneWarning = cloneWarning
			now := lastUsed
			s.creds[i].LastUsedAt = &now
			return nil
		}
	}
	return model.ErrPasskeyNotFound
}

func (s *integrationCredentialStore) RenameCredential(_ context.Context, id, userID uuid.UUID, name string) (model.PasskeyCredential, error) {
	for i, c := range s.creds {
		if c.ID == id && c.UserID == userID {
			s.creds[i].Name = name
			return s.creds[i], nil
		}
	}
	return model.PasskeyCredential{}, model.ErrPasskeyNotFound
}

func (s *integrationCredentialStore) DeleteCredential(_ context.Context, id, userID uuid.UUID) error {
	for i, c := range s.creds {
		if c.ID == id && c.UserID == userID {
			s.creds = append(s.creds[:i], s.creds[i+1:]...)
			return nil
		}
	}
	return model.ErrPasskeyNotFound
}

// ── Service factory ───────────────────────────────────────────────────────────

const integrationRPID = "localhost"

func testIntegrationPasskeyService(t *testing.T, users *integrationUserStore, creds *integrationCredentialStore) *PasskeyServiceImpl {
	t.Helper()
	cfg := auth.PasskeyMethodConfig{
		Enabled:               true,
		RPDisplayName:         "Test",
		RPID:                  integrationRPID,
		RPOrigins:             []string{"https://localhost"},
		ResidentKey:           "discouraged",
		UserVerification:      "preferred",
		RegistrationTimeout:   5 * time.Minute,
		AuthenticationTimeout: 2 * time.Minute,
	}
	svc, err := NewPasskeyService(users, creds, cfg)
	if err != nil {
		t.Fatalf("NewPasskeyService: %v", err)
	}
	return svc
}

// newRegistrationRequest wraps a body in an http.Request with the expected Content-Type.
func newRegistrationRequest(body []byte) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/passkeys/register/finish", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// newAuthenticationRequest wraps a body in an http.Request.
func newAuthenticationRequest(body []byte) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/passkeys/authenticate/finish", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// extractPublicKey parses model.PasskeyCreationOptions to get the inner publicKey JSON.
func extractPublicKey(opts model.PasskeyCreationOptions) []byte {
	return opts.PublicKey
}

// extractRequestOptions parses model.PasskeyRequestOptions.
func extractRequestOptions(opts model.PasskeyRequestOptions) []byte {
	return opts.PublicKey
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestPasskeyService_RegistrationRoundTrip(t *testing.T) {
	users := newIntegrationUserStore()
	creds := newIntegrationCredentialStore()

	webAuthnID := make([]byte, 64)
	if _, err := rand.Read(webAuthnID); err != nil {
		t.Fatalf("generate webauthn ID: %v", err)
	}
	user := model.User{
		ID:          uuid.New(),
		Email:       "alice@example.com",
		DisplayName: "Alice",
		Role:        model.RoleUser,
		Verified:    true,
		WebAuthnID:  webAuthnID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	users.addUser(user)

	svc := testIntegrationPasskeyService(t, users, creds)
	ctx := context.Background()

	// Begin registration.
	creationOpts, ceremony, err := svc.BeginRegistration(ctx, user.ID)
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}

	// Simulate authenticator: produce attestation response.
	sc, bodyBytes := simulateRegistration(t, integrationRPID, extractPublicKey(creationOpts))

	// Wrap body bytes back into a request (handler normally does this).
	req := newRegistrationRequest(bodyBytes)

	// Finish registration.
	cred, err := svc.FinishRegistration(ctx, user.ID, "My Key", ceremony, req)
	if err != nil {
		t.Fatalf("FinishRegistration: %v", err)
	}

	// Assert credential has correct fields.
	if cred.ID == (uuid.UUID{}) {
		t.Error("credential ID is zero UUID")
	}
	if len(cred.PublicKey) == 0 {
		t.Error("credential PublicKey is empty")
	}
	if len(cred.CredentialID) == 0 {
		t.Error("credential CredentialID is empty")
	}
	if !bytes.Equal(cred.CredentialID, sc.credentialID) {
		t.Errorf("CredentialID = %x, want %x", cred.CredentialID, sc.credentialID)
	}
	// PublicKeyAlg is populated from the client's JSON response field publicKeyAlgorithm.
	// Our software authenticator does not set this optional field; the key algorithm
	// is still verified by the webauthn library against the COSE key in authData.
	// Assert that the stored public key bytes are non-empty (key was saved).
	if len(cred.PublicKey) == 0 {
		t.Error("PublicKey bytes are empty, expected COSE-encoded key")
	}
	if cred.Name != "My Key" {
		t.Errorf("Name = %q, want %q", cred.Name, "My Key")
	}
}

func TestPasskeyService_CredParamsRoundTrip(t *testing.T) {
	// Regression test for T0-2: empty CredParams causes FinishRegistration to always fail
	// with ErrAttestationFormat. Verify that toCeremony(BeginRegistration) preserves CredParams.
	users := newIntegrationUserStore()
	creds := newIntegrationCredentialStore()

	webAuthnID := make([]byte, 64)
	_, _ = rand.Read(webAuthnID)
	user := model.User{
		ID:          uuid.New(),
		Email:       "bob@example.com",
		DisplayName: "Bob",
		Role:        model.RoleUser,
		Verified:    true,
		WebAuthnID:  webAuthnID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	users.addUser(user)

	svc := testIntegrationPasskeyService(t, users, creds)

	_, ceremony, err := svc.BeginRegistration(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}

	if len(ceremony.CredentialParameters) == 0 {
		t.Fatal("CredentialParameters is empty after BeginRegistration; fromCeremony will produce empty CredParams causing FinishRegistration to always fail")
	}

	// Assert the round-trip via fromCeremony preserves all parameters.
	sd := fromCeremony(ceremony)
	if len(sd.CredParams) != len(ceremony.CredentialParameters) {
		t.Fatalf("CredParams len = %d, want %d", len(sd.CredParams), len(ceremony.CredentialParameters))
	}
	for i, p := range sd.CredParams {
		want := ceremony.CredentialParameters[i]
		if int64(p.Algorithm) != want.Algorithm {
			t.Errorf("CredParams[%d].Algorithm = %d, want %d", i, p.Algorithm, want.Algorithm)
		}
		if string(p.Type) != want.Type {
			t.Errorf("CredParams[%d].Type = %q, want %q", i, p.Type, want.Type)
		}
	}
}

func TestPasskeyService_SessionDataRoundTrip(t *testing.T) {
	// Intent: toCeremony/fromCeremony must losslessly carry every field
	// webauthn.SessionData exposes. This catches future drift when go-webauthn
	// adds new SessionData fields.
	orig := webauthn.SessionData{
		Challenge:            "challenge-xyz",
		RelyingPartyID:       "rp.example.com",
		UserID:               []byte("user-handle"),
		AllowedCredentialIDs: [][]byte{[]byte("cred-1"), []byte("cred-2")},
		UserVerification:     protocol.VerificationPreferred,
		Extensions:           protocol.AuthenticationExtensions{"appid": "legacy"},
		CredParams: []protocol.CredentialParameter{
			{Type: protocol.PublicKeyCredentialType, Algorithm: webauthncose.AlgES256},
			{Type: protocol.PublicKeyCredentialType, Algorithm: webauthncose.AlgRS256},
		},
		Mediation: protocol.MediationConditional,
		Expires:   time.Now().Add(5 * time.Minute).UTC(),
	}

	got := fromCeremony(toCeremony(&orig))

	if got.Challenge != orig.Challenge {
		t.Errorf("Challenge = %q, want %q", got.Challenge, orig.Challenge)
	}
	if got.RelyingPartyID != orig.RelyingPartyID {
		t.Errorf("RelyingPartyID = %q, want %q", got.RelyingPartyID, orig.RelyingPartyID)
	}
	if !bytes.Equal(got.UserID, orig.UserID) {
		t.Errorf("UserID = %v, want %v", got.UserID, orig.UserID)
	}
	if len(got.AllowedCredentialIDs) != len(orig.AllowedCredentialIDs) {
		t.Fatalf("AllowedCredentialIDs len = %d, want %d", len(got.AllowedCredentialIDs), len(orig.AllowedCredentialIDs))
	}
	for i := range orig.AllowedCredentialIDs {
		if !bytes.Equal(got.AllowedCredentialIDs[i], orig.AllowedCredentialIDs[i]) {
			t.Errorf("AllowedCredentialIDs[%d] = %v, want %v", i, got.AllowedCredentialIDs[i], orig.AllowedCredentialIDs[i])
		}
	}
	if got.UserVerification != orig.UserVerification {
		t.Errorf("UserVerification = %q, want %q", got.UserVerification, orig.UserVerification)
	}
	if got.Mediation != orig.Mediation {
		t.Errorf("Mediation = %q, want %q", got.Mediation, orig.Mediation)
	}
	if len(got.Extensions) != len(orig.Extensions) {
		t.Fatalf("Extensions len = %d, want %d", len(got.Extensions), len(orig.Extensions))
	}
	for k, wantV := range orig.Extensions {
		if got.Extensions[k] != wantV {
			t.Errorf("Extensions[%q] = %v, want %v", k, got.Extensions[k], wantV)
		}
	}
	if len(got.CredParams) != len(orig.CredParams) {
		t.Fatalf("CredParams len = %d, want %d", len(got.CredParams), len(orig.CredParams))
	}
	for i, p := range orig.CredParams {
		if got.CredParams[i].Algorithm != p.Algorithm {
			t.Errorf("CredParams[%d].Algorithm = %d, want %d", i, got.CredParams[i].Algorithm, p.Algorithm)
		}
		if got.CredParams[i].Type != p.Type {
			t.Errorf("CredParams[%d].Type = %q, want %q", i, got.CredParams[i].Type, p.Type)
		}
	}
	if !got.Expires.Equal(orig.Expires) {
		t.Errorf("Expires = %v, want %v", got.Expires, orig.Expires)
	}
}

func TestPasskeyService_AuthenticationRoundTrip(t *testing.T) {
	users := newIntegrationUserStore()
	creds := newIntegrationCredentialStore()

	webAuthnID := make([]byte, 64)
	_, _ = rand.Read(webAuthnID)
	user := model.User{
		ID:          uuid.New(),
		Email:       "carol@example.com",
		DisplayName: "Carol",
		Role:        model.RoleUser,
		Verified:    true,
		WebAuthnID:  webAuthnID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	users.addUser(user)

	svc := testIntegrationPasskeyService(t, users, creds)
	ctx := context.Background()

	// Register a credential first.
	creationOpts, regCeremony, err := svc.BeginRegistration(ctx, user.ID)
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	sc, regBody := simulateRegistration(t, integrationRPID, extractPublicKey(creationOpts))
	regReq := newRegistrationRequest(regBody)
	_, err = svc.FinishRegistration(ctx, user.ID, "Test Key", regCeremony, regReq)
	if err != nil {
		t.Fatalf("FinishRegistration: %v", err)
	}

	// Now authenticate.
	requestOpts, authCeremony, err := svc.BeginAuthentication(ctx)
	if err != nil {
		t.Fatalf("BeginAuthentication: %v", err)
	}

	// The discoverable login handler looks up by userHandle (WebAuthnID).
	// We need to provide a valid userHandle in the assertion response.
	authBody := simulateAuthenticationWithHandle(t, integrationRPID, extractRequestOptions(requestOpts), sc, webAuthnID)
	authReq := newAuthenticationRequest(authBody)

	result, err := svc.FinishAuthentication(ctx, authCeremony, authReq)
	if err != nil {
		t.Fatalf("FinishAuthentication: %v", err)
	}
	if result.User.ID != user.ID {
		t.Errorf("result.User.ID = %s, want %s", result.User.ID, user.ID)
	}
	if result.Method != string(auth.MethodPasskey) {
		t.Errorf("result.Method = %q, want %q", result.Method, auth.MethodPasskey)
	}

	// Assert sign count advanced (was 0 after registration; now >= 1).
	stored, err := creds.GetByCredentialID(ctx, sc.credentialID)
	if err != nil {
		t.Fatalf("GetByCredentialID: %v", err)
	}
	if stored.SignCount < 1 {
		t.Errorf("SignCount = %d, want >= 1", stored.SignCount)
	}
}

func TestPasskeyService_FinishAuthentication_RejectsUnverifiedUser(t *testing.T) {
	users := newIntegrationUserStore()
	creds := newIntegrationCredentialStore()

	webAuthnID := make([]byte, 64)
	_, _ = rand.Read(webAuthnID)
	user := model.User{
		ID:          uuid.New(),
		Email:       "unverified@example.com",
		DisplayName: "Unverified",
		Role:        model.RoleUser,
		Verified:    false, // unverified — the case under test
		WebAuthnID:  webAuthnID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	users.addUser(user)

	svc := testIntegrationPasskeyService(t, users, creds)
	ctx := context.Background()

	// Register a passkey for the unverified user (registration does not gate on Verified).
	creationOpts, regCeremony, err := svc.BeginRegistration(ctx, user.ID)
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	sc, regBody := simulateRegistration(t, integrationRPID, extractPublicKey(creationOpts))
	regReq := newRegistrationRequest(regBody)
	if _, err = svc.FinishRegistration(ctx, user.ID, "Test Key", regCeremony, regReq); err != nil {
		t.Fatalf("FinishRegistration: %v", err)
	}

	// Attempt authentication — must be rejected because the user is unverified.
	requestOpts, authCeremony, err := svc.BeginAuthentication(ctx)
	if err != nil {
		t.Fatalf("BeginAuthentication: %v", err)
	}
	authBody := simulateAuthenticationWithHandle(t, integrationRPID, extractRequestOptions(requestOpts), sc, webAuthnID)
	authReq := newAuthenticationRequest(authBody)

	_, err = svc.FinishAuthentication(ctx, authCeremony, authReq)
	if !errors.Is(err, model.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want errors.Is(err, model.ErrInvalidCredentials)", err)
	}
}

// simulateAuthenticationWithHandle produces an assertion with explicit userHandle.
func simulateAuthenticationWithHandle(t *testing.T, rpID string, requestOptionsJSON []byte, sc *softwareCredential, userHandle []byte) []byte {
	t.Helper()

	var pubKey struct {
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(requestOptionsJSON, &pubKey); err != nil {
		t.Fatalf("simulateAuthenticationWithHandle: parse request options: %v", err)
	}

	sc.signCount++

	clientData := fmt.Sprintf(`{"type":"webauthn.get","challenge":%q,"origin":"https://%s","crossOrigin":false}`,
		pubKey.Challenge, rpID)

	authData := buildAuthData(rpID, 0x01, sc.signCount, nil, nil)

	clientDataHash := sha256.Sum256([]byte(clientData))
	sigInput := append(authData, clientDataHash[:]...)
	sigHash := sha256.Sum256(sigInput)

	sig, err := ecdsa.SignASN1(rand.Reader, sc.privateKey, sigHash[:])
	if err != nil {
		t.Fatalf("simulateAuthenticationWithHandle: sign: %v", err)
	}

	body := map[string]any{
		"id":    base64URLEncode(sc.credentialID),
		"rawId": base64URLEncode(sc.credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"authenticatorData": base64URLEncode(authData),
			"clientDataJSON":    base64URLEncode([]byte(clientData)),
			"signature":         base64URLEncode(sig),
			"userHandle":        base64URLEncode(userHandle),
		},
	}
	b, _ := json.Marshal(body)
	return b
}

func TestPasskeyService_ExpiredCeremony_Rejected(t *testing.T) {
	users := newIntegrationUserStore()
	creds := newIntegrationCredentialStore()

	webAuthnID := make([]byte, 64)
	_, _ = rand.Read(webAuthnID)
	user := model.User{
		ID:         uuid.New(),
		Email:      "dave@example.com",
		DisplayName: "Dave",
		Role:       model.RoleUser,
		Verified:   true,
		WebAuthnID: webAuthnID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	users.addUser(user)

	svc := testIntegrationPasskeyService(t, users, creds)
	ctx := context.Background()

	creationOpts, ceremony, err := svc.BeginRegistration(ctx, user.ID)
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	_, regBody := simulateRegistration(t, integrationRPID, extractPublicKey(creationOpts))

	// Expire the ceremony by setting Expires to the past.
	ceremony.Expires = time.Now().Add(-1 * time.Minute)

	req := newRegistrationRequest(regBody)
	_, err = svc.FinishRegistration(ctx, user.ID, "Expired Key", ceremony, req)
	if err == nil {
		t.Fatal("expected error for expired ceremony, got nil")
	}
	if !isPasskeyVerificationFailed(err) {
		t.Errorf("err = %v, want to contain ErrPasskeyVerificationFailed", err)
	}
}

// isPasskeyVerificationFailed checks whether err wraps ErrPasskeyVerificationFailed.
func isPasskeyVerificationFailed(err error) bool {
	type unwrapper interface{ Unwrap() []error }
	for err != nil {
		if err == model.ErrPasskeyVerificationFailed {
			return true
		}
		if u, ok := err.(unwrapper); ok {
			for _, e := range u.Unwrap() {
				if isPasskeyVerificationFailed(e) {
					return true
				}
			}
			return false
		}
		type singleUnwrap interface{ Unwrap() error }
		if u, ok := err.(singleUnwrap); ok {
			err = u.Unwrap()
		} else {
			break
		}
	}
	return false
}

func TestPasskeyService_DuplicateCredentialID_ReturnsAlreadyRegistered(t *testing.T) {
	users := newIntegrationUserStore()
	creds := newIntegrationCredentialStore()

	webAuthnID := make([]byte, 64)
	_, _ = rand.Read(webAuthnID)
	user := model.User{
		ID:          uuid.New(),
		Email:       "eve@example.com",
		DisplayName: "Eve",
		Role:        model.RoleUser,
		Verified:    true,
		WebAuthnID:  webAuthnID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	users.addUser(user)

	svc := testIntegrationPasskeyService(t, users, creds)
	ctx := context.Background()

	// First registration — must succeed.
	creationOpts1, ceremony1, err := svc.BeginRegistration(ctx, user.ID)
	if err != nil {
		t.Fatalf("first BeginRegistration: %v", err)
	}
	sc, bodyBytes1 := simulateRegistration(t, integrationRPID, extractPublicKey(creationOpts1))
	_, err = svc.FinishRegistration(ctx, user.ID, "Key 1", ceremony1, newRegistrationRequest(bodyBytes1))
	if err != nil {
		t.Fatalf("first FinishRegistration: %v", err)
	}

	// Second registration using the same credential ID — the credential store
	// detects the duplicate and returns ErrPasskeyAlreadyRegistered.
	// We inject it by directly creating a credential with the same CredentialID.
	dupCred := model.PasskeyCredential{
		UserID:       user.ID,
		CredentialID: sc.credentialID,
		Name:         "Dup",
		PublicKey:    []byte("fake"),
		AAGUID:       make([]byte, 16),
	}
	_, err = creds.CreateCredential(ctx, dupCred)
	if err == nil {
		t.Fatal("expected ErrPasskeyAlreadyRegistered for duplicate CredentialID, got nil")
	}
	if err != model.ErrPasskeyAlreadyRegistered {
		t.Errorf("err = %v, want ErrPasskeyAlreadyRegistered", err)
	}
}

func TestPasskeyService_CloneWarning_RejectsAuthentication(t *testing.T) {
	users := newIntegrationUserStore()
	creds := newIntegrationCredentialStore()

	webAuthnID := make([]byte, 64)
	_, _ = rand.Read(webAuthnID)
	user := model.User{
		ID:          uuid.New(),
		Email:       "frank@example.com",
		DisplayName: "Frank",
		Role:        model.RoleUser,
		Verified:    true,
		WebAuthnID:  webAuthnID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	users.addUser(user)

	svc := testIntegrationPasskeyService(t, users, creds)
	ctx := context.Background()

	// Register credential (stored with sign_count = 0).
	creationOpts, regCeremony, err := svc.BeginRegistration(ctx, user.ID)
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	sc, regBody := simulateRegistration(t, integrationRPID, extractPublicKey(creationOpts))
	_, err = svc.FinishRegistration(ctx, user.ID, "Clone Test Key", regCeremony, newRegistrationRequest(regBody))
	if err != nil {
		t.Fatalf("FinishRegistration: %v", err)
	}

	// Manually advance the stored sign count to 5 to simulate a previous use.
	stored, err := creds.GetByCredentialID(ctx, sc.credentialID)
	if err != nil {
		t.Fatalf("GetByCredentialID: %v", err)
	}
	_ = creds.TouchPasskeyCredential(ctx, stored.ID, 5, false, time.Now())

	// Now attempt authentication where the authenticator presents sign_count=3 (regression).
	// We directly set sc.signCount to produce a lower value in the assertion.
	sc.signCount = 3

	requestOpts, authCeremony, err := svc.BeginAuthentication(ctx)
	if err != nil {
		t.Fatalf("BeginAuthentication: %v", err)
	}

	// Build auth body using the regressed sign count.
	// Note: sc.signCount is 3, but the stored count is 5, so the library
	// will set CloneWarning=true.
	authBody := buildCloneWarningAuthBody(t, integrationRPID, extractRequestOptions(requestOpts), sc, webAuthnID)
	authReq := newAuthenticationRequest(authBody)

	_, err = svc.FinishAuthentication(ctx, authCeremony, authReq)
	if err == nil {
		t.Fatal("expected error for clone warning, got nil")
	}
	if !isCloneDetected(err) {
		t.Errorf("err = %v, want to contain ErrPasskeyCloneDetected", err)
	}
}

// buildCloneWarningAuthBody produces an assertion with the sc.signCount unchanged
// (does not increment), triggering a clone warning when stored sign count is higher.
func buildCloneWarningAuthBody(t *testing.T, rpID string, requestOptionsJSON []byte, sc *softwareCredential, userHandle []byte) []byte {
	t.Helper()

	var pubKey struct {
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(requestOptionsJSON, &pubKey); err != nil {
		t.Fatalf("buildCloneWarningAuthBody: parse request options: %v", err)
	}

	// Do NOT increment sign count — present the stale value.
	clientData := fmt.Sprintf(`{"type":"webauthn.get","challenge":%q,"origin":"https://%s","crossOrigin":false}`,
		pubKey.Challenge, rpID)

	authData := buildAuthData(rpID, 0x01, sc.signCount, nil, nil)

	clientDataHash := sha256.Sum256([]byte(clientData))
	sigInput := append(authData, clientDataHash[:]...)
	sigHash := sha256.Sum256(sigInput)

	sig, err := ecdsa.SignASN1(rand.Reader, sc.privateKey, sigHash[:])
	if err != nil {
		t.Fatalf("buildCloneWarningAuthBody: sign: %v", err)
	}

	body := map[string]any{
		"id":    base64URLEncode(sc.credentialID),
		"rawId": base64URLEncode(sc.credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"authenticatorData": base64URLEncode(authData),
			"clientDataJSON":    base64URLEncode([]byte(clientData)),
			"signature":         base64URLEncode(sig),
			"userHandle":        base64URLEncode(userHandle),
		},
	}
	b, _ := json.Marshal(body)
	return b
}

// isCloneDetected checks whether err wraps ErrPasskeyCloneDetected.
func isCloneDetected(err error) bool {
	type singleUnwrap interface{ Unwrap() error }
	type multiUnwrap interface{ Unwrap() []error }
	for err != nil {
		if err == model.ErrPasskeyCloneDetected {
			return true
		}
		if u, ok := err.(multiUnwrap); ok {
			for _, e := range u.Unwrap() {
				if isCloneDetected(e) {
					return true
				}
			}
			return false
		}
		if u, ok := err.(singleUnwrap); ok {
			err = u.Unwrap()
		} else {
			break
		}
	}
	return false
}
