package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type GetPackageHandler struct {
	GetPackage usecases.GetPackageUC
}

type GetPackageResponse struct {
	ID              uuid.UUID  `json:"id"`
	CondominiumID   uuid.UUID  `json:"condominiumId"`
	ApartmentID     uuid.UUID  `json:"apartmentId"`
	Block           *string    `json:"block"`
	ApartmentNumber string     `json:"apartmentNumber"`
	ReceivedBy      uuid.UUID  `json:"receivedBy"`
	ReceivedByName  string     `json:"receivedByName"`
	ReceivedAt      time.Time  `json:"receivedAt"`
	RecipientName   *string    `json:"recipientName"`
	PhotoUrl        *string    `json:"photoUrl"`
	Status          string     `json:"status"`
	WithdrawnAt     *time.Time `json:"withdrawnAt"`
	WithdrawnBy     *uuid.UUID `json:"withdrawnBy"`
}

// Handle retrieves a specific package by ID
// @Summary      Get Package Details
// @Description  Get detailed information about a package. Admins/Syndics can view any package; Residents can only view packages for their apartments.
// @Tags         Packages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Package UUID"
// @Success      200  {object}  controllers.GetPackageResponse
// @Failure      400  {object}  common.ErrResponse "Invalid ID format"
// @Failure      401  {object}  common.ErrResponse "User not authenticated"
// @Failure      403  {object}  common.ErrResponse "Permission denied"
// @Failure      404  {object}  common.ErrResponse "Package not found"
// @Failure      500  {object}  common.ErrResponse "Internal server error"
// @Router       /packages/{id} [get]
func (h *GetPackageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	packageID, err := uuid.Parse(idStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid package ID format",
		})
		return
	}

	pkg, err := h.GetPackage.Exec(r.Context(), usecases.GetPackageReq{
		UserID:    userID,
		PackageID: packageID,
	})

	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrPackageNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: "Package not found",
			})
		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "You do not have permission to view this package",
			})
		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "Failed to retrieve package details",
			})
		}
		return
	}

	response := GetPackageResponse{
		ID:              pkg.ID,
		CondominiumID:   pkg.CondominiumID,
		ApartmentID:     pkg.ApartmentID,
		Block:           pkg.Block,
		ApartmentNumber: pkg.ApartmentNumber,
		ReceivedBy:      pkg.ReceivedBy,
		ReceivedByName:  pkg.ReceivedByName,
		ReceivedAt:      pkg.ReceivedAt,
		RecipientName:   pkg.RecipientName,
		PhotoUrl:        pkg.PhotoUrl,
		Status:          pkg.Status,
		WithdrawnAt:     nil,
		WithdrawnBy:     pkg.WithdrawnBy,
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, response)
}
