package install

import (
	componentassets "github.com/go-sum/componentry/assets"
	componenticonset "github.com/go-sum/componentry/assets/iconset"
	componenticons "github.com/go-sum/componentry/icons"
)

// Registries holds the asset and icon registries a consumer can install into
// component rendering paths without relying on package-global defaults.
type Registries struct {
	Assets *componentassets.Registry
	Icons  *componenticons.Registry
}

// Config controls how component registries are populated.
type Config struct {
	PathFunc      func(string) string
	Catalog       componenticonset.Catalog
	IconOverrides map[componenticons.Key]componenticons.Ref
}

func catalogOrDefault(c Config) componenticonset.Catalog {
	if c.Catalog.Sprites == nil && c.Catalog.Icons == nil {
		return componenticonset.Default
	}
	return c.Catalog
}

func apply(r Registries, c Config) Registries {
	catalog := catalogOrDefault(c)
	if c.PathFunc != nil {
		r.Assets.SetPathFunc(c.PathFunc)
	}
	r.Assets.RegisterSprites(catalog.Sprites)
	r.Icons.RegisterSet(catalog.Icons)
	if len(c.IconOverrides) > 0 {
		r.Icons.RegisterSet(c.IconOverrides)
	}
	return r
}

// New returns isolated registries populated from the provided config.
func New(c Config) Registries {
	return apply(Registries{
		Assets: componentassets.NewRegistry(),
		Icons:  componenticons.NewRegistry(),
	}, c)
}

// ApplyDefault installs the config onto the package-global default registries.
func ApplyDefault(c Config) Registries {
	return apply(Registries{
		Assets: componentassets.DefaultRegistry,
		Icons:  componenticons.Default,
	}, c)
}
