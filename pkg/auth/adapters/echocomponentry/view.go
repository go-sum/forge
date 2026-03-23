package echocomponentry

import (
	"github.com/go-sum/auth/model"
	uiform "github.com/go-sum/componentry/form"
	"github.com/go-sum/componentry/patterns/form"
	"github.com/go-sum/componentry/ui/core"
	uidata "github.com/go-sum/componentry/ui/data"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// FormError renders a destructive alert listing validation messages.
func FormError(messages []string) g.Node {
	if len(messages) == 0 {
		return g.Text("")
	}
	items := make([]g.Node, len(messages))
	for i, msg := range messages {
		items[i] = h.Li(g.Text(msg))
	}
	return h.Div(
		h.Class("rounded-md border border-destructive bg-destructive/10 px-4 py-3 text-sm text-destructive"),
		h.Ul(h.Class("list-disc space-y-1 pl-4"), g.Group(items)),
	)
}

// SigninPage renders the full signin page inside the host layout.
func SigninPage(req Request, submission *form.Submission, input model.SigninInput, signinPath, signupPath, csrfField string) g.Node {
	return req.Page(
		"Signin",
		h.Div(
			h.Class("mx-auto w-full max-w-sm px-4 py-12 sm:py-16"),
			uidata.Card.Root(
				uidata.Card.Header(
					uidata.Card.Title(g.Text("Sign In")),
					uidata.Card.Description(g.Text("Enter your account details to continue into the application.")),
				),
				uidata.Card.Content(signinForm(req, submission, input, signinPath, signupPath, csrfField)),
			),
		),
	)
}

func signinForm(req Request, submission *form.Submission, input model.SigninInput, signinPath, signupPath, csrfField string) g.Node {
	var emailErrors, passwordErrors, formErrors []string
	if submission != nil {
		emailErrors = submission.GetFieldErrors("Email")
		passwordErrors = submission.GetFieldErrors("Password")
		formErrors = submission.GetFormErrors()
	}
	return h.Form(
		h.Method("post"),
		h.Action(signinPath),
		h.Class("w-full flex flex-col gap-4"),
		h.Input(h.Type("hidden"), h.Name(csrfField), h.Value(req.CSRFToken)),
		g.If(len(formErrors) > 0, FormError(formErrors)),
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
			h.A(h.Href(signupPath), h.Class("underline underline-offset-4 hover:text-primary"), g.Text("Signup")),
		),
	)
}

// SignupPage renders the full registration page inside the host layout.
func SignupPage(req Request, submission *form.Submission, input model.SignupInput, signinPath, signupPath, csrfField string) g.Node {
	return req.Page(
		"Signup",
		h.Div(
			h.Class("mx-auto w-full max-w-sm px-4 py-12 sm:py-16"),
			uidata.Card.Root(
				uidata.Card.Header(
					uidata.Card.Title(g.Text("Create Account")),
					uidata.Card.Description(g.Text("Set up an account so you can start working with the app.")),
				),
				uidata.Card.Content(signupForm(req, submission, input, signinPath, signupPath, csrfField)),
			),
		),
	)
}

func signupForm(req Request, submission *form.Submission, input model.SignupInput, signinPath, signupPath, csrfField string) g.Node {
	var emailErrors, nameErrors, passwordErrors, formErrors []string
	if submission != nil {
		emailErrors = submission.GetFieldErrors("Email")
		nameErrors = submission.GetFieldErrors("DisplayName")
		passwordErrors = submission.GetFieldErrors("Password")
		formErrors = submission.GetFormErrors()
	}
	return h.Form(
		h.Method("post"),
		h.Action(signupPath),
		h.Class("w-full flex flex-col gap-4"),
		h.Input(h.Type("hidden"), h.Name(csrfField), h.Value(req.CSRFToken)),
		g.If(len(formErrors) > 0, FormError(formErrors)),
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
			h.A(h.Href(signinPath), h.Class("underline underline-offset-4 hover:text-primary"), g.Text("Sign in")),
		),
	)
}
