package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGenerateAndValidateToken(t *testing.T) {
	cfg := JWTConfig{
		Secret:        "secret",
		Issuer:        "starter",
		TokenDuration: time.Hour,
	}
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	token, err := GenerateToken(cfg, userID, "ada@example.com", "admin")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := ValidateToken(cfg, token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.UserID != userID || claims.Email != "ada@example.com" || claims.Role != "admin" {
		t.Fatalf("claims = %#v", claims)
	}
}

func TestValidateTokenRejectsWrongIssuerAndMethod(t *testing.T) {
	cfg := JWTConfig{
		Secret:        "secret",
		Issuer:        "starter",
		TokenDuration: time.Hour,
	}

	wrongIssuer := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: uuid.New(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "other",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	wrongIssuerToken, err := wrongIssuer.SignedString([]byte(cfg.Secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	if _, err := ValidateToken(cfg, wrongIssuerToken); err == nil {
		t.Fatal("ValidateToken() unexpectedly accepted wrong issuer")
	}

	wrongMethod := jwt.NewWithClaims(jwt.SigningMethodHS512, Claims{
		UserID: uuid.New(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	wrongMethodToken, err := wrongMethod.SignedString([]byte(cfg.Secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	if _, err := ValidateToken(cfg, wrongMethodToken); err == nil {
		t.Fatal("ValidateToken() unexpectedly accepted wrong method")
	}
}

func TestValidateTokenRejectsExpiredToken(t *testing.T) {
	cfg := JWTConfig{
		Secret:        "secret",
		Issuer:        "starter",
		TokenDuration: -time.Minute,
	}
	token, err := GenerateToken(cfg, uuid.New(), "ada@example.com", "admin")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if _, err := ValidateToken(cfg, token); err == nil {
		t.Fatal("ValidateToken() unexpectedly accepted expired token")
	}
}
