package apicalls

import "errors"

type ErrorWrapper struct {
	Error error
	Code  string
}

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
