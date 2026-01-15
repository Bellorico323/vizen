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

type CreatePackageHandler struct {
	CreatePackage usecases.CreatePackageUC
}

type CreatePackageRequest struct {
	CondominiumID uuid.UUID `json:"condominiumId" validate:"required"`
	ApartmentID   uuid.UUID `json:"apartmentId" validate:"required"`
	RecipientName *string   `json:"recipientName"`
	PhotoUrl      *string   `json:"photoUrl"`
}

type CreatePackageResponse struct {
	Message   string    `json:"message"`
	PackageID uuid.UUID `json:"packageId"`
}

// Handle creates a new package entry
// @Summary      Register new package
// @Description  Registers a new package arrival at the concierge. Triggers a push notification to the apartment residents.
// @Tags         Packages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body controllers.CreatePackageRequest true "Package Creation Data"
// @Success      201  {object}  controllers.CreatePackageResponse
// @Failure      400  {object}  common.ErrResponse             "Invalid JSON payload"
// @Failure      401  {object}  common.ErrResponse             "User not authenticated"
// @Failure      404  {object}  common.ErrResponse             "Condominium or Apartment not found"
// @Failure      422  {object}  common.ValidationErrResponse   "Validation failed"
// @Failure      500  {object}  common.ErrResponse             "Internal server error"
// @Router       /packages [post]
func (h *CreatePackageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[CreatePackageRequest](r)
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

	userId, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not found",
		})
		return
	}

	payload := usecases.CreatePackageReq{
		CondominiumID: data.CondominiumID,
		ApartmentID:   data.ApartmentID,
		ReceivedBy:    userId,
		RecipientName: data.RecipientName,
		PhotoUrl:      data.PhotoUrl,
	}

	id, err := h.CreatePackage.Exec(r.Context(), payload)
	if err != nil {
		slog.Error("Error while creating package",
			"error", err,
			"condo_id", data.CondominiumID,
			"user_id", userId,
		)

		switch {
		case errors.Is(err, usecases.ErrCondominiumNotFound):
		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "User does not have permission to create packages in this condominium",
			})

		case errors.Is(err, usecases.ErrApartmentNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: "Apartment not found",
			})

		case errors.Is(err, usecases.ErrApartmentIsNotFromCondominium):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: "The selected apartment does not belong to this condominium",
			})

		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "An unexpected error occured while creating package",
			})
		}
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusCreated, CreatePackageResponse{
		Message:   "Package successfully created",
		PackageID: id,
	})
}
