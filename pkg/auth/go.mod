module github.com/go-sum/auth

go 1.26.0

// replace directives enable standalone development and testing without go.work (GOWORK=off).
// In the go.work workspace these are overridden automatically by the workspace use directives.
replace (
	github.com/go-sum/componentry => ../componentry
	github.com/go-sum/send => ../send
	github.com/go-sum/server => ../server
)

require (
	github.com/go-playground/validator/v10 v10.30.1
	github.com/go-sum/componentry v0.0.0
	github.com/go-sum/send v0.0.0
	github.com/go-sum/server v0.0.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/sessions v1.4.0
	github.com/labstack/echo/v5 v5.0.4
	maragu.dev/gomponents v1.2.0
)

require (
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/time v0.14.0 // indirect
)
