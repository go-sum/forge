package render

import (
	componentassets "starter/pkg/components/assets"
	componenticons "starter/pkg/components/icons"
	"starter/pkg/components/ui/core"
)

// Props returns IconProps for a concrete sprite symbol.
func Props(spriteKey, symbolID string, p core.IconProps) core.IconProps {
	p.Src = componentassets.SpritePath(spriteKey)
	p.ID = symbolID
	return p
}

// PropsForRegistry returns IconProps for a semantic component icon key.
// Unknown keys return props with empty Src and ID so callers fail softly.
func PropsForRegistry(r *componenticons.Registry, key componenticons.Key, p core.IconProps) core.IconProps {
	if r == nil {
		p.Src = ""
		p.ID = ""
		return p
	}

	ref, ok := r.Resolve(key)
	if !ok {
		p.Src = ""
		p.ID = ""
		return p
	}

	return Props(ref.Sprite, ref.ID, p)
}

// PropsFor returns IconProps for a semantic component icon key on icons.Default.
func PropsFor(key componenticons.Key, p core.IconProps) core.IconProps {
	return PropsForRegistry(componenticons.Default, key, p)
}
