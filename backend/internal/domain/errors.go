package domain

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors the outer layers translate into status codes. Callers detect
// them with errors.Is, so an intermediate layer is free to wrap one with extra
// context without breaking the check.
var (
	// ErrUserNotFound is returned by a repository when no account carries the
	// requested identifier or username.
	ErrUserNotFound = errors.New("user not found")

	// ErrUsernameTaken is returned when a registration collides with an
	// account that already exists.
	ErrUsernameTaken = errors.New("username already taken")

	// ErrInvalidCredentials covers both an unknown username and a wrong
	// password. The two are deliberately indistinguishable to the caller:
	// telling them apart would let anyone probe which accounts exist.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Field identifies the part of an entity a rule applies to. The values are
// storage and transport agnostic; each outer layer maps them to its own
// naming (JSON keys, column names) so those concerns stay out of the domain.
type Field string

const (
	FieldUsername        Field = "username"
	FieldPassword        Field = "password"
	FieldConfirmPassword Field = "confirm_password"
)

// ViolationCode names a broken rule without prescribing how it is worded.
// Turning a code into a sentence is a presentation concern, so the domain
// deliberately carries no user facing text.
type ViolationCode string

const (
	ViolationRequired      ViolationCode = "required"
	ViolationMinLength     ViolationCode = "min_length"
	ViolationMaxLength     ViolationCode = "max_length"
	ViolationInvalidFormat ViolationCode = "invalid_format"
	ViolationMismatch      ViolationCode = "mismatch"
)

// Violation is a single rule broken by a single field.
type Violation struct {
	Field Field
	Code  ViolationCode
	// Limit carries the allowed length when Code is ViolationMinLength or
	// ViolationMaxLength.
	Limit int
}

// ValidationError collects every rule broken by one operation so a caller can
// report them all at once instead of one round trip per mistake.
type ValidationError struct {
	Violations []Violation
}

func (e *ValidationError) Error() string {
	parts := make([]string, len(e.Violations))
	for index, violation := range e.Violations {
		parts[index] = fmt.Sprintf("%s=%s", violation.Field, violation.Code)
	}
	return "invalid credentials input: " + strings.Join(parts, ", ")
}

// violations accumulates rule breaches while a value is being built.
type violations []Violation

func (v *violations) add(field Field, code ViolationCode) {
	*v = append(*v, Violation{Field: field, Code: code})
}

func (v *violations) addLength(field Field, code ViolationCode, limit int) {
	*v = append(*v, Violation{Field: field, Code: code, Limit: limit})
}
