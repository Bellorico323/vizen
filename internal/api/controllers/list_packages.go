package controllers

import (
	"errors"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/google/uuid"
)

type ListPackagesHandler struct {
	ListPackages usecases.ListPackagesUC
}

// Handle lists packages based on filters
// @Summary      List Packages
// @Description  Retrieve a list of packages. Admins can view all packages for a condominium. Residents must filter by their apartment ID to view their packages.
// @Tags         Packages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        condominiumId query     string  true  "Condominium UUID"
// @Param        apartmentId   query     string  false "Apartment UUID (Mandatory for Residents)"
// @Param        status        query     string  false "Filter by status (pending, withdrawn)"
// @Success      200           {object}  []usecases.PackageListItem
// @Failure      400           {object}  common.ErrResponse "Invalid parameters"
// @Failure      401           {object}  common.ErrResponse "User not authenticated"
// @Failure      403           {object}  common.ErrResponse "Permission denied"
// @Failure      500           {object}  common.ErrResponse "Internal server error"
// @Router       /packages [get]
func (h *ListPackagesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	query := r.URL.Query()

	condoStr := query.Get("condominiumId")
	if condoStr == "" {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "condominiumId is required",
		})
		return
	}

	condoID, err := uuid.Parse(condoStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid condominiumId format",
		})
		return
	}

	var aptID *uuid.UUID
	aptStr := query.Get("apartmentId")
	if aptStr != "" {
		parsed, err := uuid.Parse(aptStr)
		if err != nil {
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: "Invalid apartmentId format",
			})
			return
		}
		aptID = &parsed
	}

	var status *string
	statusStr := query.Get("status")
	if statusStr != "" {
		status = &statusStr
	}

	packages, err := h.ListPackages.Exec(r.Context(), usecases.ListPackagesReq{
		UserID:        userID,
		CondominiumID: condoID,
		ApartmentID:   aptID,
		Status:        status,
	})

	if err != nil {
		if errors.Is(err, usecases.ErrNoPermission) {
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "Permission denied. Residents must specify their apartmentId.",
			})
			return
		}

		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
			Message: "Failed to list packages",
		})
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, packages)
}
