package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
	"github.com/google/uuid"
)

type CreateCondominiumHandler struct {
	CreateCondominium usecases.CreateCondominiumUC
}

type CreateCondominiumRequest struct {
	Name     string `json:"name" validate:"required"`
	Cnpj     string `json:"cnpj" validate:"required"`
	Address  string `json:"address" validate:"required"`
	PlanType string `json:"planType" validate:"required,oneof=basic pro enterprise"`
}

type CreateCondominiumResponse struct {
	Message       string    `json:"message"`
	CondominiumID uuid.UUID `json:"condominiumId"`
}

// Handle executes the condominium creation
// @Summary 		Create Condominium
// @Description	Creates a new condominium. Only users with the **admin** role are allowed to perform this action.
// @Security		BearerAuth
// @Tags				Condominiums
// @Accept			json
// @Produce			json
// @Param				request body controllers.CreateCondominiumRequest true "Condominium creation payload"
// @Success			201 {object} controllers.CreateCondominiumResponse "Condominium successfully created"
// @Failure 		400	{object} common.ErrResponse "Invalid JSON payload"
// @Failure			401 {object} common.ErrResponse "User not authenticated"
// @Failure     403 {object} common.ErrResponse "User does not have permission to create a condominium"
// @Failure			422 {object} common.ValidationErrResponse "Validation failed"
// @Failure			500 {object} common.ErrResponse	"Internal server error"
// @Router			/condominiums [post]
func (h *CreateCondominiumHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[CreateCondominiumRequest](r)
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

	userId, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not found",
		})
		return
	}

	useCasePayload := usecases.CreateCondominiumReq{
		Name:     data.Name,
		Cnpj:     data.Cnpj,
		Address:  data.Address,
		PlanType: data.PlanType,
		UserID:   userId,
	}

	id, err := h.CreateCondominium.Exec(r.Context(), useCasePayload)
	if err != nil {
		slog.Error("Error while creating condominium", "error", err)

		if errors.Is(err, usecases.ErrNoPermission) {
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: err.Error(),
			})
			return
		}

		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
			Message: "An unexpected error occurred while creating condominium",
		})
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusCreated, CreateCondominiumResponse{
		Message:       "Condominium successfully created",
		CondominiumID: id,
	})
}
