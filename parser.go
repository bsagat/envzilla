// Package envzilla provides a simple parser for .env files and minimal YAML support,
// along with a mechanism to load environment variables or YAML parameters
// into a struct using `env` and `default` tags.
package envzilla

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	doublequotes byte = '"' // Double quote character
	newLine      byte = '\n'
	CRLF         byte = '\r'
	hashTag      byte = '#'
	equal        byte = '='
)

// LoadAndParse loads variables from .env or YAML files and maps them into the struct v.
func LoadAndParse(v interface{}, filepaths ...string) error {
	if err := Loader(filepaths...); err != nil {
		return err
	}
	return Parse(v)
}

// Loader loads configuration from the specified files (.env or .yaml/.yml).
// If no files are specified, it defaults to ".env".
func Loader(filepaths ...string) error {
	if len(filepaths) == 0 {
		filepaths = []string{".env"}
	}

	for _, path := range filepaths {
		switch {
		case strings.HasSuffix(path, ".env"):
			m, err := loadEnv(path)
			if err != nil {
				return err
			}
			if err := setVariables(m); err != nil {
				return err
			}
		case strings.HasSuffix(path, ".yaml"), strings.HasSuffix(path, ".yml"):
			if err := loadYaml(path); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported file extension: %s", path)
		}
	}

	return nil
}

// setVariables iterates by map and sets environment values by key-value pairs
func setVariables(m map[string]string) error {
	for key, value := range m {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}

// loadEnv reads a .env file and returns a map of key-value pairs.
func loadEnv(filePath string) (map[string]string, error) {
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

// loadYaml reads a YAML file and loads variables into the environment.
// It supports nested structures using indentation (2 spaces per level).
func loadYaml(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open YAML file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentIndent := 0
	prefixStack := []string{}

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " ")

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Calculate indentation
		indent := 0
		for _, ch := range line {
			if ch != ' ' {
				break
			}
			indent++
		}

		// Update prefix stack based on indentation
		for indent < currentIndent && len(prefixStack) > 0 {
			prefixStack = prefixStack[:len(prefixStack)-1]
			currentIndent -= 2 // Standard YAML indent is 2 spaces
		}
		currentIndent = indent

		// Parse the line
		content := strings.TrimSpace(line)
		if strings.HasSuffix(content, ":") {
			// New section
			sectionName := strings.TrimSuffix(content, ":")
			prefixStack = append(prefixStack, sectionName)
			currentIndent += 2
		} else {
			// Key-value pair
			parts := strings.SplitN(content, ":", 2)
			if len(parts) != 2 {
				continue // Skip malformed lines
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Build full env var name
			fullKey := strings.Join(append(prefixStack, key), "_")
			fullKey = strings.ToUpper(fullKey)

			// Remove quotes if present
			value = strings.Trim(value, `"'`)

			// Empty values become "true" (YAML-style flags)
			if value == "" {
				value = "true"
			}

			if err := os.Setenv(fullKey, value); err != nil {
				return fmt.Errorf("could not set env var %s: %w", fullKey, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading YAML file: %w", err)
	}

	return nil
}

// BytesParser parses the content of a .env file into a map[string]string.
func BytesParser(raw []byte) map[string]string {
	var key, value, empty []byte
	var isKeyAdded, isCommented bool

	env := make(map[string]string, 5)
	for i := range raw {
		switch raw[i] {
		case CRLF:
		case newLine:
			value = bytes.TrimSpace(value)
			key = bytes.TrimSpace(key)

			if len(value) >= 2 && value[0] == doublequotes && value[len(value)-1] == doublequotes {
				if len(value) == 2 {
					value = empty
				} else {
					value = value[1 : len(value)-1]
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

		if len(value) >= 2 && value[0] == doublequotes && value[len(value)-1] == doublequotes {
			if len(value) == 2 {
				value = empty
			} else {
				value = value[1 : len(value)-1]
			}
		}

		env[string(key)] = string(value)
	}
	return env
}

// envTag and defaultTag are used in struct tags
var (
	envTag     = "env"
	defaultTag = "default"
)

// Parse reads struct tags (`env` and `default`) and fills struct fields.
func Parse(v interface{}) error {
	ptrVal := reflect.ValueOf(v)
	if ptrVal.Kind() != reflect.Ptr {
		return errors.New("provided value is not a struct pointer")
	}

	structVal := ptrVal.Elem()
	if structVal.Kind() != reflect.Struct {
		return errors.New("provided value is not a struct pointer")
	}

	return processStruct(structVal)
}

// processStruct recursively fills struct fields from environment variables or defaults.
func processStruct(structVal reflect.Value) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		if field.Kind() == reflect.Struct && fieldType.Type.Kind() == reflect.Struct {
			if err := processStruct(field); err != nil {
				return err
			}
			continue
		}

		if !field.CanSet() {
			continue
		}

		envKey, hasKey := fieldType.Tag.Lookup(envTag)
		defVal, hasDefault := fieldType.Tag.Lookup(defaultTag)

		var valueToSet string
		if hasKey && envKey != "" {
			valueToSet = os.Getenv(envKey)
		}

		if hasDefault && valueToSet == "" {
			valueToSet = defVal
		}

		if valueToSet != "" {
			if err := setField(field, valueToSet); err != nil {
				return fmt.Errorf("cannot set field %s: %w", fieldType.Name, err)
			}
		}
	}
	return nil
}

// setField supports string, int, float, bool, and time.Duration types.
func setField(field reflect.Value, value string) error {
	if !field.CanSet() {
		return errors.New("field cannot be set")
	}

	if field.Type() == reflect.TypeOf(time.Duration(0)) {
		dur, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("cannot convert %s to duration: %w", value, err)
		}
		field.Set(reflect.ValueOf(dur))
		return nil
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
