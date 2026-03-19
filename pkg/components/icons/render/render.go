package render

import (
	componentassets "starter/pkg/components/assets"
	componenticons "starter/pkg/components/icons"
	"starter/pkg/components/ui/core"
)

// PropsForAssets returns IconProps for a concrete sprite symbol using the
// provided asset registry.
func PropsForAssets(r *componentassets.Registry, spriteKey, symbolID string, p core.IconProps) core.IconProps {
	if r == nil {
		p.Src = ""
		p.ID = ""
		return p
	}
	p.Src = r.SpritePath(spriteKey)
	p.ID = symbolID
	return p
}

// Props returns IconProps for a concrete sprite symbol on assets.Default.
func Props(spriteKey, symbolID string, p core.IconProps) core.IconProps {
	return PropsForAssets(componentassets.Default, spriteKey, symbolID, p)
}

// PropsForRegistries returns IconProps for a semantic component icon key using
// the provided asset and icon registries.
func PropsForRegistries(assetRegistry *componentassets.Registry, iconRegistry *componenticons.Registry, key componenticons.Key, p core.IconProps) core.IconProps {
	if iconRegistry == nil {
		p.Src = ""
		p.ID = ""
		return p
	}

	ref, ok := iconRegistry.Resolve(key)
	if !ok {
		p.Src = ""
		p.ID = ""
		return p
	}

	return PropsForAssets(assetRegistry, ref.Sprite, ref.ID, p)
}

// PropsForRegistry returns IconProps for a semantic component icon key using
// assets.Default plus the provided icon registry.
func PropsForRegistry(r *componenticons.Registry, key componenticons.Key, p core.IconProps) core.IconProps {
	return PropsForRegistries(componentassets.Default, r, key, p)
}

// PropsFor returns IconProps for a semantic component icon key on the default registries.
func PropsFor(key componenticons.Key, p core.IconProps) core.IconProps {
	return PropsForRegistries(componentassets.Default, componenticons.Default, key, p)
}
