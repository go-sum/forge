package assets

import (
	"maps"
	"strings"
	"sync"
)

// Registry resolves named component sprite assets to public URLs.
// Applications register the sprite files they build, while component packages
// consume the registry without knowing app-specific file locations.
type Registry struct {
	mu          sync.RWMutex
	spriteFiles map[string]string
	resolvePath func(string) string
}

// NewRegistry returns an empty sprite registry with a default /public resolver.
func NewRegistry() *Registry {
	return &Registry{
		spriteFiles: make(map[string]string),
		resolvePath: func(rel string) string { return "/public/" + strings.TrimPrefix(rel, "/") },
	}
}

// RegisterSprite associates a sprite key with its relative path under public/.
func (r *Registry) RegisterSprite(key, rel string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spriteFiles[key] = rel
}

// RegisterSprites adds or replaces multiple sprite registrations.
func (r *Registry) RegisterSprites(files map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	maps.Copy(r.spriteFiles, files)
}

// SetPathFunc replaces the relative-path resolver used for all sprite URLs.
func (r *Registry) SetPathFunc(f func(string) string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if f == nil {
		r.resolvePath = func(rel string) string { return "/public/" + strings.TrimPrefix(rel, "/") }
		return
	}
	r.resolvePath = f
}

// SpritePath returns the resolved URL for the named sprite file.
// Unknown keys return an empty string so callers do not emit broken /public URLs.
func (r *Registry) SpritePath(key string) string {
	r.mu.RLock()
	rel, ok := r.spriteFiles[key]
	resolve := r.resolvePath
	r.mu.RUnlock()
	if !ok || rel == "" {
		return ""
	}
	return resolve(rel)
}

// Default is the package-level component asset registry.
var Default = NewRegistry()

// RegisterSprite adds or replaces a single sprite on Default.
func RegisterSprite(key, rel string) { Default.RegisterSprite(key, rel) }

// RegisterSprites adds or replaces multiple sprites on Default.
func RegisterSprites(files map[string]string) { Default.RegisterSprites(files) }

// SetPathFunc replaces the path resolver on Default.
func SetPathFunc(f func(string) string) { Default.SetPathFunc(f) }

// SpritePath returns the resolved URL for a sprite in Default.
func SpritePath(key string) string { return Default.SpritePath(key) }
