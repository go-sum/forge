package main

import "testing"

func TestResolveAssetBuildOptions(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    assetBuildOptions
		wantErr string
	}{
		{
			name: "default build enables every asset group",
			want: assetBuildOptions{
				ConfigPath:   ".assets.yaml",
				BuildCSS:     true,
				BuildDocs:    true,
				BuildFonts:   true,
				BuildJS:      true,
				BuildSprites: true,
			},
		},
		{
			name: "docs only disables unrelated assets and fonts",
			args: []string{"--docs-only"},
			want: assetBuildOptions{
				ConfigPath:   ".assets.yaml",
				BuildCSS:     false,
				BuildDocs:    true,
				BuildFonts:   false,
				BuildJS:      false,
				BuildSprites: false,
			},
		},
		{
			name: "css only disables unrelated assets and fonts",
			args: []string{"--css-only"},
			want: assetBuildOptions{
				ConfigPath:   ".assets.yaml",
				BuildCSS:     true,
				BuildDocs:    false,
				BuildFonts:   false,
				BuildJS:      false,
				BuildSprites: false,
			},
		},
		{
			name: "js only disables unrelated assets and fonts",
			args: []string{"--js-only"},
			want: assetBuildOptions{
				ConfigPath:   ".assets.yaml",
				BuildCSS:     false,
				BuildDocs:    false,
				BuildFonts:   false,
				BuildJS:      true,
				BuildSprites: false,
			},
		},
		{
			name: "sprites only disables unrelated assets and fonts",
			args: []string{"--sprites-only"},
			want: assetBuildOptions{
				ConfigPath:   ".assets.yaml",
				BuildCSS:     false,
				BuildDocs:    false,
				BuildFonts:   false,
				BuildJS:      false,
				BuildSprites: true,
			},
		},
		{
			name:    "only flags remain mutually exclusive",
			args:    []string{"--docs-only", "--css-only"},
			wantErr: "choose at most one of --css-only, --docs-only, --js-only, --sprites-only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, err := parseAssetBuildFlags(tt.args)
			if err != nil {
				t.Fatalf("parseAssetBuildFlags() error = %v", err)
			}

			got, err := resolveAssetBuildOptions(flags)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("resolveAssetBuildOptions() error = nil, want %q", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("resolveAssetBuildOptions() error = %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveAssetBuildOptions() error = %v", err)
			}

			if got != tt.want {
				t.Fatalf("resolveAssetBuildOptions() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
