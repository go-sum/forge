// Package page provides full-page view constructors.
package page

import (
	"starter/internal/model"
	"starter/internal/routes"
	"starter/internal/view"
	uiform "starter/pkg/components/form"
	pkgform "starter/pkg/components/patterns/form"
	"starter/pkg/components/ui/core"
	uidata "starter/pkg/components/ui/data"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// LoginPage renders the full login page inside the base layout.
func LoginPage(req view.Request, form *pkgform.Submission, input model.LoginInput) g.Node {
	return req.Page(
		"Login",
		h.Div(
			h.Class("mx-auto w-full max-w-sm px-4 py-12 sm:py-16"),
			uidata.Card.Root(
				uidata.Card.Header(
					uidata.Card.Title(g.Text("Sign In")),
					uidata.Card.Description(g.Text("Enter your account details to continue into the application.")),
				),
				uidata.Card.Content(loginForm(req, form, input)),
			),
		),
	)
}

func loginForm(req view.Request, form *pkgform.Submission, input model.LoginInput) g.Node {
	var emailErrors, passwordErrors, formErrors []string
	if form != nil {
		emailErrors = form.GetFieldErrors("Email")
		passwordErrors = form.GetFieldErrors("Password")
		formErrors = form.GetFormErrors()
	}
	return h.Form(
		h.Method("post"),
		h.Action(routes.Login),
		h.Class("w-full flex flex-col gap-4"),
		h.Input(h.Type("hidden"), h.Name("_csrf"), h.Value(req.CSRFToken)),
		g.If(len(formErrors) > 0, view.FormError(formErrors)),
		uiform.Field(uiform.FieldProps{
			ID:     "email",
			Label:  "Email",
			Errors: emailErrors,
			Control: uiform.Input(uiform.InputProps{
				ID:       "email",
				Name:     "email",
				Type:     uiform.TypeEmail,
				Value:    input.Email,
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
			h.Class("pt-1 text-center text-sm text-muted-foreground"),
			g.Text("Don't have an account? "),
			h.A(h.Href(routes.Register), h.Class("underline underline-offset-4 hover:text-primary"), g.Text("Register")),
		),
	)
}

// RegisterPage renders the full registration page inside the base layout.
func RegisterPage(req view.Request, form *pkgform.Submission, input model.CreateUserInput) g.Node {
	return req.Page(
		"Register",
		h.Div(
			h.Class("mx-auto w-full max-w-sm px-4 py-12 sm:py-16"),
			uidata.Card.Root(
				uidata.Card.Header(
					uidata.Card.Title(g.Text("Create Account")),
					uidata.Card.Description(g.Text("Set up an account so you can start working with the starter's app flows.")),
				),
				uidata.Card.Content(registerForm(req, form, input)),
			),
		),
	)
}

func registerForm(req view.Request, form *pkgform.Submission, input model.CreateUserInput) g.Node {
	var emailErrors, nameErrors, passwordErrors, formErrors []string
	if form != nil {
		emailErrors = form.GetFieldErrors("Email")
		nameErrors = form.GetFieldErrors("DisplayName")
		passwordErrors = form.GetFieldErrors("Password")
		formErrors = form.GetFormErrors()
	}
	return h.Form(
		h.Method("post"),
		h.Action(routes.Register),
		h.Class("w-full flex flex-col gap-4"),
		h.Input(h.Type("hidden"), h.Name("_csrf"), h.Value(req.CSRFToken)),
		g.If(len(formErrors) > 0, view.FormError(formErrors)),
		uiform.Field(uiform.FieldProps{
			ID:     "email",
			Label:  "Email",
			Errors: emailErrors,
			Control: uiform.Input(uiform.InputProps{
				ID:       "email",
				Name:     "email",
				Type:     uiform.TypeEmail,
				Value:    input.Email,
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
				Value:    input.DisplayName,
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
			h.Class("pt-1 text-center text-sm text-muted-foreground"),
			g.Text("Already have an account? "),
			h.A(h.Href(routes.Login), h.Class("underline underline-offset-4 hover:text-primary"), g.Text("Sign in")),
		),
	)
}
