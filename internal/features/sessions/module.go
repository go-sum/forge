package sessions

import (
	"github.com/go-sum/forge/config"
	"github.com/go-sum/server/validate"
	"github.com/go-sum/session"
)

type Module struct {
	handler *Handler
}

func NewModule(cfg *config.Config, mgr session.Manager, _ *validate.Validator) *Module {
	mm, _ := mgr.(session.MultiManager)
	return &Module{handler: NewHandler(cfg, mm)}
}

func (m *Module) Handler() *Handler { return m.handler }
