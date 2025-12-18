package common

import "github.com/Bellorico323/vizen/internal/validator"

// ErrResponse defines the standard error structure
type ErrResponse struct {
	Message string `json:"message"`
}

// ValidationErrResponse defines the validation error structure
type ValidationErrResponse struct {
	Message string                    `json:"message"`
	Errors  []validator.ErrorResponse `json:"errors"`
}
