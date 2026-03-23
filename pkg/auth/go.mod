module github.com/go-sum/auth

go 1.26.0

// replace directives enable standalone development and testing without go.work (GOWORK=off).
// In the go.work workspace these are overridden automatically by the workspace use directives.
replace (
	github.com/go-sum/componentry => ../componentry
	github.com/go-sum/server => ../server
)

require (
	github.com/go-playground/validator/v10 v10.30.1
	github.com/go-sum/componentry v0.0.0
	github.com/go-sum/server v0.0.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/sessions v1.4.0
	github.com/jackc/pgx/v5 v5.8.0
	github.com/labstack/echo/v5 v5.0.4
	golang.org/x/crypto v0.49.0
	maragu.dev/gomponents v1.2.0
)

require (
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/parsers/yaml v1.1.0 // indirect
	github.com/knadh/koanf/providers/env v1.1.0 // indirect
	github.com/knadh/koanf/providers/file v1.2.1 // indirect
	github.com/knadh/koanf/v2 v2.3.3 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	go.yaml.in/yaml/v3 v3.0.3 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/time v0.14.0 // indirect
)
