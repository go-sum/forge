module github.com/go-sum/forge

go 1.26.0

replace (
	github.com/go-sum/auth => ./pkg/auth
	github.com/go-sum/componentry => ./pkg/componentry
	github.com/go-sum/security => ./pkg/security
	github.com/go-sum/send => ./pkg/send
	github.com/go-sum/server => ./pkg/server
	github.com/go-sum/site => ./pkg/site
)

require (
	github.com/evanw/esbuild v0.27.4
	github.com/go-playground/validator/v10 v10.30.1
	github.com/go-sum/auth v0.0.0
	github.com/go-sum/componentry v0.0.0
	github.com/go-sum/security v0.0.0
	github.com/go-sum/send v0.0.0
	github.com/go-sum/server v0.0.0
	github.com/go-sum/site v0.0.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.9.1
	github.com/labstack/echo/v5 v5.0.4
	golang.org/x/sync v0.20.0
	maragu.dev/gomponents v1.2.0
)

require (
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/gorilla/sessions v1.4.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/parsers/yaml v1.1.0 // indirect
	github.com/knadh/koanf/v2 v2.3.4 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/time v0.15.0 // indirect
)
