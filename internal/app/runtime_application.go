package app

import (
	"fmt"
	"log/slog"

	auth "github.com/go-sum/auth"
	authsvc "github.com/go-sum/auth/service"
	authadapter "github.com/go-sum/forge/internal/adapters/auth"
	"github.com/go-sum/forge/internal/adapters/kvsession"
	"github.com/go-sum/send"
	"github.com/go-sum/server/validate"
	"github.com/go-sum/session"
)

// Session/auth bootstrap.
func (r *Runtime) initAuth() {
	r.Config.Service.Auth = auth.ApplyDefaults(r.Config.Service.Auth)
	cfg := r.Config.Session.Auth
	mgr, err := session.NewManager(session.Config{
		Store:      cfg.Store,
		CookieName: cfg.Name,
		AuthKey:    cfg.AuthKey,
		EncryptKey: cfg.EncryptKey,
		MaxAge:     cfg.MaxAge,
		Secure:     cfg.Secure,
		SameSite:   cfg.SameSite,
		BlobStore:  kvsession.New(r.KV),
	})
	if err != nil {
		panic(fmt.Sprintf("session: %v", err))
	}
	r.Sessions = mgr
	slog.Info("session manager initialized", "store", cfg.Store)

	r.registerQueueHandlers()
	r.initAuthService()
	r.initPasskeyService()
}

// Validation bootstrap.
func (r *Runtime) initValidator() {
	r.Validator = validate.New()
	r.Web.Validator = r.Validator
}

func (r *Runtime) initPasskeyService() {
	cfg := r.Config.Service.Auth.Methods.Passkey
	if !cfg.Enabled {
		return
	}
	svc, err := authsvc.NewPasskeyService(r.AuthStore, r.AuthStore, cfg)
	if err != nil {
		panic(fmt.Sprintf("app: passkey: %v", err))
	}
	r.PasskeyService = svc
	slog.Info("passkey service initialized")
}

func (r *Runtime) initAuthService() {
	r.AuthService = authsvc.NewAuthService(
		r.AuthStore,
		authsvc.Config{
			Method:   r.Config.Service.Auth.Methods.EmailTOTP,
			Notifier: authadapter.NewNotifier(r.Sender, send.DefaultRegistry.SendFrom(r.Config.Service.Send.Delivery)),
			TokenCodec: authsvc.NewEncryptedTokenCodec(
				r.Config.Session.Auth.AuthKey,
				r.Config.Session.Auth.EncryptKey,
			),
		},
	)
}
