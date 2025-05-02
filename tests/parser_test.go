package envzilla

import (
	"envzilla"
	"testing"
)

// Тесты

// Что будет если ...
// Закину валидный файл
// Закину два валидных файла сразу
// Закинуть рандомный файл которого нету
// Закину тестовый файл с:
// 		- с комментариями
//		- с двойными кавычками
//

func TestLoader(t *testing.T) {
	// Валидный файл
	err := envzilla.Loader("mocks/valid.env", "mocks/valid.env")
	if err != nil {
		t.Errorf("Loader test error, valid file check: %s", err.Error())
	}

	// Файл которого нету
	err = envzilla.Loader("mocks/FileThatWillNeverBeenHere.env")
	if err == nil {
		t.Errorf("Loader test error, not exist file check, expected:  file not found, got: nil")
	}
}

func TestBytesParser(t *testing.T) {
	input := []byte(`
	# Full comment line
	foo=bar # Comment line
	kairat=#nurtas
	nurtas#=kairat
	INVALID LINE 
	
	# double #comment
	# Full comment line at the end
	`)

	m, err := envzilla.BytesParser(input)
	if err == nil {
		if m["foo"] != "bar" {
			t.Errorf("Expected foo=bar, got: %s", m["foo"])
		}
		if m["kairat"] != "" {
			t.Errorf("Expected kairat= , got: %s", m["kairat"])
		}
		if m["nurtas#"] != "" {
			t.Errorf("Expected nurtas#= , got: %s", m["nurtas#"])
		}
		if _, exists := m["INVALID"]; exists {
			t.Errorf("Expected INVALID LINE to be ignored, but got: %s", m["INVALID"])
		}
	} else {
		t.Errorf("BytesParser returned error: %s", err.Error())
	}
}
