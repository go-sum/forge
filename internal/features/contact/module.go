package contact

import (
	"github.com/go-sum/componentry/patterns/form"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/queue"
)

type Module struct {
	handler *Handler
}

func NewModule(cfg *config.Config, q *queue.Client, validator form.StructValidator, contactCfg Config) *Module {
	return &Module{
		handler: NewHandler(cfg, NewService(q, contactCfg), validator),
	}
}

func NewModuleWithHandler(handler *Handler) *Module {
	return &Module{handler: handler}
}

