// Package contactpartial provides HTMX partial components for the contact-us form.
package contactpartial

import (
	uiform "github.com/go-sum/componentry/form"
	htmx "github.com/go-sum/componentry/patterns/htmx"
	"github.com/go-sum/componentry/ui/core"
	"github.com/go-sum/componentry/ui/feedback"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/view"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// ContactForm renders the contact-us form or a success confirmation when Sent is true.
// It is designed to be swapped in-place by HTMX on both validation errors and success.
func ContactForm(req view.Request, data model.ContactFormData) g.Node {
	return h.Div(
		h.ID("contact-form"),
		g.If(data.Sent, successState()),
		g.If(!data.Sent, formState(req, data)),
	)
}

func successState() g.Node {
	return feedback.Alert.Root(
		feedback.AlertProps{},
		feedback.Alert.Title(g.Text("Message sent!")),
		feedback.Alert.Description(g.Text("Thanks for reaching out. We'll get back to you as soon as possible.")),
	)
}

func formState(req view.Request, data model.ContactFormData) g.Node {
	nameID := "contact-name"
	emailID := "contact-email"
	messageID := "contact-message"

	return h.Form(
		g.Group(htmx.Attrs(htmx.AttrsProps{
			Post:   req.Path("contact.submit"),
			Target: "#contact-form",
			Swap:   htmx.SwapOuterHTML,
		})),
		h.Class("space-y-4"),
		h.Input(h.Type("hidden"), h.Name(req.CSRFFieldName), h.Value(req.CSRFToken)),
		view.FormError(data.Errors["_"]),
		uiform.Field(uiform.FieldProps{
			ID:    nameID,
			Label: "Name",
			Extra: []g.Node{h.Class("w-full")},
			Control: uiform.Input(uiform.InputProps{
				ID:          nameID,
				Name:        "name",
				Value:       data.Values.Name,
				Placeholder: "Your name",
				HasError:    len(data.Errors["Name"]) > 0,
				Extra:       uiform.FieldControlAttrs(nameID, "", "", data.Errors["Name"]),
			}),
			Errors: data.Errors["Name"],
		}),
		uiform.Field(uiform.FieldProps{
			ID:    emailID,
			Label: "Email",
			Extra: []g.Node{h.Class("w-full")},
			Control: uiform.Input(uiform.InputProps{
				ID:          emailID,
				Name:        "email",
				Type:        uiform.TypeEmail,
				Value:       data.Values.Email,
				Placeholder: "you@example.com",
				HasError:    len(data.Errors["Email"]) > 0,
				Extra:       uiform.FieldControlAttrs(emailID, "", "", data.Errors["Email"]),
			}),
			Errors: data.Errors["Email"],
		}),
		uiform.Field(uiform.FieldProps{
			ID:    messageID,
			Label: "Message",
			Extra: []g.Node{h.Class("w-full")},
			Control: uiform.Textarea(uiform.TextareaProps{
				ID:          messageID,
				Name:        "message",
				Value:       data.Values.Message,
				Placeholder: "How can we help?",
				HasError:    len(data.Errors["Message"]) > 0,
				Extra:       uiform.FieldControlAttrs(messageID, "", "", data.Errors["Message"]),
			}),
			Errors: data.Errors["Message"],
		}),
		h.Div(
			h.Class("flex justify-end"),
			core.Button(core.ButtonProps{
				Label: "Send message",
				Type:  "submit",
			}),
		),
	)
}
