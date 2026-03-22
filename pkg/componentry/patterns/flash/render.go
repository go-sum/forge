package flash

import (
	"github.com/go-sum/componentry/ui/feedback"

	g "maragu.dev/gomponents"
)

// Render maps flash messages to dismissible Toast components for injection into
// the toast container. Returns g.Text("") when msgs is empty.
func Render(msgs []Message) g.Node {
	if len(msgs) == 0 {
		return g.Text("")
	}
	nodes := make([]g.Node, len(msgs))
	for i, msg := range msgs {
		nodes[i] = feedback.Toast(feedback.ToastProps{
			Description: msg.Text,
			Variant:     toastVariant(msg.Type),
			Dismissible: true,
			// Position "" → container mode, no fixed positioning
		})
	}
	return g.Group(nodes)
}

// RenderOOB maps flash messages to toasts configured for out-of-band insertion
// into #toast-container. This keeps flash rendering server-first while allowing
// HTMX handlers to append notifications without custom client code.
func RenderOOB(msgs []Message) g.Node {
	if len(msgs) == 0 {
		return g.Text("")
	}
	nodes := make([]g.Node, len(msgs))
	for i, msg := range msgs {
		nodes[i] = feedback.Toast(feedback.ToastProps{
			Description: msg.Text,
			Variant:     toastVariant(msg.Type),
			Dismissible: true,
			Extra:       []g.Node{g.Attr("hx-swap-oob", "beforeend:#toast-container")},
		})
	}
	return g.Group(nodes)
}

func toastVariant(t Type) feedback.ToastVariant {
	switch t {
	case TypeSuccess:
		return feedback.ToastSuccess
	case TypeInfo:
		return feedback.ToastInfo
	case TypeWarning:
		return feedback.ToastWarning
	case TypeError:
		return feedback.ToastError
	default:
		return feedback.ToastDefault
	}
}
