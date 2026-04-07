package public

import (
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/forge/internal/view/page"
	sitehandlers "github.com/go-sum/site/handlers"

	"github.com/labstack/echo/v5"
)

type Module struct {
	cfg  *config.Config
	site *sitehandlers.Handler
}

func NewModule(cfg *config.Config, site *sitehandlers.Handler) *Module {
	return &Module{cfg: cfg, site: site}
}

func (m *Module) Home(c *echo.Context) error {
	req := view.NewRequest(c, m.cfg)
	return view.Render(c, req, page.HomePage(req), nil)
}
