package apicalls

import "errors"

type ErrorWrapper struct {
	Error error
	Code  string
}

// CreateErrorWrapper - combine multiple errors into single error
//
// params:
//   - code string: completion code
//   - errorMessages ...string: list of error messages to combined
//
// return type:
//   - *ErrorWrapper: combined error
func CreateErrorWrapper(code string, errorMessages ...string) *ErrorWrapper {
	var msg string = ""

	for idx, _m := range errorMessages {
		if idx == 0 {
			msg = _m
			continue
		}

		msg = msg + " " + _m
	}

	return &ErrorWrapper{
		Error: errors.New(msg),
		Code:  code,
	}
}
