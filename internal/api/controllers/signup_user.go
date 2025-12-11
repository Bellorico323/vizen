package controllers

import (
	"errors"
	"net/http"

	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
)

type SignupHandler struct {
	SignUpUseCase *usecases.SignupUserWithCredentials
}

type SignupRequest struct {
	Name      string  `json:"name" validate:"required,min=3,max=100"`
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=6"`
	Role      string  `json:"role" validate:"omitempty,oneof=resident staff admin"`
	AvatarURL *string `json:"avatar_url" validate:"omitempty,url"`
}

func (h *SignupHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[SignupRequest](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "Invalid json",
		})
		return
	}

	if validationErrors := validator.ValidateStruct(data); len(validationErrors) > 0 {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, map[string]any{
			"message": "Validation failed",
			"errors":  validationErrors,
		})
		return
	}

	role := data.Role
	if role == "" {
		role = "resident"
	}

	useCasePayload := usecases.SignupUserwithCredentialsReq{
		Name:       data.Name,
		Avatar_url: data.AvatarURL,
		Role:       data.Role,
		Email:      data.Email,
		Password:   data.Password,
	}

	id, err := h.SignUpUseCase.Exec(r.Context(), useCasePayload)
	if err != nil {
		if errors.Is(err, usecases.ErrDuplicatedEmail) {
			jsonutils.EncodeJson(w, r, http.StatusConflict, map[string]any{
				"message": err,
			})
			return
		}

		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"message": "An error occured while creating user",
		})
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"message": "User successfully created",
		"userId":  id,
	})
}
