package config

import (
	"testing"
)

// --- ApplyEnv tests ---

type envStringStruct struct {
	Host string `env:"TEST_CFG_HOST"`
}

func TestApplyEnvSetsStringField(t *testing.T) {
	t.Setenv("TEST_CFG_HOST", "example.com")
	s := envStringStruct{Host: "default"}
	ApplyEnv(&s)
	if s.Host != "example.com" {
		t.Fatalf("Host = %q, want %q", s.Host, "example.com")
	}
}

type envIntStruct struct {
	Port int `env:"TEST_CFG_PORT"`
}

func TestApplyEnvSetsIntField(t *testing.T) {
	t.Setenv("TEST_CFG_PORT", "8080")
	s := envIntStruct{Port: 3000}
	ApplyEnv(&s)
	if s.Port != 8080 {
		t.Fatalf("Port = %d, want %d", s.Port, 8080)
	}
}

type envBoolStruct struct {
	Debug bool `env:"TEST_CFG_DEBUG"`
}

func TestApplyEnvSetsBoolField(t *testing.T) {
	t.Setenv("TEST_CFG_DEBUG", "true")
	s := envBoolStruct{Debug: false}
	ApplyEnv(&s)
	if !s.Debug {
		t.Fatal("Debug = false, want true")
	}
}

type envFloatStruct struct {
	Rate float64 `env:"TEST_CFG_RATE"`
}

func TestApplyEnvSetsFloat64Field(t *testing.T) {
	t.Setenv("TEST_CFG_RATE", "1.5")
	s := envFloatStruct{Rate: 0}
	ApplyEnv(&s)
	if s.Rate != 1.5 {
		t.Fatalf("Rate = %f, want 1.5", s.Rate)
	}
}

type envSliceStruct struct {
	Tags []string `env:"TEST_CFG_TAGS"`
}

func TestApplyEnvSetsStringSliceField(t *testing.T) {
	t.Setenv("TEST_CFG_TAGS", "a,b,c")
	s := envSliceStruct{}
	ApplyEnv(&s)
	if len(s.Tags) != 3 || s.Tags[0] != "a" || s.Tags[1] != "b" || s.Tags[2] != "c" {
		t.Fatalf("Tags = %v, want [a b c]", s.Tags)
	}
}

func TestApplyEnvSkipsUnsetEnvVars(t *testing.T) {
	s := envStringStruct{Host: "default"}
	ApplyEnv(&s)
	if s.Host != "default" {
		t.Fatalf("Host = %q, want %q (should keep default)", s.Host, "default")
	}
}

type envNestedOuter struct {
	Inner envNestedInner
}

type envNestedInner struct {
	Value string `env:"TEST_CFG_NESTED_VALUE"`
}

func TestApplyEnvRecursesNestedStructs(t *testing.T) {
	t.Setenv("TEST_CFG_NESTED_VALUE", "from-env")
	s := envNestedOuter{Inner: envNestedInner{Value: "default"}}
	ApplyEnv(&s)
	if s.Inner.Value != "from-env" {
		t.Fatalf("Inner.Value = %q, want %q", s.Inner.Value, "from-env")
	}
}

type envNoTagStruct struct {
	Name string
}

func TestApplyEnvIgnoresFieldsWithoutTag(t *testing.T) {
	s := envNoTagStruct{Name: "original"}
	ApplyEnv(&s)
	if s.Name != "original" {
		t.Fatalf("Name = %q, want %q (untagged field must not change)", s.Name, "original")
	}
}

// --- Validate tests ---

type validateOKStruct struct {
	Name string `validate:"required"`
}

type validateBadStruct struct {
	Name string `validate:"required"`
}

func TestValidatePassesValidStruct(t *testing.T) {
	s := validateOKStruct{Name: "hello"}
	if err := Validate(&s); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
}

func TestValidateRejectsInvalidStruct(t *testing.T) {
	s := validateBadStruct{Name: ""}
	if err := Validate(&s); err == nil {
		t.Fatal("Validate() error = nil, want validation error")
	}
}

// --- Load integration tests ---

// loadTestCfg uses ExpandEnv in defaults — the recommended pattern.
type loadTestCfg struct {
	Host string `validate:"required"`
	Port int
}

