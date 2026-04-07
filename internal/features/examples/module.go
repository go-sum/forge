package examples

import (
	componentexamples "github.com/go-sum/componentry/examples"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/view"

	"github.com/labstack/echo/v5"
)

// Module registers the component gallery reference surface.
type Module struct {
	cfg *config.Config
}

func NewModule(cfg *config.Config) *Module {
	return &Module{cfg: cfg}
}

func (m *Module) Handle(c *echo.Context) error {
	req := view.NewRequest(c, m.cfg)
	return view.Render(c, req, req.Page("Component Examples", componentexamples.Page()), nil)
}
