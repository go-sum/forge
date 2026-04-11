package auth

import (
	"github.com/go-sum/auth/model"
	"github.com/labstack/echo/v5"
	g "maragu.dev/gomponents"
)

// PasskeyHandlerConfig configures the PasskeyHandler.
type PasskeyHandlerConfig struct {
	Sessions  SessionManager
	Pages     PasskeyPageRenderer
	HomePath  string
	HomePathFn func() string
	RequestFn func(c *echo.Context) Request
}

// PasskeyPageRenderer renders passkey management pages.
type PasskeyPageRenderer interface {
	PasskeyListPage(req Request, data PasskeyListData) g.Node
	PasskeyListRegion(req Request, data PasskeyListData) g.Node
	PasskeyRow(req Request, data PasskeyRowData) g.Node
	PasskeyEditForm(req Request, data PasskeyRowData) g.Node
}

// PasskeyListData carries data for the passkey list page/region.
type PasskeyListData struct {
	Passkeys  []model.PasskeyCredential
	CSRFToken string
}

// PasskeyRowData carries data for a single passkey row or edit form.
type PasskeyRowData struct {
	Passkey   model.PasskeyCredential
	CSRFToken string
}
