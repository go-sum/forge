// Package config provides a generic, reusable configuration loader.
// It supports Go struct literal defaults, env:"VAR" struct tag overrides,
// environment-specific overlay functions, and struct validation via
// go-playground/validator tags.
package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Load creates a new config of type T by:
//  1. Calling defaults() to get a fully populated struct.
//     Env var resolution belongs in the defaults function via ExpandEnv.
//  2. Running each override function in order (e.g. environment overlays).
//  3. Validating the result with go-playground/validator.
func Load[T any](defaults func() T, overrides ...func(*T)) (*T, error) {
	cfg := defaults()
	for _, o := range overrides {
		o(&cfg)
	}
	if err := Validate(&cfg); err != nil {
		return nil, fmt.Errorf("config: validation: %w", err)
	}
	return &cfg, nil
}

// ApplyEnv walks target (must be a pointer to a struct) recursively.
// For each exported field tagged with env:"VAR_NAME", if os.Getenv(VAR_NAME)
// is non-empty, the field is set from the env value.
//
// Supported field kinds: string, int*, bool, float*, []string (comma-separated).
// Fields without an env tag that are themselves structs are recursed into.
func ApplyEnv(target any) {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return
	}
	applyEnvStruct(v.Elem())
}

func applyEnvStruct(v reflect.Value) {
	t := v.Type()
	for i := range t.NumField() {
		field := v.Field(i)
		ft := t.Field(i)

		if !ft.IsExported() {
			continue
		}

		envVar := ft.Tag.Get("env")
		if envVar != "" {
			if val := os.Getenv(envVar); val != "" {
				setField(field, val)
			}
			continue
		}

		// Recurse into nested structs.
		switch field.Kind() {
		case reflect.Struct:
			applyEnvStruct(field)
		case reflect.Ptr:
			if !field.IsNil() && field.Elem().Kind() == reflect.Struct {
				applyEnvStruct(field.Elem())
			}
		}
	}
}

func setField(field reflect.Value, val string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			field.SetInt(n)
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(val); err == nil {
			field.SetBool(b)
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			field.SetFloat(f)
		}
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(val, ",")
			s := reflect.MakeSlice(field.Type(), len(parts), len(parts))
			for i, p := range parts {
				s.Index(i).SetString(strings.TrimSpace(p))
			}
			field.Set(s)
		}
	}
}

// Registrar is optionally implemented by config types that need cross-field
// validation rules beyond what struct tags can express. If the target passed
// to Validate implements Registrar, RegisterValidationRules is called with the validator
// instance before struct validation runs.
type Registrar interface {
	RegisterValidationRules(v *validator.Validate)
}

// Validate runs go-playground/validator on target.
// If target implements Registrar, its RegisterValidationRules method is called first
// to register any cross-field or struct-level validation rules.
func Validate(target any) error {
	v := validator.New()
	if r, ok := target.(Registrar); ok {
		r.RegisterValidationRules(v)
	}
	return v.Struct(target)
}

// ExpandEnv replaces ${VAR} and ${VAR:-default} patterns in s using os.Getenv.
// Unset or empty variables use the default when the :- form is present;
// otherwise they expand to empty string.
func ExpandEnv(s string) string {
	var buf strings.Builder
	for {
		start := strings.Index(s, "${")
		if start == -1 {
			buf.WriteString(s)
			return buf.String()
		}
		buf.WriteString(s[:start])
		s = s[start+2:]
		end := strings.Index(s, "}")
		if end == -1 {
			// Unclosed placeholder — write literal and stop.
			buf.WriteString("${")
			buf.WriteString(s)
			return buf.String()
		}
		expr := s[:end]
		s = s[end+1:]
		if before, after, ok := strings.Cut(expr, ":-"); ok {
			key, def := before, after
			if v := os.Getenv(key); v != "" {
				buf.WriteString(v)
			} else {
				buf.WriteString(def)
			}
		} else {
			buf.WriteString(os.Getenv(expr))
		}
	}
}
