package errors

import (
	"fmt"
)

type errorString struct {
	message string
}

// Error implements the standard library error interface.
func (s *errorString) Error() string {
	return s.message
}

// New returns an error with the supplied message without cause.
func New(message string) error {
	return &errorString{
		message: message,
	}
}

// Newf returns an error without cause with the formats according to a format specifier.
func Newf(format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)

	return &errorString{
		message: message,
	}
}

// Is implements future error.Is functionality.
// An Error is equivalent if err message identical.
func (s *errorString) Is(err error) bool {
	return s.message == err.Error()
}

type withMessage struct {
	message string
	err     error
}

// Error implements the standard library error interface.
func (wm *withMessage) Error() string {
	return wm.message
}

// Unwrap implements errors.Unwrap for Error.
func (wm *withMessage) Unwrap() error {
	return wm.err
}

// Wrap returns an error annotating
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf("%s: %s", message, err)

	return &withMessage{
		// message is the full concatenate error message (top to bottom)
		message: msg,
		// err is the original error
		err: err,
	}
}

// Wrapf returns an error annotating
// at the point Wrapf is called, and the supplied message.
// If err is nil, Wrapf returns nil.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	message := fmt.Sprintf(format, args...)

	return Wrap(err, message)
}

type withError struct {
	// message is the full concatenate error message (top to bottom)
	message string
	// err is the supplied error most of the time the sentinel error.
	err error
	// cause is the original error.
	cause error
}

// Error implements the standard library error interface.
func (we *withError) Error() string {
	return we.message
}

// Unwrap implements errors.Unwrap for Error.
func (we *withError) Unwrap() error {
	return we.err
}

// Cause returns the underlying cause of error.
func (we *withError) Cause() error {
	return we.cause
}

// WrapError returns an error annotating err with cause
// at the point WrapWithError is called, and the supplied err.
//
// If err is nil, WrapError returns supplied err.
// If supplied err is nil, WrapWithError returns err.
func WrapError(err error, supplied error) error {
	if err == nil {
		return supplied
	}

	if supplied == nil {
		return err
	}

	msg := fmt.Sprintf("%s: %s", supplied, err)

	return &withError{
		message: msg,
		err:     supplied,
		cause:   err,
	}
}

// Is implements future error.Is functionality.
// An Error is equivalent if err message or any of the underlying cause message are identical.
func (we *withError) Is(target error) bool {
	if Is(we.err, target) {
		return true
	}

	cause := Cause(we)
	if cause == nil {
		return false
	}

	return Is(cause, target)
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//	type causer interface {
//	       Cause() error
//	}
//
// If the error does not implement Cause, the error nil will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	//nolint:errorlint
	cause, ok := err.(causer)
	if !ok {
		return nil
	}

	return cause.Cause()
}

// tuples is a slice of keys and values, e.g. {"key1", 1, "key2", "val2"}.
type tuples []interface{}

// Fields creates a map from key-value pairs.
func (t tuples) fields() map[string]interface{} {
	if len(t) == 0 {
		return nil
	}

	result := make(map[string]interface{}, len(t))

	var (
		label string
		ok    bool
	)

	for i, l := range t {
		if label == "" {
			label, ok = l.(string)
			if !ok || label == "" {
				result["malformedFields"] = []interface{}(t[i:])

				break
			}
		} else {
			result[label] = l
			label = ""
		}
	}

	if label != "" {
		result["malformedFields"] = []interface{}{label}
	}

	return result
}

type enrichedError struct {
	err           error
	keysAndValues tuples
}

// Error implements the standard library error interface.
func (ee *enrichedError) Error() string {
	return ee.err.Error()
}

// Unwrap implements errors.Unwrap for Error.
func (ee *enrichedError) Unwrap() error {
	return ee.err
}

// Tuples returns structured data of error in form of loosely-typed key-value pairs.
func (ee *enrichedError) Tuples() []interface{} {
	return keysAndValues(ee)
}

func keysAndValues(err error) []interface{} {
	var kv []interface{}

	//nolint:errorlint
	if ee, ok := err.(*enrichedError); ok {
		kv = append(kv, ee.keysAndValues...)
	}

	uErr := Unwrap(err)
	if uErr == nil {
		return kv
	}

	kv = append(kv, keysAndValues(uErr)...)

	cause := Cause(err)
	if cause == nil {
		return kv
	}

	kv = append(kv, keysAndValues(cause)...)

	return kv
}

// Fields returns structured data of error as a map.
func (ee *enrichedError) Fields() map[string]interface{} {
	return ee.keysAndValues.fields()
}

// Enrich takes in a basic error object and appends additional relevant fields, enriching the error message to help
// diagnose and resolve the error more effectively.
//
// If err is nil, Enrich returns nil.
// If keysAndValues is nil, Enrich returns err.
// If err is enrichedError, the keysAndValues will be appended to the existing keysAndValues.
func Enrich(err error, keysAndValues ...interface{}) error {
	if err == nil {
		return nil
	}

	// keysAndValues must be a list of key-value pairs.
	if len(keysAndValues)%2 != 0 {
		return err
	}

	return &enrichedError{
		err:           err,
		keysAndValues: keysAndValues,
	}
}

// EnrichWrapError returns an enrichedError error annotating err with cause.
// @see WrapWithError and Enrich.
func EnrichWrapError(err error, supplied error, keysAndValues ...interface{}) error {
	return Enrich(WrapError(err, supplied), keysAndValues...)
}
