// Package page provides full-page view constructors.
package page

import (
	"starter/internal/model"
	"starter/internal/routes"
	"starter/internal/view/layout"
	uiform "starter/pkg/components/form"
	pkgform "starter/pkg/components/patterns/form"
	"starter/pkg/components/ui/core"
	uidata "starter/pkg/components/ui/data"
	"starter/pkg/components/ui/feedback"
	uilayout "starter/pkg/components/ui/layout"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// LoginProps configures the login page.
type LoginProps struct {
	Form            *pkgform.Submission
	Input           model.LoginInput
	CSRFToken       string
	IsAuthenticated bool
	NavConfig       uilayout.NavConfig
}

// LoginPage renders the full login page inside the base layout.
func LoginPage(p LoginProps) g.Node {
	return layout.Page(layout.Props{
		Title:           "Login",
		CSRFToken:       p.CSRFToken,
		IsAuthenticated: p.IsAuthenticated,
		NavConfig:       p.NavConfig,
		Children: []g.Node{
			h.Div(
				h.Class("max-w-sm mx-auto mt-8 sm:mt-12 px-4"),
				uidata.Card.Root(
					uidata.Card.Header(uidata.Card.Title(g.Text("Sign In"))),
					uidata.Card.Content(loginForm(p)),
				),
			),
		},
	})
}

func loginForm(p LoginProps) g.Node {
	var emailErrors, passwordErrors, formErrors []string
	if p.Form != nil {
		emailErrors = p.Form.GetFieldErrors("Email")
		passwordErrors = p.Form.GetFieldErrors("Password")
		formErrors = p.Form.GetFormErrors()
	}
	return h.Form(
		h.Method("post"),
		h.Action(routes.Login),
		h.Class("w-full flex flex-col gap-3"),
		h.Input(h.Type("hidden"), h.Name("_csrf"), h.Value(p.CSRFToken)),
		g.If(len(formErrors) > 0, formNotice(formErrors)),
		uiform.Field(uiform.FieldProps{
			ID:     "email",
			Label:  "Email",
			Errors: emailErrors,
			Control: uiform.Input(uiform.InputProps{
				ID:       "email",
				Name:     "email",
				Type:     uiform.TypeEmail,
				Value:    p.Input.Email,
				Required: true,
				HasError: len(emailErrors) > 0,
				Extra:    uiform.FieldControlAttrs("email", "", "", emailErrors),
			}),
		}),
		uiform.Field(uiform.FieldProps{
			ID:     "password",
			Label:  "Password",
			Errors: passwordErrors,
			Control: uiform.Input(uiform.InputProps{
				ID:       "password",
				Name:     "password",
				Type:     uiform.TypePassword,
				Required: true,
				HasError: len(passwordErrors) > 0,
				Extra:    uiform.FieldControlAttrs("password", "", "", passwordErrors),
			}),
		}),
		core.Button(core.ButtonProps{
			Label:     "Sign In",
			Type:      "submit",
			FullWidth: true,
			Extra:     []g.Node{h.Class("mt-2")},
		}),
		h.P(
			h.Class("text-sm text-center mt-2"),
			g.Text("Don't have an account? "),
			h.A(h.Href(routes.Register), h.Class("underline underline-offset-4 hover:text-primary"), g.Text("Register")),
		),
	)
}

// RegisterProps configures the registration page.
type RegisterProps struct {
	Form            *pkgform.Submission
	Input           model.CreateUserInput
	CSRFToken       string
	IsAuthenticated bool
	NavConfig       uilayout.NavConfig
}

// RegisterPage renders the full registration page inside the base layout.
func RegisterPage(p RegisterProps) g.Node {
	return layout.Page(layout.Props{
		Title:           "Register",
		CSRFToken:       p.CSRFToken,
		IsAuthenticated: p.IsAuthenticated,
		NavConfig:       p.NavConfig,
		Children: []g.Node{
			h.Div(
				h.Class("max-w-sm mx-auto mt-8 sm:mt-12 px-4"),
				uidata.Card.Root(
					uidata.Card.Header(uidata.Card.Title(g.Text("Create Account"))),
					uidata.Card.Content(registerForm(p)),
				),
			),
		},
	})
}

func registerForm(p RegisterProps) g.Node {
	var emailErrors, nameErrors, passwordErrors, formErrors []string
	if p.Form != nil {
		emailErrors = p.Form.GetFieldErrors("Email")
		nameErrors = p.Form.GetFieldErrors("DisplayName")
		passwordErrors = p.Form.GetFieldErrors("Password")
		formErrors = p.Form.GetFormErrors()
	}
	return h.Form(
		h.Method("post"),
		h.Action(routes.Register),
		h.Class("w-full flex flex-col gap-3"),
		h.Input(h.Type("hidden"), h.Name("_csrf"), h.Value(p.CSRFToken)),
		g.If(len(formErrors) > 0, formNotice(formErrors)),
		uiform.Field(uiform.FieldProps{
			ID:     "email",
			Label:  "Email",
			Errors: emailErrors,
			Control: uiform.Input(uiform.InputProps{
				ID:       "email",
				Name:     "email",
				Type:     uiform.TypeEmail,
				Value:    p.Input.Email,
				Required: true,
				HasError: len(emailErrors) > 0,
				Extra:    uiform.FieldControlAttrs("email", "", "", emailErrors),
			}),
		}),
		uiform.Field(uiform.FieldProps{
			ID:     "display_name",
			Label:  "Display Name",
			Errors: nameErrors,
			Control: uiform.Input(uiform.InputProps{
				ID:       "display_name",
				Name:     "display_name",
				Value:    p.Input.DisplayName,
				Required: true,
				HasError: len(nameErrors) > 0,
				Extra:    uiform.FieldControlAttrs("display_name", "", "", nameErrors),
			}),
		}),
		uiform.Field(uiform.FieldProps{
			ID:     "password",
			Label:  "Password",
			Errors: passwordErrors,
			Control: uiform.Input(uiform.InputProps{
				ID:       "password",
				Name:     "password",
				Type:     uiform.TypePassword,
				Required: true,
				HasError: len(passwordErrors) > 0,
				Extra:    uiform.FieldControlAttrs("password", "", "", passwordErrors),
			}),
		}),
		core.Button(core.ButtonProps{
			Label:     "Create Account",
			Type:      "submit",
			FullWidth: true,
			Extra:     []g.Node{h.Class("mt-2")},
		}),
		h.P(
			h.Class("text-sm text-center mt-2"),
			g.Text("Already have an account? "),
			h.A(h.Href(routes.Login), h.Class("underline underline-offset-4 hover:text-primary"), g.Text("Sign in")),
		),
	)
}

func formNotice(messages []string) g.Node {
	items := make([]g.Node, len(messages))
	for i, message := range messages {
		items[i] = h.Li(g.Text(message))
	}
	return feedback.Alert.Root(
		feedback.AlertProps{Variant: feedback.AlertDestructive},
		feedback.Alert.Description(
			h.Ul(h.Class("list-disc space-y-1 pl-4"), g.Group(items)),
		),
	)
}
