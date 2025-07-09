package envzilla

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/bsagat/envzilla"
)

func TestLoader(t *testing.T) {
	// Valid file
	err := envzilla.Loader("mocks/valid.env", "mocks/valid.env")
	if err != nil {
		t.Errorf("Loader test error, valid file check: %s", err.Error())
	}

	// Not existing file
	err = envzilla.Loader("mocks/FileThatWillNeverBeenHere.env")
	if err == nil {
		t.Errorf("Loader test error, not exist file check, expected:  file not found, got: nil")
	}
}

func TestDefaultParsing(t *testing.T) {
	slog.Info("Default case parsing test...")

	ParseAndCompare(t, []byte(`foo=bar`), "foo", "bar", "Testing default case")
	ParseAndCompare(t, []byte(`w i t h=spaces`), "w i t h", "spaces", "Testing with key space padding")
	ParseAndCompare(t, []byte(`with=s p a c e s`), "with", "s p a c e s", "Testing with value space padding")
	ParseAndCompare(t, []byte(`  with=space`), "with", "space", "Testing with key space padding (left)")
	ParseAndCompare(t, []byte(`with  =space`), "with", "space", "Testing with key space padding (right)")
	ParseAndCompare(t, []byte(`with=  space`), "with", "space", "Testing with value space padding (left)")
	ParseAndCompare(t, []byte(`with=space  `), "with", "space", "Testing with value space padding (right)")
	ParseAndCompare(t, []byte(`INVALID LINE`), "INVALID LINE", "", "Testing Invalid line")
	ParseAndCompare(t, []byte(`foo=`), "foo", "", "Testing empty value")
	ParseAndCompare(t, []byte(`=value`), "", "", "Testing empty key")
}

func TestCommentParsing(t *testing.T) {
	slog.Info("Comment parsing test...")

	ParseAndCompare(t, []byte(`# new=var`), "new", "", "Testing comment line ignoring")
	ParseAndCompare(t, []byte(`# new=#var`), "new", "", "Testing double comment line ignoring")
	ParseAndCompare(t, []byte(`foo=bar # Comment line`), "foo", "bar", "Testing comment line ignoring with environment")
	ParseAndCompare(t, []byte(`kairat=#nurtas`), "kairat", "", "Testing comment line ignoring before value")
	ParseAndCompare(t, []byte(`nurtas#=kairat`), "nurtas#", "", "Testing comment line ignoring after key")
}

func TestQuotesParser(t *testing.T) {
	slog.Info("Double quotes parsing test...")

	ParseAndCompare(t, []byte(`foo="bar"`), "foo", "bar", "Testing with two double quotes")
	ParseAndCompare(t, []byte(`gg="bro`), "gg", "\"bro", "Testing with one double quotes")
	ParseAndCompare(t, []byte(`empty=""`), "empty", "", "Testing with empty two double quotes")
	ParseAndCompare(t, []byte(`empty="`), "empty", "\"", "Testing with empty one double quotes")
}

type parseTestConfig struct {
	StringField string  `env:"TEST_STRING" default:"default_value"`
	IntField    int     `env:"TEST_INT" default:"42"`
	FloatField  float64 `env:"TEST_FLOAT" default:"3.14"`
	BoolField   bool    `env:"TEST_BOOL" default:"true"`
}

func TestParse_WithEnvValues(t *testing.T) {
	// Arrange
	os.Setenv("TEST_STRING", "from_env")
	os.Setenv("TEST_INT", "100")
	os.Setenv("TEST_FLOAT", "2.718")
	os.Setenv("TEST_BOOL", "false")
	defer func() {
		os.Unsetenv("TEST_STRING")
		os.Unsetenv("TEST_INT")
		os.Unsetenv("TEST_FLOAT")
		os.Unsetenv("TEST_BOOL")
	}()

	cfg := parseTestConfig{}

	// Act
	err := envzilla.Parse(&cfg)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	// Assert
	if cfg.StringField != "from_env" {
		t.Errorf("expected StringField=from_env, got %s", cfg.StringField)
	}
	if cfg.IntField != 100 {
		t.Errorf("expected IntField=100, got %d", cfg.IntField)
	}
	if cfg.FloatField != 2.718 {
		t.Errorf("expected FloatField=2.718, got %f", cfg.FloatField)
	}
	if cfg.BoolField != false {
		t.Errorf("expected BoolField=false, got %v", cfg.BoolField)
	}
}

func TestParse_WithDefaults(t *testing.T) {
	// Arrange
	cfg := parseTestConfig{}

	// Act
	err := envzilla.Parse(&cfg)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	// Assert
	if cfg.StringField != "default_value" {
		t.Errorf("expected StringField=default_value, got %s", cfg.StringField)
	}
	if cfg.IntField != 42 {
		t.Errorf("expected IntField=42, got %d", cfg.IntField)
	}
	if cfg.FloatField != 3.14 {
		t.Errorf("expected FloatField=3.14, got %f", cfg.FloatField)
	}
	if cfg.BoolField != true {
		t.Errorf("expected BoolField=true, got %v", cfg.BoolField)
	}
}

func TestParse_ErrorOnNonPointer(t *testing.T) {
	cfg := parseTestConfig{}
	err := envzilla.Parse(cfg) // Passing struct instead of *struct
	if err == nil {
		t.Errorf("expected error when passing non-pointer to Parse, got nil")
	}
}

func TestParse_ErrorOnNonStructPointer(t *testing.T) {
	var i int
	err := envzilla.Parse(&i) // Passing pointer to non-struct
	if err == nil {
		t.Errorf("expected error when passing non-struct pointer to Parse, got nil")
	}
}

func TestParse_MissingEnvTag(t *testing.T) {
	type badConfig struct {
		Field string `json:"no_env_tag"`
	}
	cfg := badConfig{}
	err := envzilla.Parse(&cfg)
	if err == nil {
		t.Errorf("expected error due to missing env tag, got nil")
	}
}

func TestParse_ErrorOnInvalidEnvValue(t *testing.T) {
	type badConfig struct {
		IntField int `env:"TEST_INVALID_INT"`
	}

	os.Setenv("TEST_INVALID_INT", "not_an_int")
	defer os.Unsetenv("TEST_INVALID_INT")

	cfg := badConfig{}
	err := envzilla.Parse(&cfg)
	if err == nil {
		t.Errorf("expected error due to invalid int conversion, got nil")
	}
}

func ParseAndCompare(t *testing.T, input []byte, inputKey string, expected string, description string) {
	fmt.Println(description + "...")

	m := envzilla.BytesParser(input)
	if m[inputKey] != expected {
		t.Errorf("Expected %s=%s, got: %s", inputKey, expected, m[inputKey])
	}
}
