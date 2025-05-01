package envzilla

import (
	"bufio"
	"fmt"
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

		fmt.Println(m)

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
	return BytesParser(file)
}

func BytesParser(file *os.File) (map[string]string, error) {
	var key, value, empty []byte
	var isKeyAdded, isCommented bool

	reader := bufio.NewReader(file)
	env := make(map[string]string, 5)
	for {
		b := make([]byte, 1)
		_, err := reader.Read(b)
		if err != nil {
			if err == io.EOF {
				if len(key) != 0 && isKeyAdded {
					env[string(key)] = string(value)
				}
				return env, nil
			}
			return nil, err
		}

		switch b[0] {
		case CRLF:
		case newLine:
			// Проверка на двойные скобки
			if len(value) > 2 {
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
			isKeyAdded = true
		case hashTag:
			isCommented = true
		default:
			if isCommented {
				break
			}
			if isKeyAdded {
				value = append(value, b[0])
			} else {
				key = append(key, b[0])
			}
		}
	}
}
