package form

import (
	"strings"
	"testing"

	testutil "github.com/go-sum/componentry/testutil"
)

func TestCheckboxRadioAndSwitchExposePeerFocusVisibleStyles(t *testing.T) {
	checkbox := testutil.RenderNode(t, Checkbox(CheckboxProps{ID: "newsletter"}))
	if !strings.Contains(checkbox, `peer-focus-visible:ring-[3px]`) {
		t.Fatalf("Checkbox() output missing peer focus-visible styling: %s", checkbox)
	}

	radio := testutil.RenderNode(t, Radio(RadioProps{ID: "role-admin"}))
	if !strings.Contains(radio, `peer-focus-visible:ring-[3px]`) {
		t.Fatalf("Radio() output missing peer focus-visible styling: %s", radio)
	}

	sw := testutil.RenderNode(t, Switch(SwitchProps{ID: "beta-access"}))
	if !strings.Contains(sw, `peer-focus-visible:ring-[3px]`) {
		t.Fatalf("Switch() output missing peer focus-visible styling: %s", sw)
	}
}
