package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
	"github.com/google/uuid"
)

type SignupHandler struct {
	SignUpUseCase usecases.SignupWithCredentials
}

type SignupRequest struct {
	Name      string  `json:"name" validate:"required,min=3,max=100"`
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=6"`
	Role      string  `json:"role" validate:"omitempty,oneof=resident staff admin"`
	AvatarURL *string `json:"avatar_url" validate:"omitempty,url"`
}

type SignupResponse struct {
	Message string    `json:"message"`
	UserID  uuid.UUID `json:"userId"`
}

// Handle executes the user creation
// @Summary			Create User
// @Description Creates an user and account with credentials
// @Tags				Auth
// @Accept			json
// @Produce 		json
// @Param 			request body controllers.SignupRequest true "Signup Credentials"
// @Success			201	{object}	controllers.SignupResponse "Created user successfully"
// @Failure			400	{object}	common.ErrResponse	"Invalid JSON payload"
// @Failure			422 {object}	common.ValidationErrResponse "Validation failed"
// @Failure     409  {object}  common.ErrResponse "Email already exists"
// @Failure			500	{object}	common.ErrResponse	"Internal server error"
// @Router			/users [post]
func (h *SignupHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[SignupRequest](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid json",
		})
		return
	}

	if validationErrors := validator.ValidateStruct(data); len(validationErrors) > 0 {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, common.ValidationErrResponse{
			Message: "Validation failed",
			Errors:  validationErrors,
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
		Role:       role,
		Email:      data.Email,
		Password:   data.Password,
	}

	id, err := h.SignUpUseCase.Exec(r.Context(), useCasePayload)
	if err != nil {
		slog.Error("Error while creating user", "error", err)

		if errors.Is(err, usecases.ErrDuplicatedEmail) {
			jsonutils.EncodeJson(w, r, http.StatusConflict, common.ErrResponse{
				Message: err.Error(),
			})
			return
		}

		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
			Message: "An error occurred while creating user",
		})
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusCreated, SignupResponse{
		Message: "User successfully created",
		UserID:  id,
	})
}