func defaultLoadTestCfg() loadTestCfg {
	return loadTestCfg{
		Host: ExpandEnv("${TEST_LOAD_HOST:-localhost}"),
		Port: 3000,
	}
}

func TestLoadIntegration(t *testing.T) {
	t.Setenv("TEST_LOAD_HOST", "prod.example.com")
	cfg, err := Load(defaultLoadTestCfg)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Host != "prod.example.com" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "prod.example.com")
	}
	if cfg.Port != 3000 {
		t.Fatalf("Port = %d, want 3000 (default retained)", cfg.Port)
	}
}

func TestLoadDefaultUsedWhenEnvUnset(t *testing.T) {
	cfg, err := Load(defaultLoadTestCfg)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Host != "localhost" {
		t.Fatalf("Host = %q, want localhost (ExpandEnv fallback)", cfg.Host)
	}
}

func TestLoadAppliesOverride(t *testing.T) {
	cfg, err := Load(defaultLoadTestCfg, func(c *loadTestCfg) {
		c.Host = "override.example.com"
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Host != "override.example.com" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "override.example.com")
	}
}

func TestLoadReturnsValidationError(t *testing.T) {
	_, err := Load(func() loadTestCfg {
		return loadTestCfg{Host: "", Port: 3000} // empty Host fails validate:"required"
	})
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
}

// TestLoadOverlaySeesExpandEnvValue confirms that overlays run after defaults()
// and can read or further modify values set by ExpandEnv.
func TestLoadOverlaySeesExpandEnvValue(t *testing.T) {
	t.Setenv("TEST_LOAD_HOST", "from-env.example.com")
	cfg, err := Load(defaultLoadTestCfg, func(c *loadTestCfg) {
		c.Host = c.Host + "-modified"
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Host != "from-env.example.com-modified" {
		t.Fatalf("Host = %q, want from-env.example.com-modified", cfg.Host)
	}
}

// --- ExpandEnv tests ---

func TestExpandEnvSetVar(t *testing.T) {
	t.Setenv("TEST_EXPAND_VAR", "hello")
	if got := ExpandEnv("${TEST_EXPAND_VAR}"); got != "hello" {
		t.Fatalf("ExpandEnv = %q, want %q", got, "hello")
	}
}

func TestExpandEnvUnsetVar(t *testing.T) {
	if got := ExpandEnv("${TEST_EXPAND_UNSET_XYZ_123}"); got != "" {
		t.Fatalf("ExpandEnv = %q, want empty string", got)
	}
}

func TestExpandEnvWithDefault(t *testing.T) {
	if got := ExpandEnv("${TEST_EXPAND_MISSING_XYZ_123:-fallback}"); got != "fallback" {
		t.Fatalf("ExpandEnv = %q, want %q", got, "fallback")
	}
}

func TestExpandEnvSetVarOverridesDefault(t *testing.T) {
	t.Setenv("TEST_EXPAND_WITH_DEFAULT", "actual")
	if got := ExpandEnv("${TEST_EXPAND_WITH_DEFAULT:-fallback}"); got != "actual" {
		t.Fatalf("ExpandEnv = %q, want %q", got, "actual")
	}
}

func TestExpandEnvUnclosedPlaceholder(t *testing.T) {
	if got := ExpandEnv("${UNCLOSED"); got != "${UNCLOSED" {
		t.Fatalf("ExpandEnv = %q, want %q", got, "${UNCLOSED")
	}
}

func TestExpandEnvComposed(t *testing.T) {
	t.Setenv("TEST_KV_HOST_C", "redis")
	t.Setenv("TEST_KV_PORT_C", "6380")
	if got := ExpandEnv("${TEST_KV_HOST_C:-localhost}:${TEST_KV_PORT_C:-6379}"); got != "redis:6380" {
		t.Fatalf("ExpandEnv = %q, want %q", got, "redis:6380")
	}
}

func TestExpandEnvComposedDefaultFallback(t *testing.T) {
	if got := ExpandEnv("${TEST_KV_HOST_MISS:-localhost}:${TEST_KV_PORT_MISS:-6379}"); got != "localhost:6379" {
		t.Fatalf("ExpandEnv = %q, want %q", got, "localhost:6379")
	}
}
