package htmx

import (
	"net/url"
	"strconv"

	"starter/pkg/components/ui/feedback"

	g "maragu.dev/gomponents"
)

const (
	SwapInnerHTML = "innerHTML"
	SwapOuterHTML = "outerHTML"
	SwapBeforeEnd = "beforeend"
)

// LiveSearchProps configures an input that fetches server-rendered results as the user types.
type LiveSearchProps struct {
	Path        string
	Target      string
	Swap        string
	Trigger     string
	Delay       string
	Include     string
	Indicator   string
	DisabledElt string
	PushURL     bool
}

func LiveSearch(p LiveSearchProps) []g.Node {
	trigger := p.Trigger
	if trigger == "" {
		delay := p.Delay
		if delay == "" {
			delay = "300ms"
		}
		trigger = "input changed delay:" + delay + ", search"
	}

	props := AttrsProps{
		Get:         p.Path,
		Target:      p.Target,
		Swap:        orDefault(p.Swap, SwapInnerHTML),
		Trigger:     trigger,
		Include:     p.Include,
		Indicator:   p.Indicator,
		DisabledElt: p.DisabledElt,
	}
	if p.PushURL {
		props.PushURL = "true"
	}
	return Attrs(props)
}

// InlineValidationProps configures a field that validates server-side on change/blur.
type InlineValidationProps struct {
	Path        string
	Target      string
	Swap        string
	Trigger     string
	Include     string
	Indicator   string
	DisabledElt string
	Sync        string
}

func InlineValidation(p InlineValidationProps) []g.Node {
	trigger := p.Trigger
	if trigger == "" {
		trigger = "change delay:200ms, blur"
	}
	sync := p.Sync
	if sync == "" {
		sync = "closest form:abort"
	}

	return Attrs(AttrsProps{
		Get:         p.Path,
		Target:      p.Target,
		Swap:        orDefault(p.Swap, SwapOuterHTML),
		Trigger:     trigger,
		Include:     p.Include,
		Indicator:   p.Indicator,
		DisabledElt: p.DisabledElt,
		Sync:        sync,
	})
}

// PaginatedTableProps configures a link or button that swaps a server-rendered table region.
type PaginatedTableProps struct {
	Path        string
	Page        int
	PageParam   string
	Query       map[string]string
	Target      string
	Swap        string
	Include     string
	Indicator   string
	DisabledElt string
	PushURL     bool
}

func PaginatedTableLink(p PaginatedTableProps) []g.Node {
	path := withQueryParam(p.Path, orDefault(p.PageParam, "page"), strconv.Itoa(p.Page), p.Query)
	props := AttrsProps{
		Get:         path,
		Target:      p.Target,
		Swap:        orDefault(p.Swap, SwapOuterHTML),
		Include:     p.Include,
		Indicator:   p.Indicator,
		DisabledElt: p.DisabledElt,
	}
	if p.PushURL {
		props.PushURL = "true"
	}
	return Attrs(props)
}

// AsyncDialogProps configures a trigger that opens a native dialog and fetches its body asynchronously.
type AsyncDialogProps struct {
	Path        string
	DialogID    string
	Target      string
	Swap        string
	Select      string
	Indicator   string
	DisabledElt string
}

func AsyncDialogTrigger(p AsyncDialogProps) []g.Node {
	nodes := []g.Node{
		g.Attr("data-dialog-open", p.DialogID),
		g.Attr("aria-haspopup", "dialog"),
		g.Attr("aria-controls", p.DialogID),
	}
	nodes = append(nodes, Attrs(AttrsProps{
		Get:         p.Path,
		Target:      p.Target,
		Swap:        orDefault(p.Swap, SwapInnerHTML),
		Select:      p.Select,
		Indicator:   p.Indicator,
		DisabledElt: p.DisabledElt,
	})...)
	return nodes
}

// OOBSwapProps configures an out-of-band swap attribute.
type OOBSwapProps struct {
	Strategy string
	Selector string
}

func OOBSwap(p OOBSwapProps) []g.Node {
	value := p.Strategy
	if value == "" {
		value = "true"
	}
	if p.Selector != "" {
		if value == "true" {
			value = SwapOuterHTML
		}
		value += ":" + p.Selector
	}
	return []g.Node{g.Attr("hx-swap-oob", value)}
}

func OOBAppend(selector string) []g.Node {
	return OOBSwap(OOBSwapProps{Strategy: SwapBeforeEnd, Selector: selector})
}

// ToastOOBProps wraps a feedback.Toast for out-of-band insertion into a toast container.
type ToastOOBProps struct {
	Toast    feedback.ToastProps
	Selector string
	Strategy string
}

func ToastOOB(p ToastOOBProps) g.Node {
	toast := p.Toast
	selector := p.Selector
	if selector == "" {
		selector = "#toast-container"
	}
	extra := append([]g.Node{}, OOBSwap(OOBSwapProps{
		Strategy: orDefault(p.Strategy, SwapBeforeEnd),
		Selector: selector,
	})...)
	toast.Extra = append(extra, toast.Extra...)
	return feedback.Toast(toast)
}

// DependentSelectProps configures a select that swaps a downstream field when its value changes.
type DependentSelectProps struct {
	Path        string
	Target      string
	Swap        string
	Trigger     string
	Include     string
	Indicator   string
	DisabledElt string
}

func DependentSelect(p DependentSelectProps) []g.Node {
	trigger := p.Trigger
	if trigger == "" {
		trigger = "change"
	}
	return Attrs(AttrsProps{
		Get:         p.Path,
		Target:      p.Target,
		Swap:        orDefault(p.Swap, SwapOuterHTML),
		Trigger:     trigger,
		Include:     p.Include,
		Indicator:   p.Indicator,
		DisabledElt: p.DisabledElt,
	})
}

func withQueryParam(path, key, value string, extras map[string]string) string {
	parsed, err := url.Parse(path)
	if err != nil {
		return path
	}
	query := parsed.Query()
	for name, extra := range extras {
		query.Set(name, extra)
	}
	query.Set(key, value)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
