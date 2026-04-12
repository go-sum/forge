package main

import "testing"

func TestResolveAssetsOpts(t *testing.T) {
	const defaultConfig = ".assets.yaml"

	tests := []struct {
		name  string
		css   bool
		js    bool
		fonts bool
		want  assetBuildOptions
	}{
		{
			name: "no flags defaults to all asset types",
			want: assetBuildOptions{
				ConfigPath: defaultConfig,
				BuildCSS:   true,
				BuildJS:    true,
				BuildFonts: true,
			},
		},
		{
			name: "css only",
			css:  true,
			want: assetBuildOptions{
				ConfigPath: defaultConfig,
				BuildCSS:   true,
				BuildJS:    false,
				BuildFonts: false,
			},
		},
		{
			name: "js only",
			js:   true,
			want: assetBuildOptions{
				ConfigPath: defaultConfig,
				BuildCSS:   false,
				BuildJS:    true,
				BuildFonts: false,
			},
		},
		{
			name:  "fonts only",
			fonts: true,
			want: assetBuildOptions{
				ConfigPath: defaultConfig,
				BuildCSS:   false,
				BuildJS:    false,
				BuildFonts: true,
			},
		},
		{
			name: "css and js combined",
			css:  true,
			js:   true,
			want: assetBuildOptions{
				ConfigPath: defaultConfig,
				BuildCSS:   true,
				BuildJS:    true,
				BuildFonts: false,
			},
		},
		{
			name:  "all three explicit equals default",
			css:   true,
			js:    true,
			fonts: true,
			want: assetBuildOptions{
				ConfigPath: defaultConfig,
				BuildCSS:   true,
				BuildJS:    true,
				BuildFonts: true,
			},
		},
		{
			name: "minify flag is preserved",
			want: assetBuildOptions{
				ConfigPath: defaultConfig,
				Minify:     true,
				BuildCSS:   true,
				BuildJS:    true,
				BuildFonts: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minify := tt.want.Minify
			got := resolveAssetsOpts(defaultConfig, minify, tt.css, tt.js, tt.fonts)
			if got != tt.want {
				t.Fatalf("resolveAssetsOpts() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
