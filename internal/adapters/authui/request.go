package authui

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Request carries the render-time context needed by auth pages.
// The host application provides PageFn to wrap auth content in its layout.
type Request struct {
	CSRFToken string
	PageFn    func(title string, children ...g.Node) g.Node
}

// Page delegates to PageFn for layout rendering.
func (r Request) Page(title string, children ...g.Node) g.Node {
	if r.PageFn == nil {
		return h.Div(children...)
	}
	return r.PageFn(title, children...)
}
