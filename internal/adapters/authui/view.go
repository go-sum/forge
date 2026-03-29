package authui

import (
	"fmt"

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
func SigninPage(req Request, submission *form.Submission, input model.BeginSigninInput, signinPath, signupPath, csrfField string) g.Node {
	return req.Page(
		"Sign In",
		authCard("Sign In", "Enter your email address and we'll send a verification code.", signinForm(req, submission, input, signinPath, signupPath, csrfField)),
	)
}

func signinForm(req Request, submission *form.Submission, input model.BeginSigninInput, signinPath, signupPath, csrfField string) g.Node {
	var emailErrors, formErrors []string
	if submission != nil {
		emailErrors = submission.GetFieldErrors("Email")
		formErrors = submission.GetFormErrors()
	}
	return h.Form(
		h.Method("post"),
		h.Action(signinPath),
		h.Class("w-full flex flex-col gap-4"),
		h.Input(h.Type("hidden"), h.Name(csrfField), h.Value(req.CSRFToken)),
		g.If(len(formErrors) > 0, FormError(formErrors)),
		emailField(input.Email, emailErrors),
		core.Button(core.ButtonProps{
			Label:     "Send Code",
			Type:      "submit",
			FullWidth: true,
			Extra:     []g.Node{h.Class("mt-2")},
		}),
		h.P(
			h.Class("pt-1 text-center text-sm text-muted-foreground"),
			g.Text("Need an account? "),
			h.A(h.Href(signupPath), h.Class("underline underline-offset-4 hover:text-primary"), g.Text("Sign up")),
		),
	)
}

// SignupPage renders the full registration page inside the host layout.
func SignupPage(req Request, submission *form.Submission, input model.BeginSignupInput, signinPath, signupPath, csrfField string) g.Node {
	return req.Page(
		"Sign Up",
		authCard("Create Account", "Enter your details and we'll send a verification code.", signupForm(req, submission, input, signinPath, signupPath, csrfField)),
	)
}

func signupForm(req Request, submission *form.Submission, input model.BeginSignupInput, signinPath, signupPath, csrfField string) g.Node {
	var emailErrors, nameErrors, formErrors []string
	if submission != nil {
		emailErrors = submission.GetFieldErrors("Email")
		nameErrors = submission.GetFieldErrors("DisplayName")
		formErrors = submission.GetFormErrors()
	}
	return h.Form(
		h.Method("post"),
		h.Action(signupPath),
		h.Class("w-full flex flex-col gap-4"),
		h.Input(h.Type("hidden"), h.Name(csrfField), h.Value(req.CSRFToken)),
		g.If(len(formErrors) > 0, FormError(formErrors)),
		emailField(input.Email, emailErrors),
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
		core.Button(core.ButtonProps{
			Label:     "Send Signup Code",
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

// VerifyPage renders the shared verification screen.
func VerifyPage(
	req Request,
	submission *form.Submission,
	input model.VerifyInput,
	state model.VerifyPageState,
	stateErrors []string,
	verifyPath, resendPath, csrfField string,
) g.Node {
	return req.Page(
		"Verify",
		authCard("Verify Code", verifyDescription(state), verifyContent(req, submission, input, state, stateErrors, verifyPath, resendPath, csrfField)),
	)
}

func verifyContent(
	req Request,
	submission *form.Submission,
	input model.VerifyInput,
	state model.VerifyPageState,
	stateErrors []string,
	verifyPath, resendPath, csrfField string,
) g.Node {
	return h.Div(
		h.Class("w-full flex flex-col gap-4"),
		verifyForm(req, submission, input, state, stateErrors, verifyPath, csrfField),
		g.If(state.CanResend, resendForm(req, resendPath, csrfField)),
	)
}

func verifyForm(
	req Request,
	submission *form.Submission,
	input model.VerifyInput,
	state model.VerifyPageState,
	stateErrors []string,
	verifyPath, csrfField string,
) g.Node {
	var codeErrors, formErrors []string
	if submission != nil {
		codeErrors = submission.GetFieldErrors("Code")
		formErrors = submission.GetFormErrors()
	}
	formErrors = append(formErrors, stateErrors...)
	value := input.Code
	if value == "" {
		value = state.Code
	}

	return h.Form(
		h.Method("post"),
		h.Action(verifyPath),
		h.Class("w-full flex flex-col gap-4"),
		h.Input(h.Type("hidden"), h.Name(csrfField), h.Value(req.CSRFToken)),
		h.Input(h.Type("hidden"), h.Name("token"), h.Value(input.Token)),
		g.If(len(formErrors) > 0, FormError(formErrors)),
		uiform.Field(uiform.FieldProps{
			ID:     "code",
			Label:  "Verification Code",
			Errors: codeErrors,
			Control: uiform.Input(uiform.InputProps{
				ID:          "code",
				Name:        "code",
				Value:       value,
				Required:    true,
				HasError:    len(codeErrors) > 0,
				Placeholder: "123456",
				Extra:       uiform.FieldControlAttrs("code", "", "", codeErrors),
			}),
		}),
		g.If(state.Email != "", h.P(h.Class("text-sm text-muted-foreground"), g.Text(fmt.Sprintf("Verification target: %s", state.Email)))),
		core.Button(core.ButtonProps{
			Label:     "Verify",
			Type:      "submit",
			FullWidth: true,
			Extra:     []g.Node{h.Class("mt-2")},
		}),
	)
}

func resendForm(req Request, resendPath, csrfField string) g.Node {
	return h.Form(
		h.Method("post"),
		h.Action(resendPath),
		h.Class("w-full"),
		h.Input(h.Type("hidden"), h.Name(csrfField), h.Value(req.CSRFToken)),
		core.Button(core.ButtonProps{
			Label:     "Resend Code",
			Type:      "submit",
			Variant:   core.VariantSecondary,
			FullWidth: true,
		}),
	)
}

// EmailChangePage renders the self-service email change form.
func EmailChangePage(req Request, submission *form.Submission, input model.BeginEmailChangeInput, actionPath, csrfField string) g.Node {
	return req.Page(
		"Change Email",
		authCard("Change Email", "Enter your new email address and we'll send a verification code there.", emailChangeForm(req, submission, input, actionPath, csrfField)),
	)
}

func emailChangeForm(req Request, submission *form.Submission, input model.BeginEmailChangeInput, actionPath, csrfField string) g.Node {
	var emailErrors, formErrors []string
	if submission != nil {
		emailErrors = submission.GetFieldErrors("Email")
		formErrors = submission.GetFormErrors()
	}
	return h.Form(
		h.Method("post"),
		h.Action(actionPath),
		h.Class("w-full flex flex-col gap-4"),
		h.Input(h.Type("hidden"), h.Name(csrfField), h.Value(req.CSRFToken)),
		g.If(len(formErrors) > 0, FormError(formErrors)),
		emailField(input.Email, emailErrors),
		core.Button(core.ButtonProps{
			Label:     "Send Verification Code",
			Type:      "submit",
			FullWidth: true,
			Extra:     []g.Node{h.Class("mt-2")},
		}),
	)
}

func authCard(title, description string, content g.Node) g.Node {
	return h.Div(
		h.Class("mx-auto w-full max-w-sm px-4 py-12 sm:py-16"),
		uidata.Card.Root(
			uidata.Card.Header(
				uidata.Card.Title(g.Text(title)),
				uidata.Card.Description(g.Text(description)),
			),
			uidata.Card.Content(content),
		),
	)
}

func emailField(value string, errors []string) g.Node {
	return uiform.Field(uiform.FieldProps{
		ID:     "email",
		Label:  "Email",
		Errors: errors,
		Control: uiform.Input(uiform.InputProps{
			ID:       "email",
			Name:     "email",
			Type:     uiform.TypeEmail,
			Value:    value,
			Required: true,
			HasError: len(errors) > 0,
			Extra:    uiform.FieldControlAttrs("email", "", "", errors),
		}),
	})
}

func verifyDescription(state model.VerifyPageState) string {
	switch state.Purpose {
	case model.FlowPurposeSignup:
		return "Enter the 6-digit code from your signup email to finish creating your account."
	case model.FlowPurposeEmailChange:
		return "Enter the 6-digit code we sent to your new email address to confirm the change."
	default:
		return "Enter the 6-digit code from your email to finish signing in."
	}
}
