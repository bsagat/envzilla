package envzilla

import "errors"

var (
	ErrIsNotStructPointer = errors.New("value must be a pointer to struct")
	ErrMissingEnvTag      = errors.New("missing `env` tag")
)
