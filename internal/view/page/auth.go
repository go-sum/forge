// Package page provides full-page view constructors.
package page

import (
	"starter/internal/view/layout"
	uiform "starter/pkg/components/form"
	pkgform "starter/pkg/components/patterns/form"
	"starter/pkg/components/ui/core"
	uidata "starter/pkg/components/ui/data"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// LoginProps configures the login page.
type LoginProps struct {
	Form      *pkgform.Submission
	CSRFToken string
	ErrorMsg  string
}

// LoginPage renders the full login page inside the base layout.
func LoginPage(p LoginProps) g.Node {
	return layout.Page(layout.Props{
		Title:     "Login",
		CSRFToken: p.CSRFToken,
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
	var emailErrors, passwordErrors []string
	if p.Form != nil {
		emailErrors = p.Form.GetFieldErrors("Email")
		passwordErrors = p.Form.GetFieldErrors("Password")
	}
	return h.Form(
		h.Method("post"),
		h.Action("/login"),
		h.Class("w-full flex flex-col gap-3"),
		h.Input(h.Type("hidden"), h.Name("_csrf"), h.Value(p.CSRFToken)),
		g.If(p.ErrorMsg != "", h.P(h.Class("text-destructive text-sm"), g.Text(p.ErrorMsg))),
		uiform.Field(uiform.FieldProps{
			ID:     "email",
			Label:  "Email",
			Errors: emailErrors,
			Control: uiform.Input(uiform.InputProps{
				ID:       "email",
				Name:     "email",
				Type:     uiform.TypeEmail,
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
			h.A(h.Href("/register"), h.Class("underline underline-offset-4 hover:text-primary"), g.Text("Register")),
		),
	)
}

// RegisterProps configures the registration page.
type RegisterProps struct {
	Form      *pkgform.Submission
	CSRFToken string
}

// RegisterPage renders the full registration page inside the base layout.
func RegisterPage(p RegisterProps) g.Node {
	return layout.Page(layout.Props{
		Title:     "Register",
		CSRFToken: p.CSRFToken,
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
	var emailErrors, nameErrors, passwordErrors []string
	if p.Form != nil {
		emailErrors = p.Form.GetFieldErrors("Email")
		nameErrors = p.Form.GetFieldErrors("DisplayName")
		passwordErrors = p.Form.GetFieldErrors("Password")
	}
	return h.Form(
		h.Method("post"),
		h.Action("/register"),
		h.Class("w-full flex flex-col gap-3"),
		h.Input(h.Type("hidden"), h.Name("_csrf"), h.Value(p.CSRFToken)),
		uiform.Field(uiform.FieldProps{
			ID:     "email",
			Label:  "Email",
			Errors: emailErrors,
			Control: uiform.Input(uiform.InputProps{
				ID:       "email",
				Name:     "email",
				Type:     uiform.TypeEmail,
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
			h.A(h.Href("/login"), h.Class("underline underline-offset-4 hover:text-primary"), g.Text("Sign in")),
		),
	)
}
