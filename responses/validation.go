package responses

import (
	"dcfs/constants"
	"errors"
	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	// Error code: 422
	Success          bool              `json:"success"`
	Message          string            `json:"message"`
	Code             string            `json:"code"`
	ValidationErrors []ValidationError `json:"validationErrors"`
}

func getFieldValidationErrorMessage(field validator.FieldError) string {
	switch field.Tag() {
	case "required":
		return "Field " + field.Field() + " is required."
	case "email":
		return "Invalid email address."
	case "lte":
		return "Field " + field.Field() + " must be at least " + field.Param() + " characters."
	case "gte":
		return "Field " + field.Field() + " must be at most " + field.Param() + " characters."
	case "eqfield":
		return "Field " + field.Field() + " must be equal to " + field.Param() + "."
	default:
		return "Invalid field " + field.Field() + "."
	}
}

func NewValidationErrorResponse(err error) *ValidationErrorResponse {
	var r *ValidationErrorResponse = new(ValidationErrorResponse)
	var valErr validator.ValidationErrors

	// Create response header
	r.Success = false
	r.Code = constants.VAL_VALIDATOR_ERROR
	r.Message = "Validation error"

	// Create validation errors
	if errors.As(err, &valErr) {
		r.ValidationErrors = make([]ValidationError, len(valErr))
		for i, field := range valErr {
			r.ValidationErrors[i].Field = field.Field()
			r.ValidationErrors[i].Message = getFieldValidationErrorMessage(field)
		}
	}

	return r
}

func NewValidationErrorResponseSingle(code string, field string, message string) *ValidationErrorResponse {
	var r *ValidationErrorResponse = new(ValidationErrorResponse)

	// Create response header
	r.Success = false
	r.Code = code
	r.Message = "Validation error"

	// Create validation errors array
	r.ValidationErrors = make([]ValidationError, 1)
	r.ValidationErrors[0].Field = field
	r.ValidationErrors[0].Message = message

	return r
}

func NewValidationErrorResponseEmpty(code string, message string) *ValidationErrorResponse {
	var r *ValidationErrorResponse = new(ValidationErrorResponse)

	// Create response header
	r.Success = false
	r.Code = code
	r.Message = message

	// Create empty validation errors array
	r.ValidationErrors = make([]ValidationError, 0)

	return r
}
