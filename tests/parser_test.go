package envzilla

import (
	"envzilla"
	"fmt"
	"log/slog"
	"testing"
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

func ParseAndCompare(t *testing.T, input []byte, inputKey string, expected string, description string) {
	fmt.Println(description + "...")

	m, err := envzilla.BytesParser(input)
	if err != nil {
		t.Errorf("BytesParser returned error: %s", err.Error())
	}
	if m[inputKey] != expected {
		t.Errorf("Expected %s=%s, got: %s", inputKey, expected, m[inputKey])
	}
}
