package envzilla

import (
	"bytes"
	"io"
	"os"
)

var (
	doublequotes byte = '"'
	newLine      byte = '\n'
	CRLF         byte = '\r'
	hashTag      byte = '#'
	equal        byte = '='
)

func Loader(filepaths ...string) error {
	if len(filepaths) == 0 {
		filepaths = []string{".env"}
	}

	for i := 0; i < len(filepaths); i++ {
		m, err := load(filepaths[i])
		if err != nil {
			return err
		}

		if err := setVariables(m); err != nil {
			return err
		}
	}

	return nil
}

func setVariables(m map[string]string) error {
	for key, value := range m {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}

func load(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return BytesParser(bytes)
}

func BytesParser(raw []byte) (map[string]string, error) {
	var key, value, empty []byte
	var isKeyAdded, isCommented bool

	env := make(map[string]string, 5)
	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case CRLF:
		case newLine:
			value = bytes.TrimSpace(value)
			key = bytes.TrimSpace(key)

			// Проверка на двойные скобки
			if len(value) >= 2 {
				if value[0] == doublequotes && value[len(value)-1] == doublequotes {
					if len(value) == 2 {
						value = empty
					} else {
						value = value[1 : len(value)-1]
					}
				}
			}
			if len(key) != 0 && isKeyAdded {
				env[string(key)] = string(value)
			}
			key, value = empty, empty
			isCommented, isKeyAdded = false, false
		case equal:
			if !isCommented {
				isKeyAdded = true
			}
		case hashTag:
			isCommented = true
		default:
			if isCommented {
				break
			}
			if isKeyAdded {
				value = append(value, raw[i])
			} else {
				key = append(key, raw[i])
			}
		}
	}
	if len(key) != 0 && isKeyAdded {
		value = bytes.TrimSpace(value)
		key = bytes.TrimSpace(key)

		if len(value) >= 2 {
			if value[0] == doublequotes && value[len(value)-1] == doublequotes {
				if len(value) == 2 {
					value = empty
				} else {
					value = value[1 : len(value)-1]
				}
			}
		}

		env[string(key)] = string(value)
	}
	return env, nil
}
