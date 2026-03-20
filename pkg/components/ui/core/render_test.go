package core

import (
	"strings"
	"testing"

	testutil "starter/pkg/components/testutil"
	g "maragu.dev/gomponents"
)

func TestIconBadgeAndLabelRenderAccessibleMarkup(t *testing.T) {
	icon := testutil.RenderNode(t, Icon(IconProps{Src: "/icons.svg", ID: "check", Label: "Confirmed"}))
	if !strings.Contains(icon, ` role="img"`) || !strings.Contains(icon, ` aria-label="Confirmed"`) || !strings.Contains(icon, `href="/icons.svg#check"`) {
		t.Fatalf("Icon() output = %s", icon)
	}

	badge := testutil.RenderNode(t, Badge(BadgeProps{Variant: BadgeSecondary, Children: []g.Node{g.Text("Beta")}}))
	if !strings.Contains(badge, `bg-secondary`) || !strings.Contains(badge, `>Beta</span>`) {
		t.Fatalf("Badge() output = %s", badge)
	}

	label := testutil.RenderNode(t, Label(LabelProps{For: "email", Error: "Required"}, g.Text("Email")))
	if !strings.Contains(label, ` for="email"`) || !strings.Contains(label, ` text-destructive`) {
		t.Fatalf("Label() output = %s", label)
	}
}

func TestSeparatorAndSkeletonRenderExpectedClasses(t *testing.T) {
	separator := testutil.RenderNode(t, Separator(SeparatorProps{Orientation: OrientationVertical, Decoration: DecorationDashed, Label: "OR"}))
	checks := []string{`aria-orientation="vertical"`, `border-dashed`, `>OR</span>`}
	for _, check := range checks {
		if !strings.Contains(separator, check) {
			t.Fatalf("Separator() output missing %q in %s", check, separator)
		}
	}

	skeleton := testutil.RenderNode(t, Skeleton(g.Attr("data-test", "loading")))
	if !strings.Contains(skeleton, `class="animate-pulse rounded-md bg-muted"`) || !strings.Contains(skeleton, ` data-test="loading"`) {
		t.Fatalf("Skeleton() output = %s", skeleton)
	}
}
