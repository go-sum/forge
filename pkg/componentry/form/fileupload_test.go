package form

import (
	"strings"
	"testing"

	testutil "github.com/y-goweb/componentry/testutil"
)

func TestFileUploadRendersInputAndDropZone(t *testing.T) {
	got := testutil.RenderNode(t, FileUpload(FileUploadProps{
		ID:     "avatar",
		Name:   "avatar",
		Accept: "image/*",
	}))

	checks := []string{
		`data-file-upload=""`,
		`<input`,
		`type="file"`,
		`id="avatar"`,
		`name="avatar"`,
		`accept="image/*"`,
		`for="avatar"`,
		`data-file-name=""`,
		`aria-live="polite"`,
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("FileUpload() missing %q in:\n%s", want, got)
		}
	}
}

func TestFileUploadMultipleAndError(t *testing.T) {
	got := testutil.RenderNode(t, FileUpload(FileUploadProps{
		ID:       "docs",
		Name:     "docs",
		Multiple: true,
		HasError: true,
	}))

	checks := []string{
		` multiple=""`,
		`aria-invalid="true"`,
		`border-destructive`,
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("FileUpload() missing %q in:\n%s", want, got)
		}
	}
}

func TestFileUploadDisabled(t *testing.T) {
	got := testutil.RenderNode(t, FileUpload(FileUploadProps{
		ID:       "x",
		Name:     "x",
		Disabled: true,
	}))

	checks := []string{
		` disabled`,
		`has-[:disabled]:cursor-not-allowed`,
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("FileUpload() missing %q in:\n%s", want, got)
		}
	}
}
