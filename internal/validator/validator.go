package validator

import "github.com/go-playground/validator/v10"

var validate = validator.New()

type ErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func ValidateStruct(s any) []ErrorResponse {
	var errors []ErrorResponse

	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.Field = err.Field()

			switch err.Tag() {
			case "required":
				element.Message = "Este campo é obrigatório"
			case "email":
				element.Message = "Email inválido"
			case "min":
				element.Message = "O tamanho deve ser maior que " + err.Param()
			case "max":
				element.Message = "O tamanho deve ser menor que " + err.Param()
			case "oneof":
				element.Message = "Deve ser um destes valores: " + err.Param()
			case "url":
				element.Message = "Deve ser uma URL válida"
			default:
				element.Message = "Inválido"
			}
			errors = append(errors, element)
		}
	}

	return errors
}
