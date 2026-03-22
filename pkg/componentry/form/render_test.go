package form

import (
	"strings"
	"testing"

	testutil "github.com/go-sum/componentry/testutil"
)

func TestFieldRendersDescriptionAndErrorsWithControlWiring(t *testing.T) {
	errors := []string{"Email is required"}
	got := testutil.RenderNode(t, Field(FieldProps{
		ID:          "email",
		Label:       "Email",
		Description: "Used to sign in.",
		Hint:        "We never share it.",
		Errors:      errors,
		Control: Input(InputProps{
			ID:       "email",
			Name:     "email",
			Type:     TypeEmail,
			HasError: true,
			Extra:    FieldControlAttrs("email", "Used to sign in.", "We never share it.", errors),
		}),
	}))

	checks := []string{
		`<label class="text-sm font-medium leading-none inline-block text-destructive" for="email">Email</label>`,
		` aria-describedby="email-description email-hint email-error"`,
		` aria-errormessage="email-error"`,
		`<p class="text-sm text-muted-foreground" id="email-description">Used to sign in.</p>`,
		`<p class="text-xs text-muted-foreground" id="email-hint">We never share it.</p>`,
		`<div class="grid gap-1" id="email-error"><p class="text-xs text-destructive">Email is required</p></div>`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Field() output missing %q in %s", check, got)
		}
	}
}

func TestCheckboxRadioAndSwitchRenderTypes(t *testing.T) {
	checkbox := testutil.RenderNode(t, Checkbox(CheckboxProps{ID: "terms", Name: "terms", Checked: true}))
	if !strings.Contains(checkbox, ` type="checkbox"`) || !strings.Contains(checkbox, ` checked`) {
		t.Fatalf("Checkbox() output = %s", checkbox)
	}

	radio := testutil.RenderNode(t, Radio(RadioProps{ID: "role-admin", Name: "role", Value: "admin"}))
	if !strings.Contains(radio, ` type="radio"`) || !strings.Contains(radio, ` value="admin"`) {
		t.Fatalf("Radio() output = %s", radio)
	}

	switchControl := testutil.RenderNode(t, Switch(SwitchProps{ID: "notifications", Name: "notifications", Checked: true}))
	if !strings.Contains(switchControl, ` role="switch"`) || !strings.Contains(switchControl, ` class="sr-only peer"`) {
		t.Fatalf("Switch() output = %s", switchControl)
	}
}

func TestFieldSetRendersLegendDescriptionAndErrors(t *testing.T) {
	got := testutil.RenderNode(t, FieldSet(FieldSetProps{
		ID:          "notify",
		Legend:      "Notifications",
		Description: "Pick one.",
		Errors:      []string{"Required"},
	}))

	checks := []string{
		`<fieldset`,
		`<legend`,
		`>Notifications<`,
		`id="notify-description"`,
		`id="notify-error"`,
		`aria-describedby="notify-description notify-error"`,
		`aria-errormessage="notify-error"`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("FieldSet() output missing %q in %s", check, got)
		}
	}
}

func TestFieldSetDisabledPropagates(t *testing.T) {
	got := testutil.RenderNode(t, FieldSet(FieldSetProps{Legend: "Group", Disabled: true}))
	if !strings.Contains(got, `<fieldset`) {
		t.Fatalf("FieldSet() output missing <fieldset in %s", got)
	}
	if !strings.Contains(got, ` disabled`) {
		t.Fatalf("FieldSet() output missing disabled attribute in %s", got)
	}
}

func TestTextareaRendersRowsAndErrorState(t *testing.T) {
	got := testutil.RenderNode(t, Textarea(TextareaProps{
		ID:       "bio",
		Name:     "bio",
		Rows:     4,
		HasError: true,
		Value:    "Hello",
	}))

	checks := []string{` id="bio"`, ` name="bio"`, ` rows="4"`, ` aria-invalid="true"`, `>Hello</textarea>`}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Textarea() output missing %q in %s", check, got)
		}
	}
}
