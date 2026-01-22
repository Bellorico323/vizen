package controllers

import (
	"errors"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type CreateCommonAreaHandler struct {
	CreateCommonArea usecases.CreateCommonAreaUC
}

type CreateCommonAreaRequest struct {
	CondominiumID    uuid.UUID `json:"condominiumId" validate:"required"`
	Name             string    `json:"name" validate:"required,min=3"`
	Capacity         *int32    `json:"capacity" validate:"omitempty,min=1"`
	RequiresApproval bool      `json:"requiredApproval"`
}

type CreateCommonAreaResponse struct {
	Message    string             `json:"message"`
	CommonArea pgstore.CommonArea `json:"commonArea"`
}

// Create handles the creation of a new common area
// @Summary      Create Common Area
// @Description  Creates a new bookable area (e.g., Gym, BBQ). Only Admin/Syndic can perform this action.
// @Tags         Common Areas
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body      controllers.CreateCommonAreaRequest true "Common Area Data"
// @Success      201     {object}  controllers.CreateCommonAreaResponse
// @Failure      400     {object}  common.ErrResponse
// @Failure      403     {object}  common.ErrResponse
// @Failure      409     {object}  common.ErrResponse "Conflict (Name already exists)"
// @Failure      500     {object}  common.ErrResponse
// @Router       /common-areas [post]
func (h *CreateCommonAreaHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[CreateCommonAreaRequest](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid JSON payload",
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

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "Unauthorized",
		})
		return
	}

	payload := usecases.CreateCommonAreaReq{
		CondominiumID:    data.CondominiumID,
		UserID:           userID,
		Name:             data.Name,
		Capacity:         data.Capacity,
		RequiresApproval: data.RequiresApproval,
	}

	area, err := h.CreateCommonArea.Exec(r.Context(), payload)
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{Message: "Only admins or syndics can create common areas"})

		case errors.Is(err, usecases.ErrInvalidAreaName):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: err.Error()})

		default:
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation
				jsonutils.EncodeJson(w, r, http.StatusConflict, common.ErrResponse{Message: "An area with this name already exists in this condominium"})
			}

			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{Message: "Failed to create common area"})
		}

		return
	}

	jsonutils.EncodeJson(w, r, http.StatusCreated, CreateCommonAreaResponse{
		Message:    "Common area created successfully",
		CommonArea: area,
	})
}
