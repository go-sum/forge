package model

// ContactInput holds the validated form data from the contact-us form.
type ContactInput struct {
	Name    string `form:"name"    validate:"required,max=100"`
	Email   string `form:"email"   validate:"required,email,max=255"`
	Message string `form:"message" validate:"required,max=5000"`
}

// ContactFormData bundles the state passed to the contact form view.
type ContactFormData struct {
	Values ContactInput
	Errors map[string][]string
	Sent   bool
}
