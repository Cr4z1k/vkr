package model

import "errors"

var (
	ErrBenthosValidation = errors.New("validation error")
)

type Paths struct {
	ConfigDir  string
	ConfigFile string
}
