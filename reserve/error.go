package reserve

import "fmt"

type validationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewValidationError(code, field string) validationError {
	return validationError{
		code,
		fmt.Sprintf("field %s has errors!", field),
	}
}

type AllocationError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewAllocationError(code int, message string) AllocationError {
	return AllocationError{
		code,
		message,
	}
}
