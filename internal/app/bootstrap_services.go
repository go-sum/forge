package app

import (
	"fmt"
	"log/slog"

	auth "github.com/go-sum/auth"
	authsvc "github.com/go-sum/auth/service"
	authadapter "github.com/go-sum/forge/internal/adapters/auth"
	"github.com/go-sum/forge/internal/adapters/kvsession"
	"github.com/go-sum/forge/internal/service"
	"github.com/go-sum/send"
	"github.com/go-sum/server/validate"
	"github.com/go-sum/session"
)

// Session/auth bootstrap.
func (c *Container) initAuth() {
	cfg := c.Config.App.Session
	mgr, err := session.NewManager(session.Config{
		Store:      cfg.Store,
		CookieName: cfg.Name,
		AuthKey:    cfg.AuthKey,
		EncryptKey: cfg.EncryptKey,
		MaxAge:     cfg.MaxAge,
		Secure:     cfg.Secure,
		BlobStore:  kvsession.New(c.KV),
	})
	if err != nil {
		panic(fmt.Sprintf("session: %v", err))
	}
	c.Sessions = mgr
	slog.Info("session manager initialized", "store", cfg.Store)
}

// Validation bootstrap.
func (c *Container) initValidator() {
	c.Validator = validate.New()
	c.Web.Validator = c.Validator
}

// Domain services bootstrap.
func (c *Container) initServices() {
	c.registerQueueHandlers()
	c.Services = service.NewServices(c.AuthStore, c.Queue, service.ContactConfig{
		SendTo:   c.Config.Service.Send.SendTo,
		SendFrom: send.DefaultRegistry.SendFrom(c.Config.Service.Send.Delivery),
	})

	if err := auth.ValidateConfig(c.Config.Service.Auth); err != nil {
		panic(fmt.Sprintf("app: auth: %v", err))
	}
	c.AuthService = authsvc.NewAuthService(
		c.AuthStore,
		authsvc.Config{
			Method:   c.Config.Service.Auth.Methods.EmailTOTP,
			Notifier: authadapter.NewNotifier(c.Sender, send.DefaultRegistry.SendFrom(c.Config.Service.Send.Delivery)),
			TokenCodec: authsvc.NewEncryptedTokenCodec(
				c.Config.App.Session.AuthKey,
				c.Config.App.Session.EncryptKey,
			),
		},
	)
}
