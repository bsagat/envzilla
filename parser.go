package envzilla

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
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

	return BytesParser(bytes), nil
}

func BytesParser(raw []byte) map[string]string {
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
	return env
}

var (
	envTag     = "env"
	defaultTag = "default"
)

// Sets struct fields from environment variables using reflection.
func Parse(cfg interface{}) error {
	ptrVal := reflect.ValueOf(cfg)
	if ptrVal.Kind() != reflect.Ptr {
		return ErrIsNotStructPointer
	}

	structVal := ptrVal.Elem()
	if structVal.Kind() != reflect.Struct {
		return ErrIsNotStructPointer
	}

	return processStruct(structVal)
}

// It reads the `env` tag for the environment variable key and the `default` tag for fallback values.
func processStruct(structVal reflect.Value) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		envKey, hasKey := fieldType.Tag.Lookup(envTag)
		defVal, hasDefault := fieldType.Tag.Lookup(defaultTag)

		var valueToSet string

		if hasKey && envKey != "" {
			envVal := os.Getenv(envKey)
			if len(envVal) == 0 && !hasDefault {
				return fmt.Errorf("%s field tag provided, but not found", envKey)
			}
			valueToSet = envVal
		} else {
			return fmt.Errorf("%w for field %s", ErrMissingEnvTag, fieldType.Name)
		}

		if hasDefault && defVal != "" && valueToSet == "" {
			valueToSet = defVal
		}

		if err := setField(field, valueToSet); err != nil {
			return fmt.Errorf("cannot set field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// Supports fields of type string, int, float, and bool.
func setField(field reflect.Value, value string) error {
	if !field.CanSet() {
		return errors.New("field cannot be set")
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %s to int: %w", value, err)
		}
		field.SetInt(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %s to float: %w", value, err)
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("cannot convert %s to bool: %w", value, err)
		}
		field.SetBool(b)
	default:
		return fmt.Errorf("unsupported kind: %s", field.Kind())
	}

	return nil
}
